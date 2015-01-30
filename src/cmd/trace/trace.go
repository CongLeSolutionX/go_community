// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Trace is a tool for viewing trace files.

Trace files can be generated with:
- runtime/pprof.StartTrace
- net/http/pprof package
- go test -trace

Given a trace file produced by 'go test':
	go test -trace=trace.out pkg

Open a web browser displaying trace:
	go tool trace pkg.test trace.out
*/
package main

import (
	"bufio"
	"flag"
	"fmt"
	"internal/trace"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
)

const usageMessage = "" +
	`Usage of 'go tool trace':
Given a trace file produced by 'go test':
	go test -trace=trace.out pkg

Open a web browser displaying trace:
	go tool trace pkg.test trace.out
`

var (
	// The binary file name, left here for serveSVGProfile.
	programBinary string
	// Parsed trace events.
	events []*trace.Event
)

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, usageMessage)
		os.Exit(2)
	}
	flag.Parse()

	// Usage information when no arguments.
	if flag.NArg() != 2 {
		flag.Usage()
	}

	// Open trace file.
	programBinary = flag.Arg(0)
	tracef, err := os.Open(flag.Arg(1))
	if err != nil {
		dief("failed to open trace file: %v\n", err)
	}
	defer tracef.Close()

	// Parse and symbolize.
	events, err = trace.Parse(bufio.NewReader(tracef))
	if err != nil {
		dief("failed to parse trace: %v\n", err)
	}
	err = trace.Symbolize(events, programBinary)
	if err != nil {
		dief("failed to symbolize trace: %v\n", err)
	}

	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		dief("failed to create server socket: %v\n", err)
	}
	// Open browser.
	if !startBrowser("http://" + ln.Addr().String()) {
		dief("failed to start browser\n")
	}
	// Start http server.
	err = http.Serve(ln, nil)
	dief("failed to start http server: %v\n", err)
}

// startBrowser tries to open the URL in a browser
// and reports whether it succeeds.
// Note: copied from x/tools/cmd/cover/html.go
func startBrowser(url string) bool {
	// try to start the browser
	var args []string
	switch runtime.GOOS {
	case "darwin":
		args = []string{"open"}
	case "windows":
		args = []string{"cmd", "/c", "start"}
	default:
		args = []string{"xdg-open"}
	}
	cmd := exec.Command(args[0], append(args[1:], url)...)
	return cmd.Start() == nil
}

func dief(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg, args...)
	os.Exit(1)
}
