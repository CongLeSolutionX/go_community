// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package test

import (
	"cmd/go/internal/base"
	"cmd/go/internal/cfg"
	"cmd/go/internal/load"
	"cmd/go/internal/work"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

var coverMerge struct {
	f          *os.File
	sync.Mutex // for f.Write
}

// initCoverProfile initializes the test coverage profile.
// It must be run before any calls to mergeCoverProfile or closeCoverProfile.
// Using this function clears the profile in case it existed from a previous run,
// or in case it doesn't exist and the test is going to fail to create it (or not run).
func initCoverProfile() {
	if testCoverProfile == "" || testC {
		return
	}
	if !filepath.IsAbs(testCoverProfile) {
		testCoverProfile = filepath.Join(testOutputDir.getAbs(), testCoverProfile)
	}

	// No mutex - caller's responsibility to call with no racing goroutines.
	f, err := os.Create(testCoverProfile)
	if err != nil {
		base.Fatalf("%v", err)
	}
	_, err = fmt.Fprintf(f, "mode: %s\n", cfg.BuildCoverMode)
	if err != nil {
		base.Fatalf("%v", err)
	}
	coverMerge.f = f
}

// mergeCoverProfile merges file into the profile stored in testCoverProfile.
// It prints any errors it encounters to ew.
func mergeCoverProfile(ew io.Writer, file string) {
	if coverMerge.f == nil {
		return
	}
	coverMerge.Lock()
	defer coverMerge.Unlock()

	expect := fmt.Sprintf("mode: %s\n", cfg.BuildCoverMode)
	buf := make([]byte, len(expect))
	r, err := os.Open(file)
	if err != nil {
		// Test did not create profile, which is OK.
		return
	}
	defer r.Close()

	n, err := io.ReadFull(r, buf)
	if n == 0 {
		return
	}
	if err != nil || string(buf) != expect {
		fmt.Fprintf(ew, "error: test wrote malformed coverage profile %s.\n", file)
		return
	}
	_, err = io.Copy(coverMerge.f, r)
	if err != nil {
		fmt.Fprintf(ew, "error: saving coverage profile: %v\n", err)
	}
}

func closeCoverProfile() {
	if coverMerge.f == nil {
		return
	}
	if err := coverMerge.f.Close(); err != nil {
		base.Errorf("closing coverage profile: %v", err)
	}
}

// reportCoverageNoTestPkg reports the coverage percentage for a
// package that has no *_test.go files. This includes the usual
// percent of statements covered, but also taking care of
// -coverprofile support. The percentage reporting is done with "go
// tool covdata" which is in fact a bit of overkill (since we know a
// priori that coverage will be zero) but it helps make the "-n" and
// "-x" output more comprehensible. Here "p" is the package we're
// testing, "a" is the "test run" action for the package, and "stdout"
// is the writer to which we're sending the test output.
func reportCoverageNoTestPkg(b *work.Builder, p *load.Package, a *work.Action, stdout io.Writer) error {
	// Locate the directory containing the meta-data file fragment
	// emitted for the package by cmd/cover.
	mdir, err := buildActionMetaDir(a, p)
	if err != nil {
		return err
	}
	dirHasContent := func(d string) bool {
		f, err := os.Open(d)
		if err != nil {
			return false
		}
		defer f.Close()
		_, err = f.Readdir(1)
		if err == io.EOF {
			return false
		}
		return true
	}
	// NB: the directory in question may be empty in the case where
	// there are no functions in the package (in addition to no
	// *_test.go files); in this case the cover tool won't emit a
	// meta-data file.
	if dirHasContent(mdir) || cfg.BuildN {
		if coverMerge.f != nil {
			// Generate coverprofile fragment for this package...
			cp := a.Objdir + "_cover_.out"
			if err := b.CovData(a, "textfmt", "-i", mdir, "-o", cp); err != nil {
				return err
			}
			// ... then merge into the final output coverprofile.
			mergeCoverProfile(stdout, cp)
		}
		return b.CovDataToWriter(a, stdout, "percent", "-i", mdir)
	} else {
		fmt.Fprintf(stdout, "?   \t%s\t[no test files]\n", p.ImportPath)
	}
	return nil
}

// buildActionMetaDir locates and returns the meta-data file written
// by the "go tool cover" step as part of the build action for
// a given "go test -cover" run action.
func buildActionMetaDir(runAct *work.Action, p *load.Package) (string, error) {
	// We expect one of two cases here: either a build action as a
	// prededessor (in the simple case) or the 'writeMetaFiles' dummy
	// action (in the case where -coverpkg=... is in effect).

	// Try to locate the "write-meta-files" action first. If we find it,
	// use that as the action to examine for the correct "build" action
	// predecessor. If we don't find the meta-action, just examine
	// the preds of the run action.
	cur := runAct
	for i := range cur.Deps {
		pred := cur.Deps[i]
		if pred.Mode == writeMetaActMode {
			cur = pred
			break
		}
	}
	for i := range cur.Deps {
		pred := cur.Deps[i]
		if pred.Mode != "build" || pred.Package == nil {
			continue
		}
		if pred.Package.ImportPath == p.ImportPath {
			return work.CovMetaDestDir(pred), nil
		}
	}
	return "", fmt.Errorf("internal error: unable to locate build action for package %q run action", p.ImportPath)
}
