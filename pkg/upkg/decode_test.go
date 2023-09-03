package upkg

import (
	"os"
	"testing"
)

func TestDecoder(t *testing.T) {
	f, err := os.Open("testdata/DM-Test.ut2")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	decoder := NewDecoder(f)
	pkg, err := decoder.Decode()
	if err != nil {
		t.Fatal(err)
	}

	deps := pkg.PackageDependencies()

	if len(deps) != 4 {
		t.Error("expected 4 depenencies")
	}

	for _, want := range []string{"2K4Chargers", "UT2004Weapons", "Engine", "XGame"} {
		if !inStringSlice(t, deps, want) {
			t.Errorf("expected to find %s in dependencies: %s", want, deps)
		}
	}

}

func inStringSlice(t *testing.T, haystack []string, needle string) bool {
	t.Helper()
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}
