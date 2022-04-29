// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ignore

package main

import (
	"log"
	"os"
	"os/exec"
	"strings"
)

const beginning = `// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package goroot

import (
	"strings"
)

var std = strings.Fields(` + "`\n"

const end = "\n`)\n"

func getStdpkgs() []string {
	allPkgs := make(map[string]bool)
	for _, t := range []string{"linux", "wasm,js", "windows"} { // covering set?
		stdpkgs, err := exec.Command("go", "list", "-tags", t, "std", "cmd", "builtin").Output()
		if err != nil {
			log.Fatal(err)
		}
		for _, v := range strings.Fields(string(stdpkgs)) {
			allPkgs[v] = true
		}
	}
	var l []string
	for v := range allPkgs {
		l = append(l, v)
	}
	return l
}

func main() {
	s := beginning + strings.Join(getStdpkgs(), "\n") + end

	os.WriteFile("stdpkgs.go", []byte(s), 0644)
}
