package upackage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/aldehir/ut2u/cmd/common"
)

var pkgCmd = &cobra.Command{
	Use:   "package",
	Short: "Manage packages",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Usage()
	},
}

func init() {
	pkgCmd.AddCommand(checkCmd)
	common.InitManifestArgs(checkCmd)

	pkgCmd.AddCommand(requiresCmd)
	common.InitManifestArgs(requiresCmd)
}

func EnrichCommand(cmd *cobra.Command) {
	cmd.AddCommand(pkgCmd)
}

var checkCmd = &cobra.Command{
	Use:   "check-deps [-s system-dir] ut2004-ini",
	Short: "Check package dependencies",
	Args:  cobra.ExactArgs(1),
	RunE:  doCheck,

	DisableFlagsInUseLine: true,
}

func doCheck(cmd *cobra.Command, args []string) error {
	manifest, err := common.BuildManifest(args[0])
	if err != nil {
		return err
	}

	passed := true

	provided := make(map[string]struct{})
	for _, p := range manifest.Packages {
		provided[strings.ToLower(p.Provides)] = struct{}{}
	}

	missing := make([]string, 0, 10)

	for _, p := range manifest.Packages {
		requires := p.Requires
		missing = missing[:0]

		for _, r := range requires {
			if _, found := provided[strings.ToLower(r)]; found {
				continue
			}

			missing = append(missing, r)
			passed = false
		}

		if len(missing) > 0 {
			fmt.Fprintf(os.Stderr, "Package %s has missing dependencies: %s\n", p.Name, strings.Join(missing, ", "))
		}
	}

	if !passed {
		fmt.Fprintf(os.Stderr, "Some packages have missing dependencies\n")
		os.Exit(1)
	}

	return nil
}

var requiresCmd = &cobra.Command{
	Use:   "requires [-s system-dir] ut2004-ini package",
	Short: "Find package dependents",
	Args:  cobra.ExactArgs(2),
	RunE:  doRequires,

	DisableFlagsInUseLine: true,
}

func doRequires(cmd *cobra.Command, args []string) error {
	manifest, err := common.BuildManifest(args[0])
	if err != nil {
		return err
	}

	pkgName := args[1]
	pkgName = strings.TrimSuffix(pkgName, filepath.Ext(pkgName)) // Remove extension

	for _, p := range manifest.Packages {
		for _, r := range p.Requires {
			if strings.EqualFold(r, pkgName) {
				fmt.Fprintf(os.Stdout, "%s\n", p.Name)
			}
		}
	}

	return nil
}
