package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "ut2",
	Short: "UT2004 Utility",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}
