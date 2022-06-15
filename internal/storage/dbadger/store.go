package dbadger

import (
	"bytes"
	"crypto/rand"
	"encoding/gob"
	"fmt"
	dbadger "github.com/Jameslikestea/d-badger"
	"github.com/Jameslikestea/d-badger/lock"
	"github.com/Jameslikestea/grm/internal/models"
	"github.com/Jameslikestea/grm/internal/storage"
	"github.com/dgraph-io/badger/v3"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/gocql/gocql"
	"github.com/rs/zerolog/log"
	"strings"
)

var _ storage.Storage = &Storage{}

type Storage struct {
	db *dbadger.Config
}

func (s2 Storage) StoreReferences(s string, references []storage.Reference) error {
	repo := strings.Trim(s, `'"`)
	d, err := s2.db.GetDB(repo)
	if err != nil {
		return err
	}
	txn := d.Badger().NewTransaction(true)
	for _, ref := range references {
		key := fmt.Sprintf("reference/%s", ref.Name.String())
		value := ref.Hash.String()

		err = txn.Set([]byte(key), []byte(value))
		if err != nil {
			log.Error().Err(err).Msg("Failed to commit reference")
		}
	}
	err = txn.Commit()
	if err != nil {
		return err
	}
	return d.Close()
}

func (s2 Storage) StoreObjects(s string, objects []storage.Object) error {
	repo := strings.Trim(s, `'"`)
	d, err := s2.db.GetDB(repo)
	if err != nil {
		return err
	}
	txn := d.Badger().NewTransaction(true)
	for _, obj := range objects {
		key := fmt.Sprintf("object/%s", obj.Hash.String())
		var value bytes.Buffer
		err = gob.NewEncoder(&value).Encode(obj)
		if err != nil {
			log.Error().Err(err).Msg("Failed to encode object")
			continue
		}

		err = txn.Set([]byte(key), value.Bytes())
		if err != nil {
			log.Error().Err(err).Msg("Failed to commit reference")
		}
	}
	err = txn.Commit()
	if err != nil {
		return err
	}
	return d.Close()
}

func (s2 Storage) StoreObject(s string, object storage.Object, i int) error {
	repo := strings.Trim(s, `'"`)
	d, err := s2.db.GetDB(repo)
	if err != nil {
		return err
	}
	txn := d.Badger().NewTransaction(true)
	key := fmt.Sprintf("object/%s", object.Hash.String())
	var value bytes.Buffer
	err = gob.NewEncoder(&value).Encode(object)
	if err != nil {
		log.Error().Err(err).Msg("Failed to encode object")
	}

	err = txn.Set([]byte(key), value.Bytes())
	if err != nil {
		log.Error().Err(err).Msg("Failed to commit reference")
	}
	err = txn.Commit()
	if err != nil {
		return err
	}
	return d.Close()
}

func (s2 Storage) ListReferences(s string) ([]storage.Reference, error) {
	repo := strings.Trim(s, `'"`)
	d, err := s2.db.GetDB(repo)
	defer d.Close()
	if err != nil {
		return nil, err
	}

	var refs []storage.Reference

	d.Badger().View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Seek([]byte("reference/")); it.ValidForPrefix([]byte("reference/")); it.Next() {
			item := it.Item()
			k := item.Key()
			err := item.Value(func(val []byte) error {
				refs = append(refs, storage.Reference{
					Hash: plumbing.NewHash(string(val)),
					Name: plumbing.ReferenceName(strings.TrimPrefix(string(k), "reference/")),
				})
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	return refs, nil
}

func (s2 Storage) ListObjects(s string) ([]storage.Object, error) {
	repo := strings.Trim(s, `'"`)
	d, err := s2.db.GetDB(repo)
	defer d.Close()
	if err != nil {
		return nil, err
	}

	var objs []storage.Object

	d.Badger().View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Seek([]byte("object/")); it.ValidForPrefix([]byte("object/")); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var obj storage.Object
				err := gob.NewDecoder(bytes.NewReader(val)).Decode(&obj)
				if err != nil {
					return err
				}
				objs = append(objs, obj)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	return objs, nil
}

func (s2 Storage) GetObject(s string, hash plumbing.Hash) (storage.Object, error) {
	repo := strings.Trim(s, `'"`)
	d, err := s2.db.GetDB(repo)
	defer d.Close()
	if err != nil {
		return storage.Object{}, err
	}

	var obj storage.Object

	err = d.Badger().View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(fmt.Sprintf("object/%s", hash.String())))
		if err != nil {
			return err
		}

		err = item.Value(func(val []byte) error {
			err := gob.NewDecoder(bytes.NewReader(val)).Decode(&obj)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})

	return obj, err
}

func (s2 Storage) GenerateHashKey() error {
	l, err := s2.Lock("_internal._hashkey")
	if err != nil {
		return err
	}
	defer s2.Unlock(l)

	h := plumbing.ComputeHash(0, []byte("global_hash_key"))
	_, err = s2.GetHashKey()
	if err == nil {
		return nil
	}

	r := rand.Reader
	buf := make([]byte, 256)
	r.Read(buf)

	hk := storage.HashKey{KID: gocql.TimeUUID(), K: fmt.Sprintf("%X", buf)}

	o := storage.Object{
		Hash:    h,
		Type:    0,
		Content: models.Marshal(hk),
	}

	err = s2.StoreObject("_internal._hashkey", o, 0)

	return err
}

func (s2 Storage) GetHashKey() ([]storage.HashKey, error) {
	h := plumbing.ComputeHash(0, []byte("global_hash_key"))
	o, err := s2.GetObject("_internal._hashkey", h)
	if err != nil {
		log.Warn().Err(err).Msg("Cannot Find Hash Key")
		return nil, err
	}
	var hk []storage.HashKey
	models.Unmarshal(o.Content, &hk)
	return hk, nil
}

func (s2 Storage) Lock(db string) (lock.Lock, error) {
	return s2.db.Lock.Acquire(db)
}

func (s2 Storage) Unlock(l lock.Lock) error {
	return s2.db.Lock.Release(l)
}

func New(opts ...dbadger.Opt) *Storage {
	d := dbadger.New()
	d.WithOpts(opts...)

	return &Storage{
		db: d,
	}
}
