// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package testenv

import (
	"fmt"
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
