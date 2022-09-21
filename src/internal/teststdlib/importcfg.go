package teststdlib

import (
	"bytes"
	"fmt"
	"go/importer"
	"go/token"
	"go/types"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
)

// CreateStdlibImportcfg returns an importcfg file to be passed to the
// Go compiler that contains the cached paths for the .a files for the
// standard library..
func CreateStdlibImportcfg() (string, error) {
	var icfg bytes.Buffer

	m, err := StdlibPkgfileMap()
	if err != nil {
		return "", nil
	}
	fmt.Fprintf(&icfg, "# import config")
	for importPath, export := range m {
		if importPath != "unsafe" && export != "" { // unsafe
			fmt.Fprintf(&icfg, "\npackagefile %s=%s", importPath, export)
		}
	}
	s := icfg.String()
	return s, nil
}

var (
	stdlibPkgfileMap map[string]string
	stdlibPkgfileErr error
	once             sync.Once
)

func isRace() bool {
	info, _ := debug.ReadBuildInfo()
	if info == nil {
		panic("missing build info")
	}
	for _, s := range info.Settings {
		if s.Key == "-race" && s.Value == "true" {
			return true
		}
	}
	return false
}

func StdlibPkgfileMap() (map[string]string, error) {
	once.Do(func() {
		m := make(map[string]string)
		output, err := exec.Command("go", "list", "-export", "-f", "{{.ImportPath}} {{.Export}}", "std", "cmd/...", "cmd/vendor/...").Output()
		if err != nil {
			stdlibPkgfileErr = err
		}
		for _, line := range strings.Split(string(output), "\n") {
			if line == "" {
				continue
			}
			sp := strings.SplitN(line, " ", 2)
			importPath, export := sp[0], sp[1]
			m[importPath] = export
		}
		stdlibPkgfileMap = m
	})
	return stdlibPkgfileMap, stdlibPkgfileErr
}

func Importer() types.Importer {
	return importer.ForCompiler(token.NewFileSet(), runtime.Compiler, lookupStdlib)
}

// lookupStdlib is a lookup function to be used by a types.Importer.
// It assumes that vendored paths are being imported from the stdlib,
// outside cmd.
func lookupStdlib(pkgpath string) (io.ReadCloser, error) {
	pkgpath = filepath.ToSlash(pkgpath)
	m, err := StdlibPkgfileMap()
	if err != nil {
		return nil, err
	}
	p, ok := m[pkgpath]
	if !ok {
		p = m[path.Join("vendor", pkgpath)]
	}
	return os.Open(p)
}
