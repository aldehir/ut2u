package upackage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/aldehir/ut2u/cmd/common"
	"github.com/aldehir/ut2u/pkg/redirect"
	"github.com/aldehir/ut2u/pkg/uz2"
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

	pkgCmd.AddCommand(infoCmd)
	pkgCmd.AddCommand(compressCmd)
	pkgCmd.AddCommand(decompressCmd)
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

var infoCmd = &cobra.Command{
	Use:   "info package...",
	Short: "Print package information",
	Args:  cobra.MinimumNArgs(1),
	RunE:  doInfo,

	DisableFlagsInUseLine: true,
}

func doInfo(cmd *cobra.Command, args []string) error {
	for i, p := range args {
		if i > 0 {
			fmt.Fprintln(os.Stdout)
		}

		printPackageInfo(p)
	}

	return nil
}

func printPackageInfo(path string) {
	info, err := redirect.ReadPackageMeta(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading %s: %s", path, err)
		return
	}

	fmt.Fprintf(os.Stdout, "Name:     %s\n", info.Name)
	fmt.Fprintf(os.Stdout, "GUID:     %s\n", info.GUID)
	fmt.Fprintf(os.Stdout, "Provides: %s\n", info.Provides)

	if len(info.Requires) > 0 {
		fmt.Fprintf(os.Stdout, "Requires:\n")
		for _, req := range info.Requires {
			fmt.Fprintf(os.Stdout, "  - %s\n", req)
		}
	}

	fmt.Fprintf(os.Stdout, "Checksums: \n")
	fmt.Fprintf(os.Stdout, "  MD5:    %s\n", info.Checksums.MD5)
	fmt.Fprintf(os.Stdout, "  SHA1:   %s\n", info.Checksums.SHA1)
	fmt.Fprintf(os.Stdout, "  SHA256: %s\n", info.Checksums.SHA256)
}

var compressCmd = &cobra.Command{
	Use:   "compress package",
	Short: "Compress package",
	Args:  cobra.ExactArgs(1),
	RunE:  doCompress,

	DisableFlagsInUseLine: true,
}

func doCompress(cmd *cobra.Command, args []string) error {
	p := args[0]

	result, err := compressPackage(p)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error compressing %s: %s\n", p, err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stdout, "%s -> %s\n", p, result)
	return nil
}

var decompressCmd = &cobra.Command{
	Use:   "decompress package",
	Short: "Decompress package",
	Args:  cobra.ExactArgs(1),
	RunE:  doDecompress,

	DisableFlagsInUseLine: true,
}

func doDecompress(cmd *cobra.Command, args []string) error {
	p := args[0]

	result, err := decompressPackage(p)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error decompressing %s: %s\n", p, err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stdout, "%s -> %s\n", p, result)
	return nil
}

func compressPackage(path string) (string, error) {
	result := path + ".uz2"

	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	out, err := os.Create(result)
	if err != nil {
		return "", err
	}
	defer out.Close()

	w := uz2.NewWriter(out)
	_, err = io.Copy(w, f)
	if err != nil {
		return "", err
	}

	err = w.Close()
	if err != nil {
		return "", err
	}

	return result, nil
}

func decompressPackage(path string) (string, error) {
	var result = path + ".out"

	if strings.HasSuffix(path, ".uz2") {
		result = strings.TrimSuffix(path, ".uz2")
	}

	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	out, err := os.Create(result)
	if err != nil {
		return "", err
	}
	defer out.Close()

	r := uz2.NewReader(f)
	_, err = io.Copy(out, r)
	if err != nil {
		return "", err
	}

	return result, nil
}
