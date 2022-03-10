// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main_test

import (
	"fmt"
	"internal/goexperiment"
	"internal/testenv"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, path, contents string) {
	if err := os.WriteFile(path, []byte(contents), 0666); err != nil {
		t.Fatalf("os.WriteFile(%s) failed: %v", path, err)
	}
}

func runPkgCover(t *testing.T, tag string, ppath, pname, pkclass string, regonly bool, mode string, infiles []string, outdir string) ([]string, string) {
	// Write the pkgcfg file.
	outcfg := filepath.Join(outdir, "outcfg.txt")
	cfgContents := fmt.Sprintf("# comment\n\npkgpath=%s\npkgname=%s\npkgclassification=%s\nregonly=%v\noutconfig=%s", ppath, pname, pkclass, regonly, outcfg)
	incfg := filepath.Join(outdir, tag+"incfg.txt")
	writeFile(t, incfg, cfgContents)

	// Form up the arguments and run the tool.
	outfiles := []string{}
	for _, inf := range infiles {
		base := filepath.Base(inf)
		outfiles = append(outfiles, filepath.Join(outdir, "cov."+base))
	}
	ofs := strings.Join(outfiles, ",")
	args := []string{"-pkgcfg", incfg, "-mode=" + mode, "-var=var" + tag, "-o", ofs}
	args = append(args, infiles...)
	cmd := exec.Command(testcover, args...)
	run(cmd, t)
	return outfiles, outcfg
}

const debugWorkDir = false

func TestCoverWithCfg(t *testing.T) {
	t.Parallel()
	testenv.MustHaveGoRun(t)
	buildCover(t)

	// Subdir in testdata that has our input files of interest.
	tpath := filepath.Join("testdata", "pkgcfg")

	// Helper to collect input paths (go files) for a subdir in 'pkgcfg'
	pfiles := func(subdir string) []string {
		de, err := os.ReadDir(filepath.Join(tpath, subdir))
		if err != nil {
			t.Fatalf("reading subdir %s: %v", subdir, err)
		}
		paths := []string{}
		for _, e := range de {
			if !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
				continue
			}
			paths = append(paths, filepath.Join(tpath, subdir, e.Name()))
		}
		return paths
	}

	dir := t.TempDir()
	if debugWorkDir {
		dir = "/tmp/qqq"
		os.RemoveAll(dir)
		os.Mkdir(dir, 0777)
	}
	instdira := filepath.Join(dir, "insta")
	if err := os.Mkdir(instdira, 0777); err != nil {
		t.Fatal(err)
	}

	// Instrument package "a" and then check to make sure the result
	// is buildable.
	ofs, outcfg := runPkgCover(t, "first", "cfg/a", "a", "mainmod", false, "set", pfiles("a"), instdira)
	t.Logf("outfiles: %+v\n", ofs)
	bargs := []string{"tool", "compile", "-p", "a", "-coveragecfg", outcfg}
	bargs = append(bargs, ofs...)
	cmd := exec.Command(testenv.GoToolPath(t), bargs...)
	cmd.Dir = instdira
	run(cmd, t)

	// For the things below we want to use the go command; check to
	// make sure the new stuff is turned on first.
	if !goexperiment.CoverageRedesign {
		t.Skipf("stubbed out due to goexperiment.CoverageRedesign=false")
	}

	// Run a "go build -cover" in the main subdir, then run
	// the resulting binary.
	cmd = exec.Command(testenv.GoToolPath(t), bargs...)
	cmd.Dir = instdira
	run(cmd, t)

	// More detail needed here.
}
