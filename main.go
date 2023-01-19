package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	var (
		prefix    string
		bucket    string
		accessKey string
		secretKey string
		endpoint  string
		dryRun    bool
		insecure  bool
	)
	flag.StringVar(&prefix, "prefix", "", "prefix matching the objects to be recovered")
	flag.StringVar(&bucket, "bucket", "", "bucket")
	flag.StringVar(&accessKey, "access-key", "", "MinIO access key")
	flag.StringVar(&accessKey, "ak", "", "MinIO access key")
	flag.StringVar(&secretKey, "secret-key", "", "MinIO secret key")
	flag.StringVar(&secretKey, "sk", "", "MinIO secret key")
	flag.StringVar(&endpoint, "endpoint", "", "MinIO endpoint url, e.g https://minio-lb:9000")
	flag.BoolVar(&dryRun, "dry-run", false, "doesn't recover deleted objects, simply lists the delete marker versions")
	flag.BoolVar(&dryRun, "n", false, "doesn't recover deleted objects, simply lists the delete marker versions")
	flag.BoolVar(&insecure, "insecure", false, "connect via HTTP")
	flag.BoolVar(&insecure, "k", false, "connect via HTTP")
	flag.Parse()

	if bucket == "" {
		log.Fatal("Bucket can't be empty")
	}
	if accessKey == "" || secretKey == "" {
		log.Fatal("Access key or secret key can't be empty")
	}

	s3Client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: !insecure,
	})
	if err != nil {
		log.Fatal("failed to connect to MinIO")
	}

	if dryRun {
		fmt.Println("Objects with their current version as delete markers:")
	}
	undoRemoveCh := make(chan minio.ObjectInfo)
	go func() {
		defer close(undoRemoveCh)
		opts := minio.ListObjectsOptions{
			WithVersions: true,
			Prefix:       prefix,
			Recursive:    true,
		}

		// List all objects from a bucket-name with a matching prefix.
		for obj := range s3Client.ListObjects(context.Background(), bucket, opts) {
			if obj.Err != nil {
				fmt.Println(obj.Err)
				return
			}
			if obj.IsDeleteMarker && obj.IsLatest {
				fmt.Printf("%s: %s\n", obj.Key, obj.VersionID)
				if dryRun {
					continue // skip adding objects to undoRemoveCh
				}
				undoRemoveCh <- obj
			}
		}
	}()

	for err := range s3Client.RemoveObjects(context.Background(), bucket, undoRemoveCh, minio.RemoveObjectsOptions{}) {
		fmt.Println(err)
	}
}
