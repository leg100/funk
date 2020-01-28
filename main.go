package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"github.com/leg100/funk/funk"
	"os"

	"cloud.google.com/go/storage"
)

func main() {
	// Prevent log from printing out time information
	log.SetFlags(0)

	var projectID, bucket, source, name string

	flag.StringVar(&bucket, "bucket", "", "the bucket to upload content to")
	flag.StringVar(&projectID, "project", "", "the ID of the GCP project to use")
	flag.StringVar(&source, "source", "funk/testdata", "the path to the source")
	flag.Parse()

	// If they haven't set the bucket or projectID nor specified
	// in the environment, then fail if missing.
	bucket = mustGetEnv("FUNK_BUCKET", bucket)
	projectID = mustGetEnv("CLOUDSDK_CORE_PROJECT", projectID)

	filenames := funk.ReadWorkspace(source)
	buf := funk.CreateTar(source, filenames)

	name = "test-sample.tar"

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	_, objAttrs, err := funk.Upload(client, ctx, buf, projectID, bucket, name)
	if err != nil {
		switch err {
		case storage.ErrBucketNotExist:
			log.Fatal("Please create the bucket first e.g. with `gsutil mb`")
		default:
			log.Fatal(err)
		}
	}

	log.Printf("URL: %s", objectURL(objAttrs))
	log.Printf("Size: %d", objAttrs.Size)
	log.Printf("MD5: %x", objAttrs.MD5)
	log.Printf("objAttrs: %+v", objAttrs)
}

func mustGetEnv(envKey, defaultValue string) string {
	val := os.Getenv(envKey)
	if val == "" {
		val = defaultValue
	}
	if val == "" {
		log.Fatalf("%q should be set", envKey)
	}
	return val
}

func objectURL(objAttrs *storage.ObjectAttrs) string {
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", objAttrs.Bucket, objAttrs.Name)
}
