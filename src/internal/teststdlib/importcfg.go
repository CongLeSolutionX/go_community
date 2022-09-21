package teststdlib

import (
	"go/importer"
	"go/token"
	"go/types"
	"internal/goroot"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
)

func Importer() types.Importer {
	return importer.ForCompiler(token.NewFileSet(), runtime.Compiler, lookupStdlib)
}

// lookupStdlib is a lookup function to be used by a types.Importer.
// It assumes that vendored paths are being imported from the stdlib,
// outside cmd.
func lookupStdlib(pkgpath string) (io.ReadCloser, error) {
	pkgpath = filepath.ToSlash(pkgpath)
	m, err := goroot.PkgfileMap()
	if err != nil {
		return nil, err
	}
	p, ok := m[pkgpath]
	if !ok {
		p = m[path.Join("vendor", pkgpath)]
	}
	return os.Open(p)
}
