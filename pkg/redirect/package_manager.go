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
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"golang.org/x/sync/errgroup"

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

// Upload compresses a package and uploads it to bucket/prefix/<name>/<guid>.uz2
func (b *PackageManager) Upload(ctx context.Context, pkg PackageMeta) error {
	f, err := os.Open(pkg.Path)
	if err != nil {
		return err
	}
	defer f.Close()

	g, ctx := errgroup.WithContext(ctx)

	// Create a pipe, writing the compressed object for the upload function to
	// read from the reader
	r, w := io.Pipe()

	g.Go(func() error {
		compressor := uz2.NewWriter(w)
		defer w.Close()

		_, err = io.Copy(compressor, f)
		if err != nil {
			return err
		}

		err = compressor.Close()
		if err != nil {
			return err
		}

		return nil
	})

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

	if err != nil {
		r.Close() // Close the reader so our goroutine finds a way out
	}

	g.Wait()
	return err
}

// GetPackageGUIDs returns all package GUIDs on the redirect server.
func (p *PackageManager) GetPackageGUIDs(ctx context.Context) ([]string, error) {
	prefix := p.Prefix
	if !strings.HasSuffix(prefix, "/") {
		prefix = prefix + "/"
	}

	paginator := s3.NewListObjectsV2Paginator(p.s3Client, &s3.ListObjectsV2Input{
		Bucket: &p.Bucket,
		Prefix: &prefix,
	})

	result := make([]string, 0, 20)

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, obj := range output.Contents {
			key := strings.TrimPrefix(*obj.Key, prefix)
			_, filename := filepath.Split(key)
			guid := strings.TrimSuffix(filename, filepath.Ext(filename))

			result = append(result, guid)
		}
	}

	return result, nil
}

// Exists returns true if the given package is already on the redirect server.
func (p *PackageManager) Exists(ctx context.Context, pkg PackageMeta) (bool, error) {
	key := p.packageKey(pkg)

	output, err := p.s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &p.Bucket,
		Key:    &key,
	})

	var nsk *types.NotFound
	if errors.As(err, &nsk) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	// If it does exist, then compare the checksums and see if we need to upload it
	metadata := output.Metadata
	if otherSum, ok := metadata["uncompressed-checksum-sha256"]; ok {
		if !strings.EqualFold(otherSum, pkg.Checksums.SHA256) {
			return false, nil
		}
	}

	return true, nil
}

func (b *PackageManager) packageKey(pkg PackageMeta) string {
	return filepath.Join(b.Prefix, pkg.Name+".uz2", pkg.GUID)
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
