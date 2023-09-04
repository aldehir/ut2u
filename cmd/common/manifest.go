package common

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/aldehir/ut2u/pkg/ini"
	"github.com/aldehir/ut2u/pkg/redirect"
)

var SystemDir string
var Concurrency int

func InitManifestArgs(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&SystemDir, "system", "s", "", "path to system directory")
	cmd.Flags().IntVarP(&Concurrency, "jobs", "j", -1, "number of jobs to run, defaults to number of CPUs")
}

func BuildManifest(iniFile string) (*redirect.Manifest, error) {
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
	if SystemDir == "" {
		SystemDir, _ = filepath.Split(iniFile)
	}

	builder := &redirect.ManifestBuilder{
		SystemDir:   SystemDir,
		Config:      cfg,
		Concurrency: Concurrency,
	}

	return builder.Build()
}
