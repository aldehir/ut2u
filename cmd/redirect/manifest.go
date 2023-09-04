package redirect

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/aldehir/ut2u/pkg/ini"
	"github.com/aldehir/ut2u/pkg/redirect"
)

var systemDir string
var concurrency int

func initManifestArgs(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&systemDir, "system", "s", "", "path to system directory")
	cmd.Flags().IntVarP(&concurrency, "jobs", "j", -1, "number of jobs to run, defaults to number of CPUs")
}

func buildManifest(iniFile string) (*redirect.Manifest, error) {
	f, err := os.Open(iniFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	cfg, err := ini.Parse(f)
	if err != nil {
		return nil, err
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

	return builder.Build()
}
