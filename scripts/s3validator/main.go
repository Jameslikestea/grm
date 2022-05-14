package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-git/go-git/v5/plumbing"

	"github.com/Jameslikestea/grm/internal/config"
	"github.com/Jameslikestea/grm/internal/storage"
)

var (
	hash string
	pkg  string
)

func main() {
	flag.StringVar(&hash, "hash", "", "The hash to search for")
	flag.StringVar(&pkg, "pkg", "", "The pkg to search in")
	flag.Parse()

	config.GetConfig()

	epoint := config.GetStorageS3Endpoint()
	region := config.GetStorageS3Region()
	bucket := config.GetStorageS3Bucket()

	sess, _ := session.NewSession(
		&aws.Config{
			Endpoint: &epoint,
			Credentials: credentials.NewStaticCredentials(
				config.GetStorageS3AccessKey(),
				config.GetStorageS3SecretKey(),
				"",
			),
			Region: &region,
		},
	)

	svc := s3.New(sess)

	objects := fmt.Sprintf("%s/objects.gob", pkg)

	out, err := svc.GetObject(
		&s3.GetObjectInput{
			Bucket: &bucket,
			Key:    &objects,
		},
	)

	if err != nil {
		log.Fatalln(err)
	}

	var i map[plumbing.Hash]storage.Object

	gob.NewDecoder(out.Body).Decode(&i)

	parsedHash := plumbing.NewHash(hash)
	obj, ok := i[parsedHash]

	log.Printf("Found %d objects\n", len(i))
	log.Printf("Has Hash: %s\n", strconv.FormatBool(ok))
	log.Printf(string(obj.Content))
}

// 59b19175035095196a623aac8d0f5c83481f3d77
