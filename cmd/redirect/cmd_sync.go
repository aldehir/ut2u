package redirect

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/aldehir/ut2u/cmd/common"
	"github.com/aldehir/ut2u/pkg/redirect"
)

var syncCmd = &cobra.Command{
	Use:     "sync [-b bucket] [-p prefix] [-s system-dir] ut2004-ini",
	Short:   "Upload packages found from a UT2004.ini to an S3 bucket",
	Args:    cobra.ExactArgs(1),
	PreRunE: withPackageManager,
	RunE:    doSync,

	DisableFlagsInUseLine: true,
}

var concurrentUploads int

func init() {
	redirectCmd.AddCommand(syncCmd)
	initPackageManagerArgs(syncCmd)
	common.InitManifestArgs(syncCmd)

	syncCmd.Flags().IntVarP(&concurrentUploads, "upload-jobs", "u", 0, "number of concurrent uploads")
}

func doSync(cmd *cobra.Command, args []string) error {
	manifest, err := common.BuildManifest(args[0])
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Found %d packages\n", len(manifest.Packages))

	uploader := redirect.NewManifestUploader(packageManager, func(u *redirect.ManifestUploader) {
		u.Concurrency = concurrentUploads
	})

	return uploader.Upload(context.TODO(), manifest)
}
