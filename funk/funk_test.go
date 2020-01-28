package funk

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"testing"

	"cloud.google.com/go/storage"
	"github.com/fsouza/fake-gcs-server/fakestorage"
)

func TestReadWorkspace(t *testing.T) {
	want := []string{"one.tf", "two.tf", "three.tf"}
	got := ReadWorkspace("./testdata")
	if len(got) != len(want) {
		t.Errorf("Got slice of length %d, wanted %d", len(got), len(want))
	}

	// sort slices so we can compare
	sort.Strings(want)
	sort.Strings(got)

	for i := range got {
		if got[i] != want[i] {
			t.Errorf("Got %q, but wanted %q", got[i], want[i])
		}
	}
}

func TestCreateTar(t *testing.T) {
	var want = 3
	var got = 0

	filenames := []string{"one.tf", "two.tf", "three.tf"}
	buf := CreateTar("./testdata", filenames)

	// Open and iterate through the files in the archive.
	tr := tar.NewReader(buf)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			t.Errorf("Could not find header in tarfile: %q", err)
		}
		fmt.Printf("Contents of %s:\n", hdr.Name)
		if _, err := io.Copy(os.Stdout, tr); err != nil {
			log.Fatal(err)
		}
		fmt.Println()
		got++
	}

	if got != want {
		t.Errorf("Got %d files in tar file, but expected %d", got, want)
	}
}

func setupFakeStorage() *fakestorage.Server {
	opts := fakestorage.Options{
		StorageRoot: "testdata",
	}
	server, err := fakestorage.NewServerWithOptions(opts)

	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("server started at %s", server.URL())

	return server
}

func TestUpload(t *testing.T) {
	server := setupFakeStorage()
	client := server.Client()
	ctx := context.Background()
	var buf bytes.Buffer

	_, _, err := Upload(client, ctx, &buf, "piss", "pah")
	if err != storage.ErrBucketNotExist {
		t.Errorf("expected bucket not exist error, but got %q", err)
	}

	_, _, err = Upload(client, ctx, &buf, "bucket", "pah")
	if err != nil {
		t.Errorf("expected no error, but got %q", err)
	}

	server.Stop()
}
