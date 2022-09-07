// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package testenv

import (
	"fmt"
	"internal/goexperiment"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const toolExecTemplate = `
package main

import (
	"os"
	"os/exec"
	"strings"
)

func main() {
	if strings.HasSuffix(strings.TrimSuffix(os.Args[2], ".exe"), "REPLACEME") {
		os.Args[2] = os.Args[1]
	}
	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		os.Exit(1)
	}
}
`

// BuildToolExec builds a "toolexec" wrapper suitable for use in
// testing tools that are invoked during "go build" or "go test"
// invocations, such as "compile", "link", etc. This helper also
// allows for the possibility of specializing the unit test executable
// itself to act as a stand-in for the tool (as is done with the
// cmd/go tests). Tool name is specified in 'cmd' (where cmd may be of
// the form <tool> or <tool.test>), and the wrapper executable is
// written to the directory 'dir'.
func BuildToolExec(t *testing.T, cmd string, dir string) (string, error) {
	if !HasGoBuild() {
		return "", fmt.Errorf("BuildToolExec: no 'go build' support")
	}

	// Double-check the command just to make sure it is on the short
	// list. Also allow for the scenario of using the tool unit test
	// executable in place of the tool itself.
	allowed := []string{"link", "compile", "asm", "cover", "buildid",
		"cgo", "pack", "vet"}
	found := false
	for i := range allowed {
		if cmd == allowed[i] || cmd == allowed[i]+".test" {
			found = true
			break
		}
	}
	if !found {
		return "", fmt.Errorf("BuildToolExec: unexpected cmd %q", cmd)
	}

	// Write a modified version of the source above into
	// a temp directory.
	contents := strings.Replace(toolExecTemplate, "REPLACEME", cmd, 1)
	srcpath := filepath.Join(dir, "toolexec.go")
	if err := os.WriteFile(srcpath, []byte(contents), 0666); err != nil {
		t.Fatalf("os.WriteFile(%s) failed: %v", srcpath, err)
	}

	// Build tool, return a path to binary.
	exepath := filepath.Join(dir, "toolexec.exe")
	out, err := exec.Command(GoToolPath(t), "build", "-o", exepath, srcpath).CombinedOutput()
	if len(out) > 0 {
		t.Logf("%s", out)
	}
	return exepath, err
}

// AugmentToolBuildForCoverage accepts a list of arguments for "go
// build" and augments them (if appropriate) with options to enable
// code coverage. This helper is intended to be used by tests running
// in the Go "cmd" source tree where the test builds a copy of itself
// to run tests with, as opposed to using the tool installed in
// $GOROOT/bin. Here 'gobuildargs' are the arguments that will be
// passed to "go" when doing the tool build, and 'ppath' is a package
// pattern selecting the tool itself. What things might look like for
// in the test code for "cmd/cover", which builds a copy of itself to
// test:
//
//	args := []string{"-o", toolpath, "cmd/cover"}
//	args = AugmentToolBuildForCoverage(args, "cmd/cover")
//	out, err := exec.Command(testenv.GoToolPath(t), args...).CombinedOutput()
//	...
//
// AugmentToolBuildForCoverage asks the testing package whether
// coverage is enabled, and if so, adds coverage testing options to
// the build for the tool.
func AugmentToolBuildForCoverage(gobuildargs []string, ppath string) []string {
	// First argument expected to be "build"
	if len(gobuildargs) < 1 || gobuildargs[0] != "build" {
		panic(fmt.Sprintf("invalid go build args passed to testenv.AugmentToolBuildForCoverage: %+v", gobuildargs))
	}
	// Funtionality requires redesigned coverage.
	if !goexperiment.CoverageRedesign {
		return gobuildargs
	}
	// No need to do anything if "go test -cover" is not in effect.
	if testing.CoverMode() == "" {
		return gobuildargs
	}
	// Return augmented args list.
	return append([]string{"build", "-cover", "-covermode", testing.CoverMode(), "-coverpkg", ppath}, gobuildargs[1:]...)
}
