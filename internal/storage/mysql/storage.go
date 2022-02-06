package mysql

import "C"
import (
	"context"

	"github.com/go-git/go-git/v5/plumbing"
	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog/log"

	"github.com/Jameslikestea/grm/internal/config"
	"github.com/Jameslikestea/grm/internal/storage"
	"github.com/Jameslikestea/grm/internal/storage/ent"
	"github.com/Jameslikestea/grm/internal/storage/ent/object"
	"github.com/Jameslikestea/grm/internal/storage/ent/reference"
)

type SQLLiteStorage struct {
	c *ent.Client
}

func NewSQLLiteStorage() *SQLLiteStorage {
	client, err := ent.Open("mysql", config.GetMySQLURL())
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot instantiate sqlite3")
	}

	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatal().Err(err).Msg("Cannot create schema")
	}

	return &SQLLiteStorage{
		c: client,
	}
}

func (S SQLLiteStorage) StoreReferences(s string, references []storage.Reference) error {
	ops := make([]*ent.ReferenceCreate, len(references))
	for i, ref := range references {
		ops[i] = S.c.Reference.Create().SetRef(ref.Name.String()).SetHash(ref.Hash.String()).SetPackage(s)
	}
	_, err := S.c.Reference.CreateBulk(ops...).Save(context.Background())
	return err
}

func (S SQLLiteStorage) StoreObjects(s string, objects []storage.Object) error {
	ops := make([]*ent.ObjectCreate, len(objects))
	for i, obj := range objects {
		ops[i] = S.c.Object.Create().SetHash(obj.Hash.String()).SetType(int8(obj.Type)).SetContent(obj.Content).SetPackage(s)
	}
	_, err := S.c.Object.CreateBulk(ops...).Save(context.Background())
	return err
}

func (S SQLLiteStorage) ListReferences(s string) ([]storage.Reference, error) {
	refs, err := S.c.Reference.Query().Where(reference.Package(s)).Select(
		reference.FieldPackage,
		reference.FieldRef,
		reference.FieldHash,
	).All(context.Background())

	if err != nil {
		log.Error().Err(err).Msg("Cannot Select package refs")
	}

	log.Info().Str("repo", s).Msg("Found Refs")
	var rs []storage.Reference
	for _, ref := range refs {
		rs = append(rs, storage.Reference{Name: plumbing.ReferenceName(ref.Ref), Hash: plumbing.NewHash(ref.Hash)})
	}
	return rs, nil
}

func (S SQLLiteStorage) ListObjects(s string) ([]storage.Object, error) {
	refs, err := S.c.Object.Query().Where(object.Package(s)).Select(
		object.FieldHash,
		object.FieldType,
		object.FieldContent,
	).All(context.Background())

	if err != nil {
		log.Error().Err(err).Msg("Cannot Select package refs")
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

func (S SQLLiteStorage) GetObject(s string, hash plumbing.Hash) (storage.Object, error) {
	return storage.Object{}, nil
}
