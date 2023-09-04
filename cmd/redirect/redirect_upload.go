package redirect

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/cobra"

	"github.com/aldehir/ut2u/pkg/redirect"
)

var uploadCmd = &cobra.Command{
	Use:     "upload",
	Short:   "Upload a package to an S3 bucket",
	Args:    cobra.MinimumNArgs(1),
	PreRunE: withPackageManager,
	RunE:    doUpload,
}

var packageManager *redirect.PackageManager
var bucket string
var prefix string

func init() {
	redirectCmd.AddCommand(uploadCmd)
	uploadCmd.Flags().StringVarP(&bucket, "bucket", "b", "", "bucket to upload files")
	uploadCmd.Flags().StringVarP(&prefix, "prefix", "p", "", "key prefix")
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

func withPackageManager(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return err
	}

	s3Client := s3.NewFromConfig(cfg)
	packageManager = redirect.NewPackageManager(s3Client, bucket, prefix)
	return nil
}
