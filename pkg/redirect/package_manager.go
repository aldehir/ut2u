package redirect

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/aldehir/ut2u/pkg/uz2"
)

const (
	MaxMultipartUploadSize = 5 * 1024 * 1024
)

type PackageManager struct {
	Bucket string
	Prefix string

	s3Client *s3.Client
}

func NewPackageManager(client *s3.Client, bucket string, prefix string) *PackageManager {
	return &PackageManager{
		s3Client: client,
		Bucket:   bucket,
		Prefix:   prefix,
	}
}

func (b *PackageManager) Upload(ctx context.Context, pkg PackageMeta) error {
	f, err := os.Open(pkg.Path)
	if err != nil {
		return err
	}
	defer f.Close()

	ctx, cancel := context.WithCancelCause(ctx)

	// Create a pipe, writing the compressed object for the upload function to
	// read from the reader
	r, w := io.Pipe()

	go func() {
		compressor := uz2.NewWriter(w)
		_, err = io.Copy(compressor, f)
		if err != nil {
			cancel(err)
			return
		}

		err = compressor.Close()
		if err != nil {
			cancel(err)
		}

		w.Close()
	}()

	key := b.packageKey(pkg)

	uploader := manager.NewUploader(b.s3Client, func(u *manager.Uploader) {
		u.PartSize = 10 * 1024 * 1024
	})

	_, err = uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:   &b.Bucket,
		Key:      &key,
		Metadata: s3Metadata(pkg),
		Body:     r,
	})

	if errors.Is(err, context.Canceled) {
		return ctx.Err()
	} else if err != nil {
		return err
	}

	return nil
}

func (b *PackageManager) packageKey(pkg PackageMeta) string {
	return filepath.Join(b.Prefix, pkg.Name, pkg.GUID)
}

// Encode the package meta to a map for storing in S3
func s3Metadata(pkg PackageMeta) map[string]string {
	data := make(map[string]string)
	data["name"] = pkg.Name
	data["guid"] = pkg.GUID
	data["provides"] = pkg.Provides
	data["requires"] = strings.Join(pkg.Requires, ";")
	data["uncompressed-checksum-md5"] = pkg.Checksums.MD5
	data["uncompressed-checksum-sha1"] = pkg.Checksums.SHA1
	data["uncompressed-checksum-sha256"] = pkg.Checksums.SHA256
	return data
}
