package s3

//
// import (
// 	"bytes"
// 	"context"
// 	"fmt"
// 	"strings"
// 	"sync"
//
// 	"github.com/go-git/go-git/v5/plumbing"
// 	"github.com/minio/minio-go/v7"
// 	"github.com/minio/minio-go/v7/pkg/credentials"
// 	"github.com/rs/zerolog/log"
//
// 	"github.com/Jameslikestea/grm/internal/config"
// 	"github.com/Jameslikestea/grm/internal/storage"
// )
//
// type S3Storage struct {
// 	mc *minio.Client
// }
//
// func NewS3Storage() *S3Storage {
// 	c, err := minio.New(
// 		config.GetStorageS3Endpoint(), &minio.Options{
// 			Creds:  credentials.NewStaticV4(config.GetStorageS3AccessKey(), config.GetStorageS3SecretKey(), ""),
// 			Secure: config.GetStorageS3SSL(),
// 		},
// 	)
// 	if err != nil {
// 		log.Fatal().Err(err).Msg("Cannot Connect to S3")
// 	}
// 	return &S3Storage{
// 		mc: c,
// 	}
// }
//
// func (s2 S3Storage) StoreReferences(s string, references []storage.Reference) error {
// 	return nil
// }
//
// func (s2 S3Storage) StoreObjects(s string, objects []storage.Object) error {
// 	wg := sync.WaitGroup{}
// 	bucket := config.GetStorageS3Bucket()
// 	for _, obj := range objects {
// 		wg.Add(1)
// 		go func(o storage.Object) {
// 			info, err := s2.mc.PutObject(
// 				context.Background(),
// 				bucket,
// 				fmt.Sprintf("%s/objects/%s/%s", strings.Trim(s, `'"`), o.Hash.String()[0:2], o.Hash.String()),
// 				bytes.NewReader(o.Content),
// 				int64(len(o.Content)),
// 				minio.PutObjectOptions{},
// 			)
// 			log.Trace().Err(err).Interface("info", info).Msg("uploaded object")
// 			wg.Done()
// 		}(obj)
// 	}
// 	wg.Wait()
//
// 	return nil
// }
//
// func (s2 S3Storage) ListReferences(s string) ([]storage.Reference, error) {
// 	return []storage.Reference{}, nil
// }
//
// func (s2 S3Storage) ListObjects(s string) ([]storage.Object, error) {
// 	return []storage.Object{}, nil
// }
//
// func (s2 S3Storage) GetObject(s string, hash plumbing.Hash) (storage.Object, error) {
// 	return storage.Object{}, nil
// }
