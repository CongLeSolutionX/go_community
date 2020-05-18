// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Objdump disassembles executable files.
//
// Usage:
//
//	go tool objdump [-s symregexp] binary
//
// Objdump prints a disassembly of all text symbols (code) in the binary.
// If the -s option is present, objdump only disassembles
// symbols with names matching the regular expression.
//
// Alternate usage:
//
//	go tool objdump binary start end
//
// In this mode, objdump disassembles the binary starting at the start address and
// stopping at the end address. The start and end addresses are program
// counters written in hexadecimal with optional leading 0x prefix.
// In this mode, objdump prints a sequence of stanzas of the form:
//
//	file:line
//	 address: assembly
//	 address: assembly
//	 ...
//
// Each stanza gives the disassembly for a contiguous range of addresses
// all mapped to the same original source file and line number.
// This mode is intended for use by pprof.
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"cmd/internal/objfile"
)

var printCode = flag.Bool("S", false, "print Go code alongside assembly")
var symregexp = flag.String("s", "", "only dump symbols matching this regexp")
var gnuAsm = flag.Bool("gnu", false, "print GNU assembly next to Go assembly (where supported)")
var stats = flag.Bool("stats", false, "print object statistics and section sizes, and exit")
var symRE *regexp.Regexp

func usage() {
	fmt.Fprintf(os.Stderr, "usage: go tool objdump [-S] [-gnu] [-s symregexp] binary [start end]\n\n")
	flag.PrintDefaults()
	os.Exit(2)
}

// Print some statistics out about the object file.
// Right now, we're only really printing a small amount of stuff here.
func printStats(f *objfile.File, out io.Writer) {
	for _, entry := range f.Entries() {
		name := entry.Name()
		if len(name) == 0 {
			name = "[NONE]"
		}
		io.WriteString(out, fmt.Sprintf("Name: %s\n", name))
		io.WriteString(out, "\tSection\t\tSize\n")

		if start, data, err := entry.Text(); err != nil {
			panic(err)
		} else {
			io.WriteString(out, fmt.Sprintf("\tTEXT\t\t%d\t0x%08x\n", len(data), start))
		}

		encoding := entry.Encoding()
		if data, err := entry.PCLNData(); err != nil {
			io.WriteString(out, "\tNo PCLN found.")
		} else {
			printPclnStats(out, encoding, data)
		}
	}
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("objdump: ")

	flag.Usage = usage
	flag.Parse()
	if flag.NArg() != 1 && flag.NArg() != 3 {
		usage()
	}

	if *symregexp != "" {
		re, err := regexp.Compile(*symregexp)
		if err != nil {
			log.Fatalf("invalid -s regexp: %v", err)
		}
		symRE = re
	}

	f, err := objfile.Open(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	if *stats {
		printStats(f, os.Stdout)
		return
	}

	dis, err := f.Disasm()
	if err != nil {
		log.Fatalf("disassemble %s: %v", flag.Arg(0), err)
	}

	switch flag.NArg() {
	default:
		usage()
	case 1:
		// disassembly of entire object
		dis.Print(os.Stdout, symRE, 0, ^uint64(0), *printCode, *gnuAsm)

	case 3:
		// disassembly of PC range
		start, err := strconv.ParseUint(strings.TrimPrefix(flag.Arg(1), "0x"), 16, 64)
		if err != nil {
			log.Fatalf("invalid start PC: %v", err)
		}
		end, err := strconv.ParseUint(strings.TrimPrefix(flag.Arg(2), "0x"), 16, 64)
		if err != nil {
			log.Fatalf("invalid end PC: %v", err)
		}
		dis.Print(os.Stdout, symRE, start, end, *printCode, *gnuAsm)
	}
}
