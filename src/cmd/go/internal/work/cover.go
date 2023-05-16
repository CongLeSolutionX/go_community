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
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"internal/coverage"
	"io"
	"os"
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
// file fragment for the package (used for -coverpkg=./... and to
// help with packages that have functions but no tests).
func genCovMetaFile(a *Action) bool {
	p := a.Package
	return p.Internal.CoverMode != "" && p.Internal.WriteCovMeta
}

// CovMetaDestDir returns the destination directory into which a
// meta-data file will be written for a given package if the package
// doesn't have any *_test.go files.
func CovMetaDestDir(a *Action) string {
	return a.Objdir + "covdata"
}

// CovMetaDestPath returns the destination path used to store a
// meta-data file will be written for a given package if the package
// doesn't have any *_test.go files.
func CovMetaDestPath(a *Action) string {
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

// WriteMetaFilesFile is the action function for the "writeMeta"
// pseudo action employed during "go test -coverpkg" runs where
// there are multiple tests and multiple packages covered. It
// builds up a table mapping package import path to meta-data file
// fragment and writes it out to a file where it can be read by
// the various test run actions.
func WriteMetaFilesFile(b *Builder, ctx context.Context, a *Action) error {
	// Build the metafilecollection object.
	var mfc coverage.MetaFileCollection
	for i := range a.Deps {
		dep := a.Deps[i]
		// expect only build actions
		if dep.Mode != "build" {
			panic("unexpected mode " + dep.Mode)
		}
		mff := CovMetaDestPath(dep)
		// Check to make sure the meta-data file fragment exists
		// (may be missing if package has no functions).
		if _, err := os.Stat(mff); err != nil {
			continue
		}
		mfc.ImportPaths = append(mfc.ImportPaths,
			dep.Package.ImportPath)
		mfc.MetaFileFragments = append(mfc.MetaFileFragments,
			mff)
	}

	// Serialize it.
	data, err := json.Marshal(mfc)
	if err != nil {
		return fmt.Errorf("marshal MetaFileCollection: %v", err)
	}
	data = append(data, '\n') // makes -x output more readable

	// Create a new objdir and write out the serialized collection
	// to a file in the new tempdir, then record the path of the
	// file generated as the target of this action.
	obd := b.NewObjdir()
	if err := b.Mkdir(obd); err != nil {
		return err
	}
	mfpath := obd + a.Target
	if err := b.writeFile(mfpath, data); err != nil {
		return fmt.Errorf("writing metafiles file: %v", err)
	}
	a.Target = mfpath

	// We're done.
	return nil
}

// CopyMetaFilesFile is invoked during "go test -coverpkg" runs where
// there are multiple tests and multiple packages covered; it copies
// the file generated by the "writeMeta" action into the temp coverdir
// of a coverage test that is about to run.
func (b *Builder) CopyMetaFilesFile(writeMetaAct *Action, testGoCoverDir string) error {
	fname := filepath.Base(writeMetaAct.Target)
	dst := filepath.Join(testGoCoverDir, fname)
	return b.copyFile(dst, writeMetaAct.Target, 0666, false)
}
