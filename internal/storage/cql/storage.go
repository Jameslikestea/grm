package cql

import (
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/gocql/gocql"
	"github.com/rs/zerolog/log"
	"github.com/scylladb/gocqlx/v2"
	"github.com/scylladb/gocqlx/v2/qb"
	"github.com/scylladb/gocqlx/v2/table"
	"golang.org/x/sync/errgroup"

	"github.com/Jameslikestea/grm/internal/config"
	"github.com/Jameslikestea/grm/internal/storage"
)

var _ storage.Storage = &CQLStorage{}

type CQLStorage struct {
	conn *gocqlx.Session
	obj  *table.Table
	ref  *table.Table
	key  *table.Table
}

func NewCQLStorage() *CQLStorage {
	cluster := gocql.NewCluster(config.GetStorageCQLEndpoint())

	cluster.Keyspace = config.GetStorageCQLKeyspace()
	cluster.Timeout = 5 * time.Second

	sess, err := gocqlx.WrapSession(cluster.CreateSession())
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot Connect to CQL")
	}
	err = sess.ExecStmt(
		"CREATE TABLE IF NOT EXISTS objs (" +
			"package TEXT," +
			"type TINYINT," +
			"hash TEXT," +
			"content BLOB," +
			"PRIMARY KEY (package, hash));",
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot create table")
	}

	err = sess.ExecStmt(
		"CREATE TABLE IF NOT EXISTS refs (" +
			"package TEXT," +
			"ref TEXT," +
			"hash TEXT," +
			"PRIMARY KEY (package, ref));",
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot create table")
	}

	return &CQLStorage{
		conn: &sess,
		obj: table.New(
			table.Metadata{
				Name:    "objs",
				Columns: []string{"package", "type", "hash", "content"},
				PartKey: []string{"package", "hash"},
			},
		),
		ref: table.New(
			table.Metadata{
				Name:    "refs",
				Columns: []string{"package", "ref", "hash"},
				PartKey: []string{"package", "ref"},
			},
		),
		key: table.New(
			table.Metadata{
				Name:    "hash_key",
				Columns: []string{"kid", "k"},
				PartKey: []string{"kid"},
			},
		),
	}
}

func (C CQLStorage) GenerateHashKey() error {
	_, err := C.GetHashKey()
	if err == nil {
		return nil
	}

	r := rand.Reader
	buf := make([]byte, 256)
	r.Read(buf)

	log.Info().Bytes("key", buf).Msg("Generated Hash Key")
	err = C.conn.ExecStmt("CREATE TABLE IF NOT EXISTS hash_key(kid uuid primary key, k varchar);")
	if err != nil {
		log.Error().Err(err).Msg("Cannot create key table")
	}

	C.conn.Query(C.key.InsertBuilder().Unique().ToCql()).Bind(gocql.TimeUUID(), fmt.Sprintf("%X", buf)).Exec()

	return nil
}

func (C CQLStorage) GetHashKey() ([]storage.HashKey, error) {
	var keys []storage.HashKey
	err := C.conn.Query(C.key.SelectAll()).Select(&keys)
	if err != nil {
		return nil, err
	}
	if len(keys) == 0 {
		return nil, errors.New("no keys")
	}
	return keys, nil
}

func (C CQLStorage) StoreReferences(s string, references []storage.Reference) error {
	for _, ref := range references {
		go func(r storage.Reference, p string) {
			err := C.conn.Query(C.ref.Insert()).Bind(
				p,
				r.Name.String(),
				r.Hash.String(),
			).Exec()
			if err != nil {
				log.Error().Err(err).Msg("Failed to insert reference")
			}
		}(ref, s)
	}
	return nil
}

func (C CQLStorage) StoreObjects(s string, objects []storage.Object) error {
	var e errgroup.Group
	for _, obj := range objects {
		e.Go(func() error { return C.StoreObject(s, obj, 0) })
	}
	return e.Wait()
}

func (C CQLStorage) StoreObject(s string, object storage.Object, ttl int) error {
	bldr := C.obj.InsertBuilder()
	if ttl > 0 {
		bldr.TTL(time.Second * time.Duration(ttl))
	}
	err := C.conn.Query(bldr.ToCql()).Bind(
		s,
		object.Type,
		object.Hash.String(),
		object.Content,
	).Exec()
	if err != nil {
		log.Error().Err(err).Msg("Failed to insert object")
		return err
	}
	return nil
}

func (C CQLStorage) ListReferences(s string) ([]storage.Reference, error) {
	type ref struct {
		Package string
		Ref     string
		Hash    string
	}
	var refs []ref
	stmt, names := qb.Select(C.ref.Name()).Columns(
		"ref",
		"hash",
	).Where(qb.Eq("package")).ToCql() // C.ref.SelectBuilder("ref", "hash").Where(qb.Eq("package")).ToCql()
	err := C.conn.Query(stmt, names).Bind(s).Select(&refs)
	if err != nil {
		log.Error().Err(err).Str("statement", stmt).Msg("Cannot Select package refs")
	}

	log.Info().Str("repo", s).Msg("Found Refs")
	var rs []storage.Reference
	for _, ref := range refs {
		rs = append(rs, storage.Reference{Name: plumbing.ReferenceName(ref.Ref), Hash: plumbing.NewHash(ref.Hash)})
	}
	return rs, nil
}

func (C CQLStorage) ListObjects(s string) ([]storage.Object, error) {
	type obj struct {
		Package string
		Hash    string
		Content []byte
		Type    uint8
	}
	var refs []obj
	stmt, names := qb.Select(C.obj.Name()).Columns(
		"hash",
		"content",
		"type",
	).Where(qb.Eq("package")).ToCql() // C.ref.SelectBuilder("ref", "hash").Where(qb.Eq("package")).ToCql()
	err := C.conn.Query(stmt, names).Bind(s).Select(&refs)
	if err != nil {
		log.Error().Err(err).Str("statement", stmt).Msg("Cannot Select package refs")
	}

	var os []storage.Object
	for _, o := range refs {
		os = append(
			os, storage.Object{
				Hash:    plumbing.NewHash(o.Hash),
				Type:    plumbing.ObjectType(o.Type),
				Content: o.Content,
			},
		)
	}

	return os, nil
}

// GetObject will get the object containing the specified hash. The hash must be a full hash
func (C CQLStorage) GetObject(s string, hash plumbing.Hash) (storage.Object, error) {
	type obj struct {
		Package string
		Hash    string
		Content []byte
		Type    uint8
	}

	var refs obj
	stmt, names := qb.Select(C.obj.Name()).Columns(
		"package",
		"hash",
		"content",
		"type",
	).Where(
		qb.Eq("hash"),
		qb.Eq("package"),
	).ToCql()

	err := C.conn.Query(stmt, names).Bind(hash.String(), s).Get(&refs)
	if err != nil {
		log.Error().Err(err).Str("statement", stmt).Msg("Cannot Select package refs")
	}

	if refs.Hash == "" {
		return storage.Object{}, errors.New("no such object")
	}

	var os storage.Object
	os = storage.Object{
		Hash:    plumbing.NewHash(refs.Hash),
		Type:    plumbing.ObjectType(refs.Type),
		Content: refs.Content,
	}

	return os, nil
}
