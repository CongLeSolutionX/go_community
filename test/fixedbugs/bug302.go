// run

// Copyright 2010 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

var thechar = map[string]string{
	"arm": "5",
	"amd64": "6",
	"amd64p32": "6",
	"386": "8",
	"power64": "9",
	"power64le": "9",
}

func main() {
	a := thechar[runtime.GOARCH]
	if a == "" {
		fmt.Println("BUG: unknown GOARCH")
		os.Exit(1)
	}

	run("go", "tool", a+"g", "fixedbugs/bug302.dir/p.go")
	run("go", "tool", "pack", "grc", "pp.a", "p."+a)
	run("go", "tool", a+"g", "fixedbugs/bug302.dir/main.go")
	os.Remove("p."+a)
	os.Remove("pp.a")
	os.Remove("main."+a)
}

func run(cmd string, args ...string) {
	out, err := exec.Command(cmd, args...).CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		fmt.Println(err)
		os.Exit(1)
	}
}
