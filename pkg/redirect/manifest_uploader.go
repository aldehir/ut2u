package redirect

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/sync/errgroup"
)

type ManifestUploader struct {
	// Number of active uploads. If zero, it uses DefaultManifestUploaderConcurrency
	Concurrency int

	pm *PackageManager
}

var (
	DefaultManifestUploaderConcurrency = 5
)

type ManifestUploaderOption func(u *ManifestUploader)

// NewManifestUploader returns a ManifestUploader capable of uploading an
// entire manifest of packages. ManifestUploader will not overwrite existing
// files.
func NewManifestUploader(pm *PackageManager, opts ...ManifestUploaderOption) *ManifestUploader {
	uploader := &ManifestUploader{pm: pm}
	for _, fn := range opts {
		fn(uploader)
	}
	return uploader
}

func (u *ManifestUploader) Upload(ctx context.Context, manifest *Manifest) error {
	g, ctx := errgroup.WithContext(ctx)

	concurrency := u.Concurrency
	if concurrency <= 0 {
		concurrency = DefaultManifestUploaderConcurrency
	}

	sem := make(chan struct{}, concurrency)

	for _, pkg := range manifest.Packages {
		pkg := pkg // Avoid the late binding bug :)

		g.Go(func() error {
			sem <- struct{}{}
			defer func() {
				<-sem
			}()

			exists, err := u.pm.Exists(ctx, pkg)
			if err != nil {
				return err
			}

			if exists {
				return nil
			}

			fmt.Fprintf(os.Stderr, "Uploading %s\n", pkg.Name)
			err = u.pm.Upload(ctx, pkg)
			if err != nil {
				return err
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}
