// run
//go:build !js && !wasip1

// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const src = `package main

import "runtime"

var x = [1024]byte{}

var ch = make(chan bool)

func main() {
	go func() {
		runtime.Gosched()
		var y = [len(x)]byte{}
		eq := x == y
		ch <- eq
	}()
	runtime.Gosched()
	for k := range x {
		x[k]++
	}
	println(<-ch)
}
`

func main() {
	dir, err := os.MkdirTemp("", "issue61204")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	filename := filepath.Join(dir, "main.go")
	if err := os.WriteFile(filename, []byte(src), 0644); err != nil {
		panic(err)
	}

	cmd := exec.Command("go", "run", "-race", "main.go")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err == nil {
		panic("program run succeeded unexpectedly")
	}
	if bytes.Contains(output, []byte("-race is not supported ")) {
		return
	}
	if !bytes.Contains(output, []byte("WARNING: DATA RACE")) {
		panic(fmt.Sprintf("missing data race report in output; got:\n\n%s", output))
	}
}
