package redirect

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/aldehir/ut2u/pkg/ini"
)

func TestManifestBuilder(t *testing.T) {
	iniFile, ok := os.LookupEnv("MANIFEST_BUILD_INI_FILE")
	if !ok {
		t.Skip("Pass MANIFEST_BUILD_INI_FILE to run")
	}

	systemDir, ok := os.LookupEnv("MANIFEST_BUILD_SYSTEM_DIR")
	if !ok {
		t.Skip("Pass MANIFEST_BUILD_SYSTEM_DIR to run")
	}

	f, err := os.Open(iniFile)
	if err != nil {
		t.Error(err)
	}
	defer f.Close()

	cfg, err := ini.Parse(f)
	if err != nil {
		t.Error(err)
	}

	builder := &ManifestBuilder{
		SystemDir: systemDir,
		Config:    cfg,
	}

	manifest, err := builder.Build()
	if err != nil {
		t.Error(err)
	}

	asJSON, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Error(err)
	}

	t.Log(string(asJSON))
}
