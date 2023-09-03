package uz2

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"testing"
)

func TestWriter(t *testing.T) {
	f, err := os.Open("testdata/the-adventures-of-sherlock-homes-by-arthur-conan-doyle.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	// Unfortunately, the zlib implementation in Go does not produce
	// byte-equal compression with the Unreal Engine. However,
	// experimentation shows it can decompress our files.
	//
	// The best test we can do is verify we can decompress our own compression.
	compressed := bytes.NewBuffer(make([]byte, 0, 524288))

	// Compress the file contents
	w := NewWriter(compressed)
	_, err = io.Copy(w, f)
	if err != nil {
		t.Fatal(err)
	}

	err = w.Close()
	if err != nil {
		t.Fatal(err)
	}

	// Decompress and compare checksum
	h := sha256.New()
	reader := NewReader(compressed)

	_, err = io.Copy(h, reader)
	if err != nil {
		t.Fatal(err)
	}

	checksum := fmt.Sprintf("%x", h.Sum(nil))
	expected := "41bbdab67e6c128ab07641d85c643f5599ec221f1e7713c1eda99651f8bfe68e"

	if checksum != expected {
		t.Errorf("Checksum mismatch, want: %q, got: %q", expected, checksum)
	}
}
