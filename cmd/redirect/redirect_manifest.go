package redirect

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/aldehir/ut2u/pkg/ini"
	"github.com/aldehir/ut2u/pkg/redirect"
)

var manifestCmd = &cobra.Command{
	Use:     "manifest",
	Short:   "Generate a manifest of your UT2 packages",
	PreRunE: cobra.ExactArgs(1),
	RunE:    doManifest,
}

var systemDir string
var concurrency int

func init() {
	redirectCmd.AddCommand(manifestCmd)

	manifestCmd.Flags().StringVarP(&systemDir, "system", "s", "", "path to system directory")
	manifestCmd.Flags().IntVarP(&concurrency, "jobs", "j", -1, "number of jobs to run, defaults to number of CPUs")
}

func doManifest(cmd *cobra.Command, args []string) error {
	iniFile := args[0]

	f, err := os.Open(iniFile)
	if err != nil {
		return err
	}
	defer f.Close()

	cfg, err := ini.Parse(f)
	if err != nil {
		return err
	}

	// Infer system directory from ini file
	if systemDir == "" {
		systemDir, _ = filepath.Split(iniFile)
	}

	builder := &redirect.ManifestBuilder{
		SystemDir:   systemDir,
		Config:      cfg,
		Concurrency: concurrency,
	}

	manifest, err := builder.Build()
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	encoder.Encode(manifest)
	os.Stdout.WriteString("\n")

	return nil
}
