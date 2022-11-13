// run
// +build !nacl,!js,gc

// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Run the sinit test.

package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	out, err := exec.Command("go", "build", "-gcflags=-S", "sinit.go").CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		fmt.Println(err)
		os.Exit(1)
	}

	if len(bytes.TrimSpace(out)) == 0 {
		fmt.Println("'go tool compile -S sinit.go' printed no output")
		os.Exit(1)
	}
	init := false
	ok := true
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, ".init STEXT ") {
			init = true
			continue
		}
		if strings.Contains(line, " STEXT ") {
			init = false
		}
		if strings.Contains(line, "FUNCDATA") || strings.Contains(line, "RET") || strings.Contains(line, "\tTEXT\t") || !strings.Contains(line, "sinit.go:") {
			continue
		}
		if init {
			if ok {
				fmt.Printf("sinit has init code:\n")
				ok = false
			}
			fmt.Printf("\t%s\n", line)
		}
	}
	if !ok {
		os.Exit(1)
	}
}
