// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Command lockcheck records and analyzes the Go runtime lock graph.
//
// For detailed usage run lockcheck -help.
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"

	"lockcheck/lockgraph"
	"lockcheck/server"

	"github.com/aclements/go-moremath/graph/graphout"
)

func main() {
	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, `Usage: %s [flags] {-load dump.json | subcommand...}

This tool records the lock graph of Go runtime locks and produces
reports on the lock graph. It exits with a non-zero status if the
subcommand fails or the lock graph contains cycles.

Each node in the lock graph represents a lock class and the edges
indicate which locks are acquired while other locks are held. Cycles
in the lock graph indicate potential deadlocks.

When given a subcommand, it executes that subcommand and records the
combined lock graph of all processes spawned. Alternatively, it can
load and analyze a previously-recorded lock graph.

For example, to record the lock graph of the runtime test:

   (export GOFLAGS="-tags=locklog -gcflags=all=-d=maymorestack=runtime.lockLogMoreStack"
    go test -c && \
    ../../misc/lockcheck/lockcheck -dump runtime.locks ./runtime.test -test.short)

Locks involved in cycles can then be viewed with:

   ../../misc/lockcheck/lockcheck -load runtime.locks -dot - | dot -Tx11

If no output is specified, it defaults to a text report on stdout.

Flags:
`, os.Args[0])
		flag.PrintDefaults()
	}
	flagLoad := flag.String("load", "", "read lock graph from `path` instead of recording a subcommand")
	flagDump := flag.String("dump", "", "output lock graph to `path`")
	flagFull := flag.Bool("full", false, "report full lock graph (default: report only locks involved in cycles)")
	flagDot := flag.String("dot", "", "output dot for lock graph to `file`")
	flagText := flag.String("text", "", "output text report with stacks to `file`")

	flag.Parse()
	if (*flagLoad == "") == (len(flag.Args()) == 0) {
		flag.Usage()
		os.Exit(2)
	}

	// Use text output if no other output is specified.
	if *flagDump == "" && *flagDot == "" && *flagText == "" {
		*flagText = "-"
	}

	// Collect the lock graph by monitoring a subcommand or just
	// load it from a previous dump.
	var err error
	var lockGraph *lockgraph.Graph
	if *flagLoad == "" {
		args := flag.Args()
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		builder := server.NewGraphBuilder()
		var any bool
		any, err = server.Run(cmd, builder)
		if any {
			lockGraph = builder.Finish()
		}
	} else if *flagLoad == "-" {
		lockGraph, err = lockgraph.Load(os.Stdin)
	} else {
		func() {
			var f *os.File
			f, err = os.Open(*flagLoad)
			if err != nil {
				return
			}
			defer f.Close()
			lockGraph, err = lockgraph.Load(f)
		}()
	}
	if lockGraph == nil && err != nil {
		log.Print(err)
		os.Exit(1)
	}

	// Dump the lock graph in JSON so it can be loaded back later.
	// We always dump the full graph, since later analysis can filter it.
	if *flagDump != "" {
		func() {
			f, err := createArg(*flagDump)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()
			if err = lockgraph.Dump(f, lockGraph); err != nil {
				log.Fatal(err)
			}
		}()
	}

	// Reduce the graph to just its cycles.
	cycles := lockGraph.Filter(lockgraph.Cycles(lockGraph))

	// Filter the reporting graph unless the user asked for the
	// whole thing.
	if !*flagFull {
		lockGraph = cycles
	}

	if *flagDot != "" {
		func() {
			f, err := createArg(*flagDot)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()
			graphout.Dot{Label: lockGraph.Label}.Fprint(f, lockGraph)
		}()
	}

	if *flagText != "" {
		func() {
			f, err := createArg(*flagText)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()
			lockgraph.ReportText(f, lockGraph)
		}()
	}

	// Report problems.
	exit := 0
	if cycles.NumNodes() != 0 {
		fmt.Println("warning: runtime lock graph contains cycles")
		exit = 1
	}
	if err != nil {
		log.Print(err)
		exit = 2
	}
	os.Exit(exit)
}

// createArg creates file path as specified by a command-line flag. If
// path is "-", it returns os.Stdout with a no-op Close method.
func createArg(path string) (io.WriteCloser, error) {
	if path == "-" {
		return &noCloser{os.Stdout}, nil
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	return f, err
}

type noCloser struct{ f *os.File }

func (c *noCloser) Write(p []byte) (n int, err error) { return c.f.Write(p) }
func (c *noCloser) Close() error                      { return nil }
