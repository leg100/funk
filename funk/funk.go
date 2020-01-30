package funk

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"log"
	"path/filepath"

	"cloud.google.com/go/storage"
)

// Given a path to a directory, returns a list of terraform configuration
// files matching the fileglob "*.tf"
func ReadWorkspace(path string) (matches []string) {
	pattern := "*.tf"
	matches, err := filepath.Glob(filepath.Join(path, pattern))
	if err != nil {
		log.Fatal(err)
	}
	// strip directories from path
	var filenames []string
	for _, f := range matches {
		filenames = append(filenames, filepath.Base(f))
	}
	return filenames
}

func CreateTar(dir string, filenames []string) *bytes.Buffer {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	for _, f := range filenames {
		path := filepath.Join(dir, f)
		data, err := ioutil.ReadFile(path)
		if err != nil {
			log.Fatal(err)
		}
		hdr := &tar.Header{
			Name: f,
			Mode: 0600,
			Size: int64(len(data)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			log.Fatal(err)
		}
		if _, err := tw.Write(data); err != nil {
			log.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		log.Fatal(err)
	}
	return &buf
}

type StorageClient interface {
	Bucket(name string) *storage.BucketHandle
}

func Upload(client StorageClient, ctx context.Context, buf *bytes.Buffer, bucket, name string) (*storage.ObjectHandle, *storage.ObjectAttrs, error) {
	bh := client.Bucket(bucket)
	// Next check if the bucket exists
	if _, err := bh.Attrs(ctx); err != nil {
		return nil, nil, err
	}

	obj := bh.Object(name)
	w := obj.NewWriter(ctx)
	if _, err := w.Write(buf.Bytes()); err != nil {
		return nil, nil, err
	}
	if err := w.Close(); err != nil {
		return nil, nil, err
	}

	attrs, err := obj.Attrs(ctx)
	return obj, attrs, err
}

func Download(client StorageClient, ctx context.Context, bucket, name string) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	bh := client.Bucket(bucket)
	// Next check if the bucket exists
	if _, err := bh.Attrs(ctx); err != nil {
		return nil, err
	}

	obj := bh.Object(name)
	r, err := obj.NewReader(ctx)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	if _, err := io.ReadFull(r, buf.Bytes()); err != nil {
		return nil, err
	}

	return &buf, err
}
