package redirect

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/aldehir/ut2u/pkg/redirect"
)

var uploadCmd = &cobra.Command{
	Use:     "upload [-b bucket] [-p prefix] packages...",
	Short:   "Upload package to an S3 bucket",
	Args:    cobra.MinimumNArgs(1),
	PreRunE: withPackageManager,
	RunE:    doUpload,

	DisableFlagsInUseLine: true,
}

func init() {
	redirectCmd.AddCommand(uploadCmd)
	initPackageManagerArgs(uploadCmd)
}

func doUpload(cmd *cobra.Command, args []string) error {
	for _, file := range args {
		meta, err := redirect.ReadPackageMeta(file)
		if err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "Uploading %s...\n", file)
		err = packageManager.Upload(context.TODO(), meta)
		if err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "Uploaded!\n")
	}

	return nil
}
