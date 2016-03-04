// +build !nacl
// run

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"log"
	"os/exec"
)

func main() {
	run("go", "tool", "link", "-B", "")     // -B argument must start with 0x:
	run("go", "tool", "link", "-B", "0")    // -B argument must start with 0x: 0
	run("go", "tool", "link", "-B", "0x")   // usage: link [options] main.o
	run("go", "tool", "link", "-B", "0x0")  // -B argument must have even number of digits: 0x0
	run("go", "tool", "link", "-B", "0x00") // usage: link [options] main.o
}

func run(name string, args ...string) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	if err == nil {
		log.Fatalf("expected cmd/link to fail")
	}

	if bytes.HasPrefix(out, []byte("panic")) {
		log.Fatalf("cmd/link panicked:\n%s", out)
	}
}
