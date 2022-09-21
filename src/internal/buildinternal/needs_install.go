// This package provides internal functions used by go/build
// that need to be used by other packages too.
package buildinternal

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"sync"
)

// NeedsInstalledDotA returns true if the given stdlib package
// needs an installed .a file in the stdlib.
func NeedsInstalledDotA(importPath string) bool {
	return importPath == "net" || importPath == "os/signal" || importPath == "os/user" || importPath == "plugin" ||
		importPath == "runtime/cgo"
}

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
