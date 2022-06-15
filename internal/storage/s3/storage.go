package s3

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/Jameslikestea/d-badger/lock"
	"io/ioutil"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/gocql/gocql"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rs/zerolog/log"

	"github.com/Jameslikestea/grm/internal/config"
	"github.com/Jameslikestea/grm/internal/models"
	"github.com/Jameslikestea/grm/internal/storage"
)

var _ storage.Storage = &S3Storage{}

type S3Storage struct {
	mc *minio.Client
}

func NewS3Storage() *S3Storage {
	checkLimit()

	c, err := minio.New(
		config.GetStorageS3Endpoint(), &minio.Options{
			Creds:  credentials.NewStaticV4(config.GetStorageS3AccessKey(), config.GetStorageS3SecretKey(), ""),
			Secure: config.GetStorageS3SSL(),
		},
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot Connect to S3")
	}
	return &S3Storage{
		mc: c,
	}
}

func checkLimit() {
	var nofiles syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &nofiles)
	if err != nil {
		log.Panic().Err(err).Msg("Cannot Check File Descriptors")
	}

	if nofiles.Cur < 65535 {
		log.Panic().Msg("S3 Implementation Requires File Limits To Be Above 65534")
	}
}

func (s2 S3Storage) StoreObject(s string, object storage.Object, i int) error {
	bucket := config.GetStorageS3Bucket()

	options := minio.PutObjectOptions{
		UserTags: map[string]string{
			"type": fmt.Sprintf("%d", object.Type),
		},
	}

	if i > 0 {
		options.UserTags["grm-expire"] = time.Now().Add(time.Duration(i) * time.Second).UTC().Format(time.RFC3339)
	}

	info, err := s2.mc.PutObject(
		context.Background(),
		bucket,
		fmt.Sprintf("%s/objects/%s/%s", strings.Trim(s, `'"`), object.Hash.String()[0:2], object.Hash.String()),
		bytes.NewReader(object.Content),
		int64(len(object.Content)),
		options,
	)
	log.Trace().Err(err).Interface("info", info).Msg("uploaded object")
	return err
}

func (s2 S3Storage) GenerateHashKey() error {
	h := plumbing.ComputeHash(0, []byte("global_hash_key"))
	_, err := s2.GetHashKey()
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

func (s2 S3Storage) GetHashKey() ([]storage.HashKey, error) {
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

func (s2 S3Storage) StoreReferences(s string, references []storage.Reference) error {
	bucket := config.GetStorageS3Bucket()

	for _, ref := range references {
		r := fmt.Sprintf("%s/references/%X", strings.Trim(s, `'"`), ref.Name.String())
		info, err := s2.mc.PutObject(
			context.Background(),
			bucket,
			r,
			strings.NewReader(ref.Hash.String()),
			int64(len(ref.Hash.String())),
			minio.PutObjectOptions{},
		)
		log.Debug().Interface("info", info).Msg("Uploaded Reference")
		if err != nil {
			log.Warn().Err(err).Msg("Cannot upload reference")
		}
	}
	return nil
}

func (s2 S3Storage) StoreObjects(s string, objects []storage.Object) error {
	concurrency := config.GetStorageS3Concurrency()
	c := make(chan storage.Object, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(in <-chan storage.Object) {
			for {
				o, ok := <-in
				if !ok {
					break
				}
				s2.StoreObject(s, o, 0)
			}
		}(c)
	}

	for _, obj := range objects {
		c <- obj
	}
	close(c)

	return nil
}

func (s2 S3Storage) ListReferences(s string) ([]storage.Reference, error) {
	bucket := config.GetStorageS3Bucket()
	prefix := fmt.Sprintf("%s/references/", s)
	objs := s2.mc.ListObjects(
		context.Background(), bucket, minio.ListObjectsOptions{
			Prefix: prefix,
		},
	)
	log.Info().Msg("Listing References")

	var refs []storage.Reference

	for obj := range objs {
		if obj.Err != nil {
			log.Warn().Err(obj.Err).Msg("Error getting objects")
			continue
		}
		o, err := s2.mc.GetObject(context.Background(), bucket, obj.Key, minio.GetObjectOptions{})
		if err != nil {
			log.Warn().Err(err).Str("key", obj.Key).Msg("Cannot get object")
			continue
		}

		b, _ := ioutil.ReadAll(o)
		h := plumbing.NewHash(string(b))

		refname, err := hex.DecodeString(strings.TrimPrefix(obj.Key, prefix))
		if err != nil {
			continue
		}
		refs = append(
			refs, storage.Reference{
				Hash: h,
				Name: plumbing.ReferenceName(refname),
			},
		)
	}

	log.Info().Int("refs", len(refs)).Msg("Found References")
	if len(refs) > 0 {
		log.Info().Str("ref", refs[0].Name.Short()).Bool("tag", refs[0].Name.IsTag()).Msg("First Ref")
	}
	return refs, nil
}

func (s2 S3Storage) ListObjects(s string) ([]storage.Object, error) {
	bucket := config.GetStorageS3Bucket()
	prefix := fmt.Sprintf("%s/objects/", s)
	// concurrent := config.GetStorageS3Concurrency()

	objs := s2.mc.ListObjects(
		context.Background(), bucket, minio.ListObjectsOptions{
			Prefix:    prefix,
			Recursive: true,
		},
	)

	var objects []storage.Object
	// c := make(chan plumbing.Hash, concurrent)

	for obj := range objs {
		if obj.Err != nil {
			log.Warn().Err(obj.Err).Msg("Error getting objects")
			continue
		}

		key := strings.TrimPrefix(obj.Key, prefix)[3:]
		log.Info().Str("key", key).Msg("Getting Object")
		ob, err := s2.GetObject(s, plumbing.NewHash(key))
		if err != nil {
			log.Warn().Err(err).Msg("Failed to get object")
			continue
		}
		objects = append(objects, ob)
	}

	return objects, nil
}

func (s2 S3Storage) GetObject(s string, hash plumbing.Hash) (storage.Object, error) {
	bucket := config.GetStorageS3Bucket()
	key := fmt.Sprintf("%s/objects/%s/%s", s, hash.String()[0:2], hash.String())

	obj, err := s2.mc.GetObject(context.Background(), bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return storage.Object{}, err
	}

	tags, err := s2.mc.GetObjectTagging(context.Background(), bucket, key, minio.GetObjectTaggingOptions{})
	if err != nil {
		return storage.Object{}, err
	}

	if tags == nil {
		return storage.Object{}, errors.New("cannot determine object type: nil metadata")
	}

	t, ok := tags.ToMap()["type"]
	if !ok {
		return storage.Object{}, errors.New("cannot determine object type: nil type")
	}

	ot, err := strconv.Atoi(t)
	if err != nil {
		return storage.Object{}, errors.New("cannot determine object type: cannot convert")
	}

	pt := plumbing.ObjectType(ot)
	content, _ := ioutil.ReadAll(obj)

	return storage.Object{
		Type:    pt,
		Content: content,
		Hash:    hash,
	}, nil
}

func (s2 S3Storage) Lock(string) (lock.Lock, error) {
	return nil, nil
}

func (s2 S3Storage) Unlock(lock.Lock) error {
	return nil
}
