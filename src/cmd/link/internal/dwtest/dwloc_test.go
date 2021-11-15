// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dwtest_test

// This file contains a set of DWARF variable location generation
// tests that are intended to compliment the existing linker DWARF
// tests. The tests make use of a harness / utility program
// "dwdumploc" that is built during test setup (TestMain) and then
// invoked (fork+exec) in testpoints. We do things this way (as
// opposed to just incorporating all of the source code from
// testdata/dwdumploc.go into this file) so that the dumper code can
// import packages from Delve without needing to vendor everything
// into the Go distribution itself.

import (
	"flag"
	"fmt"
	"internal/testenv"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var verbflag = flag.Int("v", 0, "verbosity level")
var keepflag = flag.Bool("keep", false, "preserve work dir")

var harnessPath string
var workDir string
var goToolPath string

func verb(vlevel int, s string, a ...interface{}) {
	if *verbflag >= vlevel {
		fmt.Printf(s, a...)
		fmt.Printf("\n")
	}
}

// copyFilesForHarness copies various files into the build dir for the
// harness, including the main package, go.mod, and a copy of the
// dwtest package.
func copyFilesForHarness() (bd string, cd string, err error) {
	mkdir := func(d string) bool {
		err = os.Mkdir(d, 0777)
		return err == nil
	}
	cp := func(from, to string) bool {
		var payload []byte
		payload, err = ioutil.ReadFile(from)
		if err != nil {
			return false
		}
		err = ioutil.WriteFile(to, payload, 0644)
		if err != nil {
			return false
		}
		return true
	}
	join := filepath.Join
	bd = join(workDir, "build")
	cd = join(workDir, "cache")
	bdt := join(bd, "dwtest")
	if !mkdir(bd) || !mkdir(bdt) || !mkdir(cd) ||
		!cp(join("testdata", "dwdumploc.go"), join(bd, "main.go")) ||
		!cp(join("testdata", "go.mod"), join(bd, "go.mod")) ||
		!cp(join("testdata", "go.sum"), join(bd, "go.sum")) ||
		!cp("dwtest.go", join(bdt, "dwtest.go")) {
		return
	}
	return
}

// replaceEnv returns a new environment derived from env
// by removing any existing definition of ev and adding ev=evv.
func replaceEnv(env []string, ev string, evv string) []string {
	if env == nil {
		env = os.Environ()
	}
	evplus := ev + "="
	var found bool
	for i, v := range env {
		if strings.HasPrefix(v, evplus) {
			found = true
			env[i] = evplus + evv
		}
	}
	if !found {
		env = append(env, evplus+evv)
	}
	return env
}

// buildHarness builds the helper program "dwdumploc.exe".
func buildHarness() (err error) {
	goToolPath, err = testenv.GoTool()
	if err != nil {
		return
	}
	harnessPath = filepath.Join(workDir, "dumpdwloc.exe")

	// Copy source files into build dir.
	var cd, bd string
	bd, cd, err = copyFilesForHarness()
	if err != nil {
		return
	}

	// Run build.
	cmd := exec.Command(goToolPath, "build", "-o", harnessPath)
	cmd.Dir = bd
	cmd.Env = replaceEnv(cmd.Env, "GOCACHE", cd)
	if b, err2 := cmd.CombinedOutput(); err2 != nil {
		err = fmt.Errorf("build failed (%v): %s", err2, b)
		return
	}
	return
}

// testMain contains the guts of the test setup code, mainly
// building the harness executable.
func testMain(m *testing.M) (int, error) {
	var err error
	workDir, err = os.MkdirTemp("", "dwloctest")
	if err != nil {
		return 0, err
	}
	verb(1, "workdir is: %s", workDir)
	if !*keepflag {
		defer os.RemoveAll(workDir)
	}
	if err := buildHarness(); err != nil {
		return 1, err
	}
	return m.Run(), nil
}

func TestMain(m *testing.M) {
	if !testenv.HasGoBuild() {
		os.Exit(0)
	}
	flag.Parse()
	exitCode, err := testMain(m)
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(exitCode)
}

// runHarness runs our previously built harness exec on a Go binary
// 'exePath' for function 'fcn' and returns the results.
func runHarness(t *testing.T, exePath string, fcn string) string {
	t.Helper()
	cmd := exec.Command(harnessPath, "-m", exePath, "-f", fcn)
	b, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("running 'harness -m %s -f %s': %v", exePath, fcn, err)
		return ""
	}
	return strings.TrimSpace(string(b))
}

func gobuild(t *testing.T, tag, sourceCode string) string {
	t.Helper()
	spath := filepath.Join(workDir, tag+".go")
	err := ioutil.WriteFile(spath, []byte(sourceCode), 0644)
	if err != nil {
		t.Fatalf("write to %s failed: %s", spath, err)
		return ""
	}
	epath := filepath.Join(workDir, tag+".exe")
	cmd := exec.Command(goToolPath, "build", "-o", epath, spath)
	b, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("%% build output: %s\n", b)
		t.Fatalf("build of %s failed: %s", tag, err)
	}
	return epath
}

const programSourceCode = `
package main

var G int

//go:noinline
func another(x int) {
	println(G)
}

//go:noinline
func docall(f func()) {
	f()
}

//go:noinline
func Issue47354(s string) {
	docall(func() {
		println("s is", s)
	})
	G++
	another(int(s[0]))
}

func main() {
	Issue47354("poo")
}

`

func TestDwarfVariableLocations(t *testing.T) {
	switch runtime.GOOS {
	case "linux", "windows", "darwin":
	default:
		t.Skipf("unsupported OS %s", runtime.GOOS)
	}
	switch runtime.GOARCH {
	case "amd64", "arm64":
	default:
		t.Skipf("unsupported ARCH %s", runtime.GOARCH)
	}

	type stringmap map[string]string

	testcases := map[string]stringmap{
		"Issue47354": stringmap{
			"amd64": "1: in-param \"s\" loc=\"{ [0: S=8 RAX] [1: S=8 RBX] }\"",
			"arm64": "1: in-param \"s\" loc=\"{ [0: S=8 R0] [1: S=8 R1] }\"",
		},
	}

	// Build
	ppath := gobuild(t, "TestDwarfVariableLocations", programSourceCode)

	// Examine.
	for fname, expectedMap := range testcases {
		// Run harness.
		got := runHarness(t, ppath, "main."+fname)
		want := expectedMap[runtime.GOARCH]
		if got != want {
			t.Errorf("TestDwarfVariableLocations: failed on Issue47354 testcase arch %s:\ngot: %q\nwant: %q", runtime.GOARCH, got, want)
		}
	}
}
