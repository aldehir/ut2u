package cmd

import (
	"github.com/spf13/cobra"

	"github.com/aldehir/ut2u/cmd/query"
	"github.com/aldehir/ut2u/cmd/redirect"
	"github.com/aldehir/ut2u/cmd/upackage"
)

var rootCmd = &cobra.Command{
	Use:   "ut2u",
	Short: "UT2004 Utility",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Usage()
	},
}

func init() {
	redirect.EnrichCommand(rootCmd)
	upackage.EnrichCommand(rootCmd)
	query.EnrichCommand(rootCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
