// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main_test

import (
	"fmt"
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

func writePkgConfig(t *testing.T, outdir, tag, ppath, pname, pkclass string, regonly bool) string {
	outcfg := filepath.Join(outdir, "outcfg.txt")
	cfgContents := fmt.Sprintf("# comment\n\npkgpath=%s\npkgname=%s\npkgclassification=%s\nregonly=%v\noutconfig=%s", ppath, pname, pkclass, regonly, outcfg)
	incfg := filepath.Join(outdir, tag+"incfg.txt")
	writeFile(t, incfg, cfgContents)
	return incfg
}

func runPkgCover(t *testing.T, outdir string, tag string, incfg string, mode string, infiles []string, errExpected bool) ([]string, string, string) {
	// Write the pkgcfg file.
	outcfg := filepath.Join(outdir, "outcfg.txt")

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
	if errExpected {
		errmsg := runExpectingError(cmd, t)
		return nil, "", errmsg
	} else {
		run(cmd, t)
		return outfiles, outcfg, ""
	}
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

	// Instrument package "a".
	tag := "first"
	ppath := "cfg/a"
	pname := "a"
	pkgclass := "mainmod"
	regonly := false
	mode := "set"
	incfg := writePkgConfig(t, instdira, tag, ppath, pname, pkgclass, regonly)
	ofs, outcfg, _ := runPkgCover(t, instdira, tag, incfg, mode,
		pfiles("a"), false)
	t.Logf("outfiles: %+v\n", ofs)

	// Run the compiler on the files to make sure the result is
	// buildable.
	bargs := []string{"tool", "compile", "-p", "a", "-coveragecfg", outcfg}
	bargs = append(bargs, ofs...)
	cmd := exec.Command(testenv.GoToolPath(t), bargs...)
	cmd.Dir = instdira
	run(cmd, t)

	// Do some error testing to ensure that various bad options and
	// combinations are properly rejected.

	// Expect error if config file inaccessible/unreadable.
	errExpected := true
	_, _, errmsg := runPkgCover(t, instdira, tag, "/not/a/file", mode,
		pfiles("a"), errExpected)
	want := "error reading pkgconfig file"
	if !strings.Contains(errmsg, want) {
		t.Errorf("'bad config file' test: wanted %s got %s", want, errmsg)
	}

	//
}
