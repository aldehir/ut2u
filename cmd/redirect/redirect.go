package redirect

import "github.com/spf13/cobra"

var redirectCmd = &cobra.Command{
	Use:   "redirect",
	Short: "Run a ut2 redirect server",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func EnrichCommand(cmd *cobra.Command) {
	cmd.AddCommand(redirectCmd)
}
