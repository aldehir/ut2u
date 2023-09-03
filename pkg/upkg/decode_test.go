package upkg

import (
	"bytes"
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

	gotGUID := pkg.GUID()

	// 017b d58b 654e ee4c bef5 ae23 89ce 6d1c
	wantGUID := []byte{0x8b, 0xd5, 0x7b, 0x01, 0x4c, 0xee, 0x4e, 0x65, 0x23, 0xae, 0xf5, 0xbe, 0x1c, 0x6d, 0xce, 0x89}

	if !bytes.Equal(wantGUID, gotGUID) {
		t.Errorf("GUID mismatch, want: %v, got: %v", wantGUID, gotGUID)
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
