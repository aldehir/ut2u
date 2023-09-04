package redirect

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/aldehir/ut2u/pkg/ini"
	"github.com/aldehir/ut2u/pkg/upkg"
)

type Manifest struct {
	Version  string        `json:"version"`
	Packages []PackageMeta `json:"packages"`
}

type PackageMeta struct {
	Path      string `json:"-"`
	Name      string `json:"name"`
	GUID      string `json:"guid"`
	Checksums struct {
		// UE still uses MD5, so it might be helpful to keep this around
		MD5    string `json:"md5"`
		SHA1   string `json:"sha1"`
		SHA256 string `json:"sha256"`
	} `json:"checksums"`

	// Dependency management
	Provides string   `json:"provides"`
	Requires []string `json:"requires"`
}

func ReadPackageMeta(file string) (PackageMeta, error) {
	f, err := os.Open(file)
	if err != nil {
		return PackageMeta{}, err
	}

	decoder := upkg.NewDecoder(f)
	pkg, err := decoder.Decode()
	if err != nil {
		return PackageMeta{}, fmt.Errorf("failed to decode package %s, %w", file, err)
	}

	// Seek to the start and compute checksums
	f.Seek(0, io.SeekStart)

	hashMD5 := md5.New()
	hashSHA1 := sha1.New()
	hashSHA256 := sha256.New()

	hash := io.MultiWriter(hashMD5, hashSHA1, hashSHA256)

	_, err = io.Copy(hash, f)
	if err != nil {
		return PackageMeta{}, fmt.Errorf("failed to compute checksums for %s, %w", file, err)
	}

	var meta PackageMeta
	meta.Path = file
	meta.Name = filepath.Base(file)
	meta.GUID = fmt.Sprintf("%X", pkg.GUID())
	meta.Checksums.MD5 = fmt.Sprintf("%x", hashMD5.Sum(nil))
	meta.Checksums.SHA1 = fmt.Sprintf("%x", hashSHA1.Sum(nil))
	meta.Checksums.SHA256 = fmt.Sprintf("%x", hashSHA256.Sum(nil))

	// TODO: Perhaps get this from the package instead of inferring it from the
	// name?
	meta.Provides = strings.TrimSuffix(meta.Name, filepath.Ext(meta.Name))
	meta.Requires = pkg.PackageDependencies()

	return meta, nil
}

type ManifestBuilder struct {
	// SystemDir is the path to the UT2004 System directory. It is necessary as
	// paths are relative to the system directory
	SystemDir   string
	Config      *ini.Config
	Concurrency int

	files []string
	jobs  chan string
	sem   chan struct{}
	wg    sync.WaitGroup

	packages      []PackageMeta
	packagesMutex sync.Mutex
}

// BuildManifest returns a Manifest generated from inspecting Paths in the
// given configuration
func (b *ManifestBuilder) Build() (*Manifest, error) {
	if err := b.findPackages(); err != nil {
		return nil, err
	}

	b.spawnWorkers()

	// Sort packages
	sort.Slice(b.packages, func(i, j int) bool {
		return b.packages[i].Name < b.packages[j].Name
	})

	return &Manifest{
		Version:  "1",
		Packages: b.packages,
	}, nil
}

func (b *ManifestBuilder) spawnWorkers() {
	concurrency := b.Concurrency
	if concurrency <= 0 {
		concurrency = runtime.NumCPU()
	}

	fmt.Fprintf(os.Stderr, "Running %d concurrent jobs\n", concurrency)

	b.sem = make(chan struct{}, concurrency)
	defer close(b.sem)

	for _, file := range b.files {
		b.wg.Add(1)

		go func(file string) {
			b.sem <- struct{}{}
			b.processFile(file)
			<-b.sem
			b.wg.Done()
		}(file)
	}

	b.wg.Wait()
}

func (b *ManifestBuilder) processFile(file string) {
	fmt.Fprintf(os.Stderr, "Processing: %s\n", file)

	pkgMeta, err := ReadPackageMeta(file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}

	b.packagesMutex.Lock()
	defer b.packagesMutex.Unlock()
	b.packages = append(b.packages, pkgMeta)
}

func (b *ManifestBuilder) findPackages() error {
	paths, ok := b.Config.Values("Core.System", "Paths")
	if !ok {
		return errors.New("no Paths in Core.System section")
	}

	for _, p := range paths {
		pattern := filepath.Join(b.SystemDir, p)
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return err
		}

		b.files = append(b.files, matches...)
	}

	return nil
}
