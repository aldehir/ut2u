package cmd

import (
	"github.com/spf13/cobra"

	"github.com/aldehir/ut2u/cmd/redirect"
)

var rootCmd = &cobra.Command{
	Use:   "ut2u",
	Short: "UT2004 Utility",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	redirect.EnrichCommand(rootCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
