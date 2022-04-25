// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Functions for working the "go test -json" format.

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type testEvent struct {
	Time    time.Time // encodes as an RFC3339-format string
	Action  string
	Package string
	Test    string
	Elapsed float64 // seconds
	Output  string
}

// readTestEvents reads test2json events from r until it reaches EOF,
// emitting them on the returned channel. At EOF, it closes the
// channel.
//
// It can synthesize a new type of event whose Action field is
// "error", with the Output field set to an error in processing the
// stream.
func readTestEvents(r io.Reader) <-chan testEvent {
	out := make(chan testEvent)
	go func() {
		defer close(out)

		// We do our own line reading instead of using
		// json.Decoder.More because this allows us to recover
		// if there's an error.
		scanner := bufio.NewScanner(r)
		// Set a large line length limit.
		scanner.Buffer(nil, 16<<20)
		for scanner.Scan() {
			var ev testEvent
			err := json.Unmarshal(scanner.Bytes(), &ev)
			if err == nil {
				out <- ev
			} else {
				out <- testEvent{
					Action: "error",
					Output: fmt.Sprintf("non-test2json line: %s", scanner.Text()),
				}
			}
		}
		if err := scanner.Err(); err != nil {
			out <- testEvent{
				Action: "error",
				Output: fmt.Sprintf("I/O error reading test output: %s", err),
			}
		}
	}()
	return out
}

// unjson reads "go test -json" data from r and writes non-verbose,
// plain text test output to w. If there are parsing errors, they are
// reported to stderr. If there is an error writing to w, it is
// returned.
//
// pkgs is an optional slice of package names known to be in the
// output. If non-nil, unjson will deterministically emit package
// output in this order. Any packages that appear in r but not in pkgs
// will be emitted after pkgs in the order they were first observed.
func unjson(r io.Reader, w io.Writer, pkgs []string) error {
	events := readTestEvents(r)

	// Make a copy of pkgs. This will also serve to track the
	// observation order of unexpected packages and to flush
	// packages as they complete.
	pkgs = append([]string(nil), pkgs)

	// Accumulate test output so we can emit only failure output.
	testOut := make(map[string]*pkg)
	for _, p := range pkgs {
		testOut[p] = newPkg()
	}

	for ev := range events {
		// If we see any output from a package, record that
		// we've seen that package.
		if ev.Package != "" && testOut[ev.Package] == nil {
			testOut[ev.Package] = newPkg()
			pkgs = append(pkgs, ev.Package)
		}

		switch ev.Action {
		case "error":
			// Error reading JSON.
			fmt.Fprintln(os.Stderr, ev.Output)

		case "run":
			if ev.Test != "" {
				testOut[ev.Package].tests[ev.Test] = new(lines)
			}

		case "output":
			if ev.Test == "" {
				// Top-level package output.
				//
				// Ignore just "PASS" because this
				// isn't printed in non-verbose mode
				// and the next line will be the "ok"
				// line we do want to print.
				if ev.Output == "PASS\n" {
					break
				}
				testOut[ev.Package].extra.add(ev.Output)
				break
			}
			// Lines starting with "=== " are progress
			// updates only shown in verbose mode (like
			// "=== RUN"). These are never indented.
			if strings.HasPrefix(ev.Output, "=== ") {
				break
			}
			// Accumulate output in case this test
			// fails.
			if lines, ok := testOut[ev.Package].tests[ev.Test]; ok {
				lines.add(ev.Output)
			} else {
				fmt.Fprintf(os.Stderr, "\"output\" event from unexpected test: %+v", ev)
			}

		case "fail":
			if ev.Test == "" {
				// Package failed.
				testOut[ev.Package].done = true
				testOut[ev.Package].failed = true
				break
			}
			// Leave failed tests in the map.

		case "pass", "skip":
			if ev.Test == "" {
				// Package passed, so mark it done.
				testOut[ev.Package].done = true
				break
			}

			// The test passed, so delete accumulated
			// output.
			delete(testOut[ev.Package].tests, ev.Test)
		}

		// Flush completed tests.
		for len(pkgs) > 0 && testOut[pkgs[0]].done {
			pkg := testOut[pkgs[0]]
			delete(testOut, pkgs[0])
			pkgs = pkgs[1:]
			if err := pkg.emitTests(w); err != nil {
				return err
			}
		}
	}

	if len(testOut) != 0 {
		fmt.Fprintf(os.Stderr, "packages neither passed nor failed:\n")
		for pkgName, pkg := range testOut {
			fmt.Fprintf(os.Stderr, "%s\n", pkgName)
			pkg.emitTests(w)
		}
	}
	return nil
}

// pkg records test output from a package.
type pkg struct {
	// tests records output lines for in-flight and failed tests.
	tests map[string]*lines

	// extra records package-level test output.
	extra lines

	// done is set when all tests from this package are done, and
	// failed indicates if a completed package failed.
	done, failed bool
}

func newPkg() *pkg {
	return &pkg{tests: make(map[string]*lines)}
}

func (p *pkg) emitTests(w io.Writer) error {
	// Sort tests.
	tests := make([]string, 0, len(p.tests))
	for k := range p.tests {
		tests = append(tests, k)
	}
	// Emit each test.
	for _, test := range tests {
		if err := p.tests[test].emit(w); err != nil {
			return err
		}
	}
	// Emit package-level output.
	return p.extra.emit(w)
}

type lines struct {
	lines []string
}

func (l *lines) add(line string) {
	l.lines = append(l.lines, line)
}

func (l *lines) emit(w io.Writer) error {
	lines := l.lines
	// The last line could be a (possibly indented) "--- FAIL". In
	// non-verbose mode, this is printed *before* the test log.
	if len(lines) > 0 && isFailLine(lines[len(lines)-1]) {
		if _, err := io.WriteString(w, lines[len(lines)-1]); err != nil {
			return err
		}
		lines = lines[:len(lines)-1]
	}
	for _, line := range lines {
		if _, err := io.WriteString(w, line); err != nil {
			return err
		}
	}
	return nil
}

func isFailLine(line string) bool {
	// The line may be indented.
	line = strings.TrimLeft(line, " ")
	return strings.HasPrefix(line, "--- FAIL: ")
}
