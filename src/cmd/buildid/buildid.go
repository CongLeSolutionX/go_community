// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"cmd/internal/buildid"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: go tool buildid [-w] file\n")
	flag.PrintDefaults()
	os.Exit(2)
}

var wflag = flag.Bool("w", false, "write build ID")

func main() {
	log.SetPrefix("buildid: ")
	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() != 1 {
		usage()
	}

	file := flag.Arg(0)
	id, err := buildid.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	if !*wflag {
		fmt.Printf("%s\n", id)
		return
	}

	err, _ = buildid.BuildAndFindHash(id, file, "/", file, true)

	if err == buildid.ErrEmptyMatchLen {
		return
	}

	if err != nil {
		log.Fatal(err)
	}
}
