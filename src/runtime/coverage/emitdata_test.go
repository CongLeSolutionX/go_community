// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package coverage

import (
	"fmt"
	"internal/coverage"
	"internal/goexperiment"
	"internal/testenv"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Set to true for debugging.
const fixedTestDir = false

func TestCoverageApis(t *testing.T) {
	if !goexperiment.CoverageRedesign {
		t.Skipf("skipping new coverage tests (experiment not enabled)")
	}
	testenv.MustHaveGoBuild(t)
	dir := t.TempDir()
	if fixedTestDir {
		dir = "/tmp/zzz"
		os.RemoveAll(dir)
		mkdir(t, dir)
	}

	// Build harness.
	bdir := mkdir(t, filepath.Join(dir, "build"))
	rdir1 := mkdir(t, filepath.Join(dir, "runDir1"))
	rdir2 := mkdir(t, filepath.Join(dir, "runDir2"))
	edir := mkdir(t, filepath.Join(dir, "emitDir"))
	wdir := mkdir(t, filepath.Join(dir, "writerDir"))
	harnessPath := buildHarness(t, bdir, []string{"-cover"})

	t.Logf("harness path is %s", harnessPath)

	// Sub-tests for each API we want to inspect, plus
	// extras for error testing.
	t.Run("emitToDir", func(t *testing.T) {
		t.Parallel()
		testEmitToDir(t, harnessPath, rdir1, edir)
	})
	t.Run("emitToWriter", func(t *testing.T) {
		t.Parallel()
		testEmitToWriter(t, harnessPath, rdir2, wdir)
	})
}

// buildHarness builds the helper program "dwdumploc.exe".
func buildHarness(t *testing.T, dir string, opts []string) string {
	harnessPath := filepath.Join(dir, "harness.exe")
	harnessSrc := filepath.Join("testdata", "harness.go")
	args := []string{"build", "-gcflags=all=-l -N", "-o", harnessPath}
	args = append(args, opts...)
	args = append(args, harnessSrc)
	cmd := exec.Command(testenv.GoToolPath(t), args...)
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed (%v): %s", err, b)
	}
	return harnessPath
}

func mkdir(t *testing.T, d string) string {
	if err := os.Mkdir(d, 0777); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	return d
}

func TestApisOnNocoverBinary(t *testing.T) {
	testenv.MustHaveGoBuild(t)
	dir := t.TempDir()

	// Build harness with no -cover.
	bdir := mkdir(t, filepath.Join(dir, "nocover"))
	edir := mkdir(t, filepath.Join(dir, "emitDirNo"))
	harnessPath := buildHarness(t, bdir, nil)
	output, err := runHarness(t, harnessPath, "emitToDir", edir, edir)
	if err == nil {
		t.Fatalf("expected error on TestApisOnNocoverBinary harness run")
	}
	const want = "not built with -cover"
	if !strings.Contains(output, want) {
		t.Errorf("error output does not contain %q: %s", want, output)
	}
}

func addGoCoverDir(env []string, gcd string) []string {
	rv := []string{}
	found := false
	for _, v := range env {
		if strings.HasPrefix(v, "GOCOVERDIR=") {
			v = "GOCOVERDIR=" + gcd
			found = true
		}
		rv = append(rv, v)
	}
	if !found {
		rv = append(rv, "GOCOVERDIR="+gcd)
	}
	return rv
}

func runHarness(t *testing.T, harnessPath string, tp string, rdir, edir string) (string, error) {
	t.Logf("running: %s -tp %s -o %s with GOCOVERDIR=%s", harnessPath, tp, edir, rdir)
	cmd := exec.Command(harnessPath, "-tp", tp, "-o", edir)
	cmd.Dir = rdir
	cmd.Env = addGoCoverDir(os.Environ(), rdir)
	b, err := cmd.CombinedOutput()
	t.Logf("harness run output: %s\n", string(b))
	return string(b), err
}

func testEmitToDir(t *testing.T, harnessPath string, rdir, edir string) {
	output, err := runHarness(t, harnessPath, "emitToDir", rdir, edir)
	if err != nil {
		t.Logf("%s", output)
		t.Fatalf("running 'harness -tp emitDir': %v", err)
	}

	// Just check to make sure meta-data file and counter data file were
	// written. Another alternative would be to run "go tool covdata"
	// or equivalent, but for now, this is what we've got.
	dents, err := os.ReadDir(edir)
	if err != nil {
		t.Fatalf("os.ReadDir(%s) failed: %v", edir, err)
	}
	mfc := 0
	cdc := 0
	for _, e := range dents {
		if e.IsDir() {
			continue
		}
		if strings.HasPrefix(e.Name(), coverage.MetaFilePref) {
			mfc++
		} else if strings.HasPrefix(e.Name(), coverage.CounterFilePref) {
			cdc++
		}
	}
	wantmf := 1
	wantcf := 1
	if mfc != wantmf {
		t.Errorf("EmitToDir: want %d meta-data files, got %d\n", wantmf, mfc)
	}
	if cdc != wantcf {
		t.Errorf("EmitToDir: want %d counter-data files, got %d\n", wantcf, cdc)
	}
}

func testForSpecificFunctions(t *testing.T, dir string, want []string, avoid []string) string {
	args := []string{"tool", "covdata", "dump",
		"-live", "-emitdump", "-pkg=^main$", "-i=" + dir}
	t.Logf("running: go %v\n", args)
	cmd := exec.Command(testenv.GoToolPath(t), args...)
	b, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("'go tool covdata failed (%v): %s", err, b)
	}
	output := string(b)
	rval := ""
	for _, f := range want {
		wf := "Func: " + f
		if strings.Contains(output, wf) {
			continue
		}
		rval += fmt.Sprintf("error: output should contain %q but does not\n", wf)
	}
	for _, f := range avoid {
		wf := "Func: " + f
		if strings.Contains(output, wf) {
			rval += fmt.Sprintf("error: output should not contain %q but does\n", wf)
		}
	}
	return rval
}

func testEmitToWriter(t *testing.T, harnessPath string, rdir, edir string) {
	tp := "emitToWriter"
	output, err := runHarness(t, harnessPath, tp, rdir, edir)
	if err != nil {
		t.Logf("%s", output)
		t.Fatalf("running 'harness -tp %s': %v", tp, err)
	}
	want := []string{"main", tp}
	avoid := []string{"final"}
	if msg := testForSpecificFunctions(t, edir, want, avoid); msg != "" {
		t.Errorf("coverage data from %q output match failed: %s", tp, msg)
	}
}
