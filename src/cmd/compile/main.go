// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"cmd/internal/obj"
	"fmt"
	"log"
	"os"
)

var archMain = map[string]func(){}

func main() {
	// disable timestamps for reproducible output
	log.SetFlags(0)
	log.SetPrefix("compile: ")

	if mainFn, ok := archMain[obj.GOARCH]; ok {
		mainFn()
	} else {
		fmt.Fprintf(os.Stderr, "compile: unknown architecture %q\n", obj.GOARCH)
		os.Exit(2)
	}

	/*
		switch obj.GOARCH {
		default:
			fmt.Fprintf(os.Stderr, "compile: unknown architecture %q\n", obj.GOARCH)
			os.Exit(2)
		case "386":
			x86.Main()
		case "amd64", "amd64p32":
			amd64.Main()
		case "arm":
			arm.Main()
		case "arm64":
			arm64.Main()
		case "mips64", "mips64le":
			mips64.Main()
		case "ppc64", "ppc64le":
			ppc64.Main()
		case "s390x":
			s390x.Main()
		}
	*/
}
