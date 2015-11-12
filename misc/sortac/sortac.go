// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Sortac sorts the AUTHORS and CONTRIBUTORS files.

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

func main() {
	flag.Parse()
	var rc io.ReadCloser
	switch flag.NArg() {
	case 0:
		rc = os.Stdin
	case 1:
		var err error
		rc, err = os.Open(flag.Arg(0))
		if err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatal("Usage: sortac [file]")
	}
	bs := bufio.NewScanner(rc)
	var header []string
	var lines []string
	for bs.Scan() {
		t := bs.Text()
		lines = append(lines, t)
		if t == "# Please keep the list sorted." {
			header = lines
			lines = nil
			continue
		}
	}
	if err := bs.Err(); err != nil {
		log.Fatal(err)
	}
	rc.Close()

	var wc io.WriteCloser = os.Stdout
	if flag.NArg() == 1 {
		var err error
		wc, err = os.Create(flag.Arg(0))
		if err != nil {
			log.Fatal(err)
		}

	}
	c := collate.New(language.Und, collate.Loose)
	c.SortStrings(lines)
	for _, l := range header {
		fmt.Fprintln(wc, l)
	}
	for _, l := range lines {
		fmt.Fprintln(wc, l)
	}
	if err := wc.Close(); err != nil {
		log.Fatal(err)
	}
}
