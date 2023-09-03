package uz2

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"testing"
)

func TestReader(t *testing.T) {
	f, err := os.Open("testdata/the-adventures-of-sherlock-homes-by-arthur-conan-doyle.txt.uz2")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	h := sha256.New()
	reader := NewReader(f)

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
