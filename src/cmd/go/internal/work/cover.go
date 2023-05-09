// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Action graph execution methods related to coverage.

package work

import (
	"cmd/go/internal/base"
	"cmd/go/internal/cache"
	"cmd/go/internal/cfg"
	"cmd/go/internal/str"
	"crypto/sha256"
	"fmt"
	"internal/coverage"
	"io"
	"path/filepath"
)

// covMetaFileName returns the name of the meta-data file to generate
// for a given package whose build action is "a". It is assumed that
// genCovMetaFile(a) returns TRUE for this package.
func covMetaFileName(a *Action) string {
	var r [32]byte
	sum := sha256.Sum256([]byte(a.Package.ImportPath))
	copy(r[:], sum[:])
	return coverage.MetaFilePref + fmt.Sprintf(".%x", r)
}

// genCovMetaFile returns TRUE if the compile action 'a' has coverage
// enabled and we want to have the cover tool emit a static meta-data
// file fragment for the package (needed for helping with packages
// that have functions but no tests).
func genCovMetaFile(a *Action) bool {
	p := a.Package
	return p.Internal.CoverMode != "" &&
		len(p.TestGoFiles)+len(p.XTestGoFiles) == 0

}

// CovMetaDestDir returns the destination directory into which a
// meta-data file will be written for a given package if the package
// doesn't have any *_test.go files.
func CovMetaDestDir(a *Action) string {
	return a.Objdir + "covdata"
}

// covMetaDestPath returns the destination path used to store a
// meta-data file will be written for a given package if the package
// doesn't have any *_test.go files.
func covMetaDestPath(a *Action) string {
	return filepath.Join(CovMetaDestDir(a), covMetaFileName(a))
}

// cacheCovMetaFile caches the coverage meta-data file
// fragment for a "go test -cover" package build.
func (b *Builder) cacheCovMetaFile(build *Action) {
	c := cache.Default()
	mfn := covMetaFileName(build)
	b.cacheObjdirFile(build, c, mfn)
}

// loadCachedCovMetaFile loads the coverage meta-data file
// fragment for a "go test -cover" package build that hits
// in the cache.
func (b *Builder) loadCachedCovMetaFile(build *Action) error {
	c := cache.Default()
	mfn := covMetaFileName(build)
	return b.loadCachedObjdirFile(build, c, mfn)
}

// CovDataToWriter invokes "go tool covdata" with the specified
// arguments as part of the execution of action "a", then writes the
// resulting output to the writer 'w', formatting if needed.
func (b *Builder) CovDataToWriter(a *Action, w io.Writer, cmdargs ...any) error {
	cmdline := str.StringList(cmdargs...)
	args := append([]string{}, cfg.BuildToolexec...)
	args = append(args, base.Tool("covdata"))
	args = append(args, cmdline...)
	output, err := b.runOut(a, a.Objdir, nil, args)
	if err != nil {
		p := a.Package
		return formatOutput(b.WorkDir, p.Dir, p.ImportPath, p.Desc(), string(output))
	}
	_, werr := w.Write(output)
	return werr
}

// CovData invokes "go tool covdata" with the specified arguments
// as part of the execution of action "a".
func (b *Builder) CovData(a *Action, cmdargs ...any) error {
	cmdline := str.StringList(cmdargs...)
	args := append([]string{}, cfg.BuildToolexec...)
	args = append(args, base.Tool("covdata"))
	args = append(args, cmdline...)
	return b.run(a, a.Objdir, "", nil, args)
}
