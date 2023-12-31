package redirect

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"

	"github.com/aldehir/ut2u/cmd/common"
)

var manifestCmd = &cobra.Command{
	Use:   "manifest [-s system-dir] ut2004-ini",
	Short: "Generate a manifest of packages",
	Args:  cobra.ExactArgs(1),
	RunE:  doManifest,

	DisableFlagsInUseLine: true,
}

func init() {
	redirectCmd.AddCommand(manifestCmd)
	common.InitManifestArgs(manifestCmd)
}

func doManifest(cmd *cobra.Command, args []string) error {
	manifest, err := common.BuildManifest(args[0])
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	encoder.Encode(manifest)
	os.Stdout.WriteString("\n")

	return nil
}
