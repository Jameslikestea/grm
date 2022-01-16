package cql

import (
	"time"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/gocql/gocql"
	"github.com/rs/zerolog/log"
	"github.com/scylladb/gocqlx/v2"
	"github.com/scylladb/gocqlx/v2/qb"
	"github.com/scylladb/gocqlx/v2/table"

	"github.com/Jameslikestea/grm/internal/config"
	"github.com/Jameslikestea/grm/internal/storage"
)

type CQLStorage struct {
	conn *gocqlx.Session
	obj  *table.Table
	ref  *table.Table
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
	}
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
	for _, obj := range objects {
		go func(o storage.Object, p string) {
			err := C.conn.Query(C.obj.Insert()).Bind(
				p,
				o.Type,
				o.Hash.String(),
				o.Content,
			).Exec()
			if err != nil {
				log.Error().Err(err).Msg("Failed to insert object")
			}
		}(obj, s)
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

func (C CQLStorage) GetObject(s string, hash plumbing.Hash) (storage.Object, error) {
	return storage.Object{}, nil
}
