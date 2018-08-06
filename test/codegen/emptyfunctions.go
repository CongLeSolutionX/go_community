// asmcheck

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package codegen

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"
)

// This file contains codegen tests related to empty functions
// compile level optimizations.

// TestDeferOptimization makes sure defer calls to empty function are not in the assembly
func TestDeferOptimization(t *testing.T) {
	// Make a directory to work in.
	dir, err := ioutil.TempDir("", "issue26534-")
	if err != nil {
		log.Fatalf("could not create directory: %v", err)
	}
	defer os.RemoveAll(dir)

	// Create source.
	src := filepath.Join(dir, "test.go")
	f, err := os.Create(src)
	if err != nil {
		log.Fatalf("could not create source file: %v", err)
	}
	f.Write([]byte(`
package main
func assert(cond bool) { /* NOOP in this build tag. */ }
var x = 0
func main() {
	assert(x == 0)
	func() {}()
	func() { assert(x == 1) }()
	defer assert(x == 0)
	defer func() {}()
	defer func() { assert(x == 1) }()
}	
`))
	f.Close()
	cmd := exec.Command("go", "build", "-gcflags", "-S", "-o", filepath.Join(dir, "test"), src)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("could not build target: %v", err)
	}
	re, _ := regexp.Compile(`(?s)"".main .*?go:([8-9]|1\d).*?"".main.func1 STEXT`)
	if re.Match(out) {
		println(string(out))
		panic("Deferred calls to empty functions")
	}
}

// TestGoroutineOptimization makes sure goroutine calls to empty function are not in the assembly
func TestGoroutineOptimization(t *testing.T) {
	// Make a directory to work in.
	dir, err := ioutil.TempDir("", "issue26534-")
	if err != nil {
		log.Fatalf("could not create directory: %v", err)
	}
	defer os.RemoveAll(dir)

	// Create source.
	src := filepath.Join(dir, "test.go")
	f, err := os.Create(src)
	if err != nil {
		log.Fatalf("could not create source file: %v", err)
	}
	f.Write([]byte(`
package main
func assert(cond bool) { /* NOOP in this build tag. */ }
var x = 0
func main() {
	assert(x == 0)
	func() {}()
	func() { assert(x == 1) }()
	go assert(x == 0)
	go func() {}()
	go func() { assert(x == 1) }()
}	
`))
	f.Close()
	cmd := exec.Command("go", "build", "-gcflags", "-S", "-o", filepath.Join(dir, "test"), src)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("could not build target: %v", err)
	}
	re, _ := regexp.Compile(`(?s)"".main .*?go:([8-9]|1\d).*?"".main.func1 STEXT`)
	if re.Match(out) {
		println(string(out))
		panic("Go calls to empty functions")
	}
}
