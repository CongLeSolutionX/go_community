// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main_test

import (
	"encoding/json"
	"fmt"
	"internal/coverage"
	"internal/testenv"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, path string, contents []byte) {
	if err := os.WriteFile(path, contents, 0666); err != nil {
		t.Fatalf("os.WriteFile(%s) failed: %v", path, err)
	}
}

func writePkgConfig(t *testing.T, outdir, tag, ppath, pname string, gran string, mpath string) string {
	incfg := filepath.Join(outdir, tag+"incfg.txt")
	outcfg := filepath.Join(outdir, "outcfg.txt")
	p := coverage.CoverPkgConfig{
		PkgPath:      ppath,
		PkgName:      pname,
		Granularity:  gran,
		OutConfig:    outcfg,
		EmitMetaFile: mpath,
	}
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}
	writeFile(t, incfg, data)
	return incfg
}

func writeOutFileList(t *testing.T, infiles []string, outdir, tag string) ([]string, string) {
	outfilelist := filepath.Join(outdir, tag+"outfilelist.txt")
	var sb strings.Builder
	outfs := []string{}
	for _, inf := range infiles {
		base := filepath.Base(inf)
		of := filepath.Join(outdir, tag+".cov."+base)
		outfs = append(outfs, of)
		fmt.Fprintf(&sb, "%s\n", of)
	}
	if err := os.WriteFile(outfilelist, []byte(sb.String()), 0666); err != nil {
		t.Fatalf("writing %s: %v", outfilelist, err)
	}
	return outfs, outfilelist
}

func runPkgCover(t *testing.T, outdir string, tag string, incfg string, mode string, infiles []string, errExpected bool) ([]string, string, string) {
	// Write the pkgcfg file.
	outcfg := filepath.Join(outdir, "outcfg.txt")

	// Form up the arguments and run the tool.
	outfiles, outfilelist := writeOutFileList(t, infiles, outdir, tag)
	args := []string{"-pkgcfg", incfg, "-mode=" + mode, "-var=var" + tag, "-outfilelist", outfilelist}
	args = append(args, infiles...)
	cmd := testenv.Command(t, testcover(t), args...)
	if errExpected {
		errmsg := runExpectingError(cmd, t)
		return nil, "", errmsg
	} else {
		run(cmd, t)
		return outfiles, outcfg, ""
	}
}

func TestCoverWithCfg(t *testing.T) {
	testenv.MustHaveGoRun(t)

	t.Parallel()

	// Subdir in testdata that has our input files of interest.
	tpath := filepath.Join("testdata", "pkgcfg")
	dir := tempDir(t)
	instdira := filepath.Join(dir, "insta")
	if err := os.Mkdir(instdira, 0777); err != nil {
		t.Fatal(err)
	}

	scenarios := []struct {
		mode, gran string
	}{
		{
			mode: "count",
			gran: "perblock",
		},
		{
			mode: "set",
			gran: "perfunc",
		},
		{
			mode: "regonly",
			gran: "perblock",
		},
	}

	var incfg string
	apkgfiles := []string{filepath.Join(tpath, "a", "a.go")}
	for _, scenario := range scenarios {
		// Instrument package "a", producing a set of instrumented output
		// files and an 'output config' file to pass on to the compiler.
		ppath := "cfg/a"
		pname := "a"
		mode := scenario.mode
		gran := scenario.gran
		tag := mode + "_" + gran
		incfg = writePkgConfig(t, instdira, tag, ppath, pname, gran, "")
		ofs, outcfg, _ := runPkgCover(t, instdira, tag, incfg, mode,
			apkgfiles, false)
		t.Logf("outfiles: %+v\n", ofs)

		// Run the compiler on the files to make sure the result is
		// buildable.
		bargs := []string{"tool", "compile", "-p", "a", "-coveragecfg", outcfg}
		bargs = append(bargs, ofs...)
		cmd := testenv.Command(t, testenv.GoToolPath(t), bargs...)
		cmd.Dir = instdira
		run(cmd, t)
	}

	// Do some error testing to ensure that various bad options and
	// combinations are properly rejected.

	// Expect error if config file inaccessible/unreadable.
	mode := "atomic"
	errExpected := true
	tag := "errors"
	_, _, errmsg := runPkgCover(t, instdira, tag, "/not/a/file", mode,
		apkgfiles, errExpected)
	want := "error reading pkgconfig file"
	if !strings.Contains(errmsg, want) {
		t.Errorf("'bad config file' test: wanted %s got %s", want, errmsg)
	}

	// Expect err if config file contains unknown stuff.
	t.Logf("mangling in config")
	writeFile(t, incfg, []byte("blah=foo\n"))
	_, _, errmsg = runPkgCover(t, instdira, tag, incfg, mode,
		apkgfiles, errExpected)
	want = "error reading pkgconfig file"
	if !strings.Contains(errmsg, want) {
		t.Errorf("'bad config file' test: wanted %s got %s", want, errmsg)
	}

	// Expect error on empty config file.
	t.Logf("writing empty config")
	writeFile(t, incfg, []byte("\n"))
	_, _, errmsg = runPkgCover(t, instdira, tag, incfg, mode,
		apkgfiles, errExpected)
	if !strings.Contains(errmsg, want) {
		t.Errorf("'bad config file' test: wanted %s got %s", want, errmsg)
	}
}

func TestCoverOnPackageWithNoTestFiles(t *testing.T) {
	testenv.MustHaveGoRun(t)
	t.Parallel()
	dir := tempDir(t)

	// For packages with no test files, the new "go test -cover"
	// strategy is to run cmd/cover on the package in a special
	// "EmitMetaFile" mode. When running in this mode, cmd/cover walks
	// the package doing instrumention, but when finished, instead of
	// writing out instrumented source files, it directly emits a
	// meta-data file for the package in question, essentially
	// simulating the effect that you would get if you added a dummy
	// "no-op" x_test.go file and then did a build and run of the test.

	// Case 1: go files with functions but no test files.
	tpath := filepath.Join("testdata", "pkgcfg")
	pkg1files := []string{filepath.Join(tpath, "yesFuncsNoTests", "yfnt.go")}
	ppath := "cfg/yesFuncsNoTests"
	pname := "yesFuncsNoTests"
	mode := "count"
	gran := "perblock"
	tag := mode + "_" + gran
	instdir := filepath.Join(dir, "inst")
	if err := os.Mkdir(instdir, 0777); err != nil {
		t.Fatal(err)
	}
	mdir := filepath.Join(dir, "meta")
	if err := os.Mkdir(mdir, 0777); err != nil {
		t.Fatal(err)
	}
	mpath := filepath.Join(mdir, "covmeta.xxx")
	incfg := writePkgConfig(t, instdir, tag, ppath, pname, gran, mpath)
	_, _, errmsg := runPkgCover(t, instdir, tag, incfg, mode,
		pkg1files, false)
	if errmsg != "" {
		t.Fatalf("runPkgCover err: %q", errmsg)
	}

	// Check for existence of meta-data file.
	if inf, err := os.Open(mpath); err != nil {
		t.Fatalf("meta-data file not created: %v", err)
	} else {
		inf.Close()
	}

	// Make sure it is digestible.
	cdargs := []string{"tool", "covdata", "percent", "-i", mdir}
	cmd := testenv.Command(t, testenv.GoToolPath(t), cdargs...)
	run(cmd, t)

	// Case 1: go file with no functions and no tests.
	pkg2files := []string{filepath.Join(tpath, "noFuncsNoTests", "nfnt.go")}
	pname2 := "noFuncsNoTests"
	ppath2 := "cfg/" + pname
	tag2 := mode + "_" + gran
	instdir2 := filepath.Join(dir, "inst2")
	if err := os.Mkdir(instdir2, 0777); err != nil {
		t.Fatal(err)
	}
	mdir2 := filepath.Join(dir, "meta2")
	if err := os.Mkdir(mdir2, 0777); err != nil {
		t.Fatal(err)
	}
	mpath2 := filepath.Join(mdir2, "covmeta.yyy")
	incfg2 := writePkgConfig(t, instdir2, tag2, ppath2, pname2, gran, mpath2)
	_, _, errmsg2 := runPkgCover(t, instdir2, tag2, incfg2, mode,
		pkg2files, false)
	if errmsg2 != "" {
		t.Fatalf("runPkgCover err: %q", errmsg2)
	}

	// This time around we don't expect to see the meta-data file.
	if inf, err := os.Open(mpath2); err == nil {
		t.Fatalf("meta-data file was created: %v", err)
		inf.Close()
	}
}
