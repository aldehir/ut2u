package ini

import (
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParse(t *testing.T) {
	buf := strings.NewReader(exampleINI)

	cfg, err := Parse(buf)
	if err != nil {
		t.Error(err)
	}

	expect := &Config{
		Sections: []*Section{
			{
				Name: "URL",
				Items: []*Item{
					{"Protocol", []string{"ut2004"}},
					{"ProtocolDescription", []string{"Unreal Protocol"}},
				},
			},
			{
				Name: "Engine.Engine",
				Items: []*Item{
					{"RenderDevice", []string{"D3DDrv.D3DRenderDevice"}},
				},
			},
			{
				Name: "Core.System",
				Items: []*Item{
					{
						Key: "Paths",
						Values: []string{
							"../System/*.u",
							"../Maps/*.ut2",
							"../Textures/*.utx",
							"../Sounds/*.uax",
						},
					},
				},
			},
			{
				Name: "1\x1bon\x1b1\x1bDeathmatch MaplistRecord",
				Items: []*Item{
					{"DefaultTitle", []string{"1 on 1 Deathmatch"}},
					{
						Key: "DefaultMaps",
						Values: []string{
							"DM-1on1-Albatross",
							"DM-1on1-Crash",
							"DM-1on1-Desolation",
						},
					},
					{"DefaultActive", []string{"0"}},
				},
			},
		},
	}

	if diff := cmp.Diff(expect, cfg); diff != "" {
		t.Errorf("Parse() mismatch (-want +got):\n%s", diff)
	}
}

func TestParseExample(t *testing.T) {
	f, err := os.Open("testdata/Example.ini")
	if err != nil {
		t.Error(err)
	}
	defer f.Close()

	// Verify it doesn't choke on a real file
	_, err = Parse(f)
	if err != nil {
		t.Error(err)
	}
}

var exampleINI = `[URL]
Protocol=ut2004  ; Unreal Protocol
ProtocolDescription=Unreal Protocol

[Engine.Engine]
RenderDevice=D3DDrv.D3DRenderDevice
;RenderDevice=D3D9Drv.D3DRenderDevice

[Core.System]
Paths=../System/*.u
Paths=../Maps/*.ut2
Paths=../Textures/*.utx
Paths=../Sounds/*.uax

; These items use \x1b for spaces in the map list name
` + "[1\x1bon\x1b1\x1bDeathmatch MaplistRecord]" + `
DefaultTitle=1 on 1 Deathmatch
DefaultMaps=DM-1on1-Albatross
DefaultMaps=DM-1on1-Crash
DefaultActive=0
DefaultMaps=DM-1on1-Desolation
`
