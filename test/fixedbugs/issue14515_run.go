// run

// Copyright 2016 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Make sure -S generates assembly code.

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func main() {
	if runtime.Compiler != "gc" {
		return
	}
	out, err := exec.Command("go", "tool", "compile", "-S", filepath.Join("fixedbugs", "issue14515.go")).CombinedOutput()
	os.Remove("issue14515.o")
	if err != nil {
		panic(err)
	}

	patterns := []string{
		// It is hard to look for actual instructions in an
		// arch-independent way.  So we'll just look for
		// pseudo-ops that are arch-independent.
		"\tTEXT\t",
		"\tFUNCDATA\t",
		"\tPCDATA\t",
	}
	outstr := string(out)
	for _, p := range patterns {
		if !strings.Contains(outstr, p) {
			println(outstr)
			panic("can't find pattern " + p)
		}
	}
}
