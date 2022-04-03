package s3

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go/aws"
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

type objectList struct {
	list  []storage.Object
	hash  []plumbing.Hash
	count int64
	mu    sync.Mutex
}

func (o *objectList) Append(object storage.Object) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.list = append(o.list, object)
}

func (o *objectList) AppendHash(hash plumbing.Hash) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.hash = append(o.hash, hash)
}

func (o *objectList) List() []storage.Object {
	return o.list
}

func (o *objectList) ListHash() []plumbing.Hash {
	return o.hash
}

func (o *objectList) Add() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.count++
}

func (o *objectList) Sub() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.count--
}

func (o *objectList) Wait() {
	t := time.Now().Add(time.Second * 10)
	for {
		if o.count <= 0 {
			break
		}

		if time.Now().After(t) {
			log.Info().Int64("count", o.count).Msg("Waiting for Objects")
			t = time.Now().Add(time.Second * 10)
		}
	}
}

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
		log.Panic().Msg("S3 Implementation Requires File Limits To Be Above 65534")
	}
}

func (s2 S3Storage) StoreObject(s string, object storage.Object, i int) error {
	bucket := config.GetStorageS3Bucket()
	hash := object.Hash.String()
	key := fmt.Sprintf("%s/objects/%s/%s", s, hash[0:2], hash)
	content := bytes.NewReader(object.Content)

	input := &s3.PutObjectInput{
		Bucket:  &bucket,
		Key:     &key,
		Body:    content,
		Tagging: aws.String(fmt.Sprintf("type=%d", object.Type)),
	}

	_, err := s2.sc.PutObject(input)

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

func (s2 S3Storage) storeReference(s string, reference storage.Reference) error {
	bucket := config.GetStorageS3Bucket()
	key := fmt.Sprintf("%s/references/%X", s, reference.Name.String())
	content := strings.NewReader(fmt.Sprintf("%s", reference.Hash.String()))

	input := &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &key,
		Body:   content,
	}

	_, err := s2.sc.PutObject(input)

	return err

	return nil
}

func (s2 S3Storage) StoreReferences(s string, references []storage.Reference) error {
	for _, ref := range references {
		s2.storeReference(s, ref)
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

func (s2 S3Storage) getReference(s, refname string) (storage.Reference, error) {
	bucket := config.GetStorageS3Bucket()
	key := fmt.Sprintf("%s/references/%X", s, refname)

	getObjectInput := &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}

	obj, err := s2.sc.GetObject(getObjectInput)

	if err != nil {
		log.Warn().Str("Key", key).Err(err).Msg("Cannot Get Object")
		return storage.Reference{}, err
	}

	content, err := ioutil.ReadAll(obj.Body)
	if err != nil {
		return storage.Reference{}, err
	}

	hash := plumbing.NewHash(string(content))

	return storage.Reference{
		Name: plumbing.ReferenceName(refname),
		Hash: hash,
	}, nil
}

func (s2 S3Storage) ListReferences(s string) ([]storage.Reference, error) {
	bucket := config.GetStorageS3Bucket()
	prefix := fmt.Sprintf("%s/references/", s)

	objectListInput := &s3.ListObjectsV2Input{
		Bucket: &bucket,
		Prefix: &prefix,
	}

	var st []storage.Reference

	s2.sc.ListObjectsV2Pages(
		objectListInput, func(output *s3.ListObjectsV2Output, b bool) bool {
			log.Trace().Msg("Got Page")
			for _, ref := range output.Contents {
				str := strings.TrimPrefix(*ref.Key, prefix)
				name, err := hex.DecodeString(str)
				if err != nil {
					continue
				}

				reference, err := s2.getReference(s, string(name))
				if err != nil {
					log.Warn().Err(err).Msg("Cannot Get Reference")
				}

				st = append(st, reference)
			}
			return b
		},
	)

	return st, nil
}

func (s2 S3Storage) ListObjects(s string) ([]storage.Object, error) {
	bucket := config.GetStorageS3Bucket()
	concurrency := config.GetStorageS3Concurrency()
	prefix := fmt.Sprintf("%s/objects/", s)

	objectListInput := &s3.ListObjectsV2Input{
		Bucket: &bucket,
		Prefix: &prefix,
	}

	os := new(objectList)

	s2.sc.ListObjectsV2Pages(
		objectListInput, func(output *s3.ListObjectsV2Output, b bool) bool {
			log.Trace().Msg("Got Page")
			for _, object := range output.Contents {
				log.Debug().Str("key", *object.Key).Msg("Listed Object")
				hashString := strings.TrimPrefix(*object.Key, prefix)[3:]

				hash := plumbing.NewHash(hashString)
				os.Add()
				os.AppendHash(hash)

			}
			log.Debug().Bool("continue", !b).Msg("Last Page")
			return !b
		},
	)

	c := make(chan plumbing.Hash, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(s string, h chan plumbing.Hash) {
			for {
				hash, ok := <-h
				if !ok {
					break
				}
				obj, err := s2.GetObject(s, hash)
				if err != nil {
					log.Warn().Err(err).Msg("Cannot Get Object")
					continue
				}

				os.Append(obj)
				os.Sub()
			}
		}(s, c)
	}

	for _, h := range os.hash {
		c <- h
	}

	close(c)
	os.Wait()

	log.Info().Int("objects", len(os.List())).Msg("Receieved Objects")

	return os.List(), nil
}

func (s2 S3Storage) GetObject(s string, hash plumbing.Hash) (storage.Object, error) {
	bucket := config.GetStorageS3Bucket()
	key := fmt.Sprintf("%s/objects/%s/%s", s, hash.String()[0:2], hash.String())

	getInput := &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}
	getTagInput := &s3.GetObjectTaggingInput{
		Bucket: &bucket,
		Key:    &key,
	}

	tagOut, err := s2.sc.GetObjectTagging(getTagInput)
	if err != nil {
		return storage.Object{}, err
	}

	objOut, err := s2.sc.GetObject(getInput)
	if err != nil {
		return storage.Object{}, err
	}

	t := -1

	for _, tag := range tagOut.TagSet {
		if *tag.Key == "type" {
			t, err = strconv.Atoi(*tag.Value)
			if err != nil {
				log.Warn().Err(err).Msg("Cannot convert tag to type")
			}
		}
	}

	content, _ := ioutil.ReadAll(objOut.Body)
	pt := plumbing.ObjectType(t)

	return storage.Object{
		Hash:    hash,
		Content: content,
		Type:    pt,
	}, nil
}
