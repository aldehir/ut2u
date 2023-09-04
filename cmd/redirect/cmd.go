package redirect

import "github.com/spf13/cobra"

var redirectCmd = &cobra.Command{
	Use:   "redirect",
	Short: "Manage a UT2004 redirect server",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Usage()
	},
}

func EnrichCommand(cmd *cobra.Command) {
	cmd.AddCommand(redirectCmd)
}
