// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// expdump prints the contents of the export data section.
//
// Usage:
//
//	expdump binary

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

var dump = flag.Bool("d", false, "dump export data")

func usage() {
	fmt.Fprintf(os.Stderr, "usage: expdump [flags] binary\n\n")
	flag.PrintDefaults()
	os.Exit(2)
}

type reader struct {
	r    *bufio.Reader
	read int
	dump bool
}

func (r *reader) next() int {
	b, err := r.r.ReadByte()
	if err == io.EOF {
		return -1
	}
	if err != nil {
		log.Fatal(err)
	}
	r.read++
	if r.dump {
		fmt.Printf("%c", b)
	}
	return int(b)
}

func (r *reader) seek(s string) {
	c := r.next()
	for c >= 0 {
		for i := 0; i < len(s) && c == int(s[i]); i++ {
			if i+1 == len(s) {
				return // match
			}
			c = r.next()
		}
		c = r.next()
	}
	log.Fatal("no start of export data found")
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("expdump: ")

	flag.Usage = usage
	flag.Parse()
	if flag.NArg() != 1 {
		usage()
	}

	f, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// find start of export section
	r := &reader{r: bufio.NewReader(f)}
	r.seek("$$")
	start := r.read
	r.dump = *dump

	// find end of export section
	r.seek("$$")
	if r.dump {
		fmt.Println()
	}

	fmt.Printf("%10d\t%s\n", r.read-start, flag.Arg(0))
}
