// +build !nacl,!js

// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 31645: Ensure that multi-byte characters are properly
// encoded for WASM in command-line arguments.
//
// It runs a cross-compiled test for GOOS=js GOARCH=wasm and ironically
// doesn't build on js/wasm since exec has issues on there.
// However, it'll require node.js to be available in the user's path
// in order for the test to be run.

package multibyte_argv_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestMultibyteArgv(t *testing.T) {
	_, err := exec.LookPath("node")
	if err != nil {
		// This shouldn't be an error since node might just not be on the system.
		t.Logf(`"node" not defined in path`)
		return
	}
	goRoot, err := findGoRoot()
	if err != nil {
		t.Fatalf("Failed to find GOROOT: %v", err)
	}
	tmpDir, err := ioutil.TempDir("", "issue31645")
	if err != nil {
		t.Fatalf("Failed to create tempdir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	// Cross compile the main.go code to main_wasm
	const src = `
package main
import "os"
import "fmt"
func main() {
    fmt.Println(os.Args)
}`
	srcFile := filepath.Join(tmpDir, "main.go")
	if err := ioutil.WriteFile(srcFile, []byte(src), 0600); err != nil {
		t.Fatalf("Failed to create main.go file: %v", err)
	}
	mainWasmFullPath := filepath.Join(tmpDir, "main_wasm")
	cmd := exec.Command("go", "build", "-o", mainWasmFullPath, srcFile)
	cmd.Env = append(cmd.Env, "GOOS=js", "GOARCH=wasm", "GOCACHE="+tmpDir, "GOPATH="+goRoot)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run cross compilation, error: %v\nOutput: \n%s", err, output)
	}
	// After cross compilation the next step is to run:
	//     node misc/wasm/wasm_exec.js main_wasm hello 世界
	miscWasmExecFile := filepath.Join(goRoot, "misc", "wasm", "wasm_exec.js")
	nodeCmd := exec.Command("node", miscWasmExecFile, mainWasmFullPath, "hello", "世界")
	output, err = nodeCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run node command %s: %v\nOutput: \n%s", nodeCmd, err, output)
	}
	want := "[" + mainWasmFullPath + " hello 世界]\n"
	if got := string(output); got != want {
		t.Fatalf("Output mismatch\nGot:  %q\nWant: %q", got, want)
	}
}
func findGoRoot() (string, error) {
	const src = `
package main
import "runtime"
func main() { println(runtime.GOROOT()) }`
	tmpDir, err := ioutil.TempDir(os.TempDir(), "findgoroot")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmpDir)
	mainPath := filepath.Join(tmpDir, "main.go")
	if err := ioutil.WriteFile(mainPath, []byte(src), 0600); err != nil {
		return "", err
	}
	output, err := exec.Command("go", "run", mainPath).CombinedOutput()
	if err != nil {
		return "", err
	}
	goroot := strings.TrimSpace(string(output))
	return goroot, nil
}
