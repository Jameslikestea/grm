package s3

import (
	"bytes"
	"crypto/rand"
	"encoding/gob"
	"errors"
	"fmt"
	"syscall"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/gocql/gocql"
	"github.com/rs/zerolog/log"

	"github.com/Jameslikestea/grm/internal/config"
	"github.com/Jameslikestea/grm/internal/models"
	"github.com/Jameslikestea/grm/internal/storage"
)

var _ storage.Storage = &S3Storage{}

type S3Storage struct {
	sess *session.Session
	sc   *s3.S3
}

func NewS3Storage() *S3Storage {
	checkLimit()

	conf := aws.Config{
		Credentials: credentials.NewStaticCredentials(
			config.GetStorageS3AccessKey(),
			config.GetStorageS3SecretKey(),
			"",
		),
		Endpoint:         aws.String(config.GetStorageS3Endpoint()),
		Region:           aws.String(config.GetStorageS3Region()),
		DisableSSL:       aws.Bool(!config.GetStorageS3SSL()),
		S3ForcePathStyle: aws.Bool(true),
	}

	sess, err := session.NewSession(&conf)
	if err != nil {
		log.Panic().Err(err).Msg("Cannot configure S3")
	}

	sc := s3.New(sess)

	return &S3Storage{
		sess: sess,
		sc:   sc,
	}
}

func checkLimit() {
	var nofiles syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &nofiles)
	if err != nil {
		log.Panic().Err(err).Msg("Cannot Check File Descriptors")
	}

	if nofiles.Cur < 65535 {
		log.Panic().Uint64("current", nofiles.Cur).Msg("S3 Implementation Requires File Limits To Be Above 65534")
	}
}

func (s2 S3Storage) getReferenceGob(s string) (map[plumbing.ReferenceName]plumbing.Hash, error) {
	bucket := config.GetStorageS3Bucket()
	key := fmt.Sprintf("%s/references.gob", s)
	m := map[plumbing.ReferenceName]plumbing.Hash{}

	input := &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}

	o, err := s2.sc.GetObject(input)
	if err != nil {
		return m, err
	}

	err = gob.NewDecoder(o.Body).Decode(&m)
	if err != nil {
		return m, err
	}

	return m, nil
}

func (s2 S3Storage) storeReferenceGob(s string, refs map[plumbing.ReferenceName]plumbing.Hash) error {
	bucket := config.GetStorageS3Bucket()
	key := fmt.Sprintf("%s/references.gob", s)

	b := bytes.NewBuffer([]byte{})

	err := gob.NewEncoder(b).Encode(refs)
	if err != nil {
		return err
	}

	input := &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &key,
		Body:   bytes.NewReader(b.Bytes()),
	}

	_, err = s2.sc.PutObject(input)

	return err
}

func (s2 S3Storage) getObjectsGob(s string) (map[plumbing.Hash]storage.Object, error) {
	bucket := config.GetStorageS3Bucket()
	key := fmt.Sprintf("%s/objects.gob", s)
	m := map[plumbing.Hash]storage.Object{}

	input := &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}

	o, err := s2.sc.GetObject(input)
	if err != nil {
		return m, err
	}

	err = gob.NewDecoder(o.Body).Decode(&m)
	if err != nil {
		return m, err
	}

	return m, nil
}

func (s2 S3Storage) storeObjectGob(s string, objs map[plumbing.Hash]storage.Object) error {
	bucket := config.GetStorageS3Bucket()
	key := fmt.Sprintf("%s/objects.gob", s)

	b := bytes.NewBuffer([]byte{})

	err := gob.NewEncoder(b).Encode(objs)
	if err != nil {
		return err
	}

	input := &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &key,
		Body:   bytes.NewReader(b.Bytes()),
	}

	_, err = s2.sc.PutObject(input)

	return err
}

func (s2 S3Storage) StoreObject(s string, object storage.Object, i int) error {
	objs, err := s2.getObjectsGob(s)
	log.Debug().Err(err).Interface("objects", objs).Msg("get object gob")
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			switch awsErr.Code() {
			case s3.ErrCodeNoSuchKey:
				// Do nothing
			default:
				return err
			}
		} else {
			return err
		}
	}

	objs[object.Hash] = object
	log.Debug().Interface("objects", objs).Msg("updated hashmap")

	err = s2.storeObjectGob(s, objs)
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

	log.Debug().Msg("Generated Hash Key")
	err = s2.StoreObject("_internal._hashkey", o, 0)
	log.Debug().Err(err).Msg("Storing Hash Key")
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
	refs, err := s2.getReferenceGob(s)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			switch awsErr.Code() {
			case s3.ErrCodeNoSuchKey:
				// Do nothing
			default:
				return err
			}
		} else {
			return err
		}
	}

	for _, ref := range references {
		refs[ref.Name] = ref.Hash
	}

	err = s2.storeReferenceGob(s, refs)

	return err
}

func (s2 S3Storage) StoreObjects(s string, objects []storage.Object) error {
	objs, err := s2.getObjectsGob(s)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			switch awsErr.Code() {
			case s3.ErrCodeNoSuchKey:
				// Do nothing
			default:
				return err
			}
		} else {
			return err
		}
	}

	for _, obj := range objects {
		objs[obj.Hash] = obj
	}

	err = s2.storeObjectGob(s, objs)

	return err
}

func (s2 S3Storage) ListReferences(s string) ([]storage.Reference, error) {
	refs, err := s2.getReferenceGob(s)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			switch awsErr.Code() {
			case s3.ErrCodeNoSuchKey:
				// Do nothing
			default:
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	var references []storage.Reference

	for ref, hash := range refs {
		references = append(references, storage.Reference{Name: ref, Hash: hash})
	}

	return references, nil
}

func (s2 S3Storage) ListObjects(s string) ([]storage.Object, error) {
	objs, err := s2.getObjectsGob(s)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			switch awsErr.Code() {
			case s3.ErrCodeNoSuchKey:
				// Do nothing
			default:
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	var os []storage.Object

	for _, obj := range objs {
		os = append(os, obj)
	}

	return os, nil
}

func (s2 S3Storage) GetObject(s string, hash plumbing.Hash) (storage.Object, error) {
	objs, err := s2.getObjectsGob(s)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			switch awsErr.Code() {
			case s3.ErrCodeNoSuchKey:
				// Do nothing
			default:
				return storage.Object{}, err
			}
		} else {
			return storage.Object{}, err
		}
	}

	o, ok := objs[hash]
	if !ok {
		return storage.Object{}, errors.New("cannot find object")
	}

	return o, nil
}
