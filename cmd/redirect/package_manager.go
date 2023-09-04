package redirect

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/cobra"

	"github.com/aldehir/ut2u/pkg/redirect"
)

var packageManager *redirect.PackageManager
var bucket string
var prefix string

func initPackageManagerArgs(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&bucket, "bucket", "b", "", "bucket to upload files")
	cmd.Flags().StringVarP(&prefix, "prefix", "p", "", "key prefix")
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
