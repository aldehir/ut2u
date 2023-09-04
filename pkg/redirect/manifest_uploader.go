package redirect

import (
	"context"
	"fmt"
	"os"
	"strings"

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

	// Build a map of GUIDs that exist on the server
	guids, err := u.pm.GetPackageGUIDs(ctx)
	if err != nil {
		return err
	}

	guidSet := make(map[string]struct{}, len(guids))
	for _, guid := range guids {
		guidSet[strings.ToUpper(guid)] = struct{}{}
	}

	sem := make(chan struct{}, concurrency)

	for _, pkg := range manifest.Packages {
		pkg := pkg // Avoid the late binding bug :)

		g.Go(func() error {
			sem <- struct{}{}
			defer func() {
				<-sem
			}()

			_, exists := guidSet[strings.ToUpper(pkg.GUID)]
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
