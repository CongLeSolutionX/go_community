// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"cmd/internal/cov"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"runtime"
	"runtime/pprof"
	"strings"
)

var verbflag = flag.Int("v", 0, "Verbose trace output level")
var hflag = flag.Bool("h", false, "Panic on fatal errors (for stack trace)")
var sumcountsflag = flag.Bool("sumcounts", false, "Sum coverage counts across input files.")
var whflag = flag.Bool("hw", false, "Panic on warnings (for stack trace)")
var indirsflag = flag.String("i", "", "Input dirs to examine (comma separated)")
var liveflag = flag.Bool("live", false, "Select only live (executed) funcs.")
var pkgreflag = flag.String("pkg", "", "Restrict output to package matching this regular expression.")
var cpuprofileflag = flag.String("cpuprofile", "", "Write CPU profile to specified file")
var memprofileflag = flag.String("memprofile", "", "Write CPU profile to specified file")
var memprofilerateflag = flag.Int("memprofilerate", 0, "Write CPU profile to specified file")

var pkgre *regexp.Regexp

var atExitFuncs []func()

func atExit(f func()) {
	atExitFuncs = append(atExitFuncs, f)
}

func Exit(code int) {
	for i := len(atExitFuncs) - 1; i >= 0; i-- {
		f := atExitFuncs[i]
		atExitFuncs = atExitFuncs[:i]
		f()
	}
	os.Exit(code)
}

func verb(vlevel int, s string, a ...interface{}) {
	if *verbflag >= vlevel {
		fmt.Printf(s, a...)
		fmt.Printf("\n")
	}
}

func warn(s string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "warning: ")
	fmt.Fprintf(os.Stderr, s, a...)
	fmt.Fprintf(os.Stderr, "\n")
	if *whflag {
		panic("unexpected warning")
	}
}

func fatal(s string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "error: ")
	fmt.Fprintf(os.Stderr, s, a...)
	fmt.Fprintf(os.Stderr, "\n")
	if *hflag {
		panic("fatal error")
	}
	Exit(1)
}

func usage(msg string) {
	if len(msg) > 0 {
		fmt.Fprintf(os.Stderr, "error: %s\n", msg)
	}
	fmt.Fprintf(os.Stderr, "usage: go tool covdata [command]\n\n")
	fmt.Fprintf(os.Stderr, "Commands are:\n\n")
	fmt.Fprintf(os.Stderr, "dump        dump coverage data in text format\n")
	fmt.Fprintf(os.Stderr, "merge       merge data files together\n")
	fmt.Fprintf(os.Stderr, "subtract    subtract one set of data files from another set\n")
	fmt.Fprintf(os.Stderr, "intersect   generate intersection of two sets of data files\n")
	Exit(2)
}

type covOperation interface {
	cov.CovDataVisitor
	Setup()
	Usage(string)
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("covdata: ")

	// First argument should be mode/subcommand.
	if len(os.Args) < 2 {
		usage("missing command selector")
	}

	// Select mode
	var op covOperation
	cmd := os.Args[1]
	switch cmd {
	case "merge":
		op = makeMergeOp()
	case "dump":
		op = makeDumpOp()
	case "subtract":
		op = makeSubtractIntersectOp("subtract")
	case "intersect":
		op = makeSubtractIntersectOp("intersect")
	default:
		usage(fmt.Sprintf("unknown command selector %q", cmd))
	}

	// Edit out command selector, then parse flags.
	os.Args = append(os.Args[:1], os.Args[2:]...)
	flag.Usage = func() {
		op.Usage("")
	}
	flag.Parse()

	// Mode-independent flag setup
	verb(1, "starting mode-independent setup")
	if flag.NArg() != 0 {
		op.Usage("unknown extra arguments")
	}
	if *pkgreflag != "" {
		p, err := regexp.Compile(*pkgreflag)
		if err != nil {
			op.Usage(fmt.Sprintf("unable to compile -pkg argument %q", *pkgreflag))
		}
		pkgre = p
	}
	if *cpuprofileflag != "" {
		f, err := os.Create(*cpuprofileflag)
		if err != nil {
			fatal("%v", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			fatal("%v", err)
		}
		atExit(pprof.StopCPUProfile)
	}
	if *memprofileflag != "" {
		if *memprofilerateflag != 0 {
			runtime.MemProfileRate = *memprofilerateflag
		}
		f, err := os.Create(*memprofileflag)
		if err != nil {
			fatal("%v", err)
		}
		atExit(func() {
			runtime.GC()
			const writeLegacyFormat = 1
			if err := pprof.Lookup("heap").WriteTo(f, writeLegacyFormat); err != nil {
				fatal("%v", err)
			}
		})
	} else {
		// Not doing memory profiling; disable it entirely.
		runtime.MemProfileRate = 0
	}

	// Mode-dependent setup.
	op.Setup()

	// ... off and running now.
	verb(1, "starting perform")

	indirs := strings.Split(*indirsflag, ",")
	vis := cov.CovDataVisitor(op)
	var flags cov.CovDataReaderFlags
	if *hflag {
		flags |= cov.PanicOnError
	}
	if *whflag {
		flags |= cov.PanicOnWarning
	}
	reader := cov.MakeCovDataReader(vis, indirs, *verbflag, flags, pkgre)
	if err := reader.Visit(); err != nil {
		log.Fatalf("%v\n", err)
	}
	verb(1, "leaving main")
	Exit(0)
}
