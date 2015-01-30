// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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

func usage() {
	fmt.Fprintln(os.Stderr, usageMessage)
	os.Exit(2)
}

var (
	bin    string
	events []*trace.Event
)

func main() {
	flag.Usage = usage
	flag.Parse()

	// Usage information when no arguments.
	if flag.NArg() != 2 {
		flag.Usage()
	}

	// Open trace file.
	bin = flag.Arg(0)
	tracef, err := os.Open(flag.Arg(1))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open trace file: %v\n", err)
		os.Exit(1)
	}
	defer tracef.Close()

	// Parse and symbolize.
	events, err = trace.Parse(bufio.NewReader(tracef))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse trace: %v\n", err)
		os.Exit(1)
	}
	err = trace.Symbolize(events, bin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to symbolize trace: %v\n", err)
		os.Exit(1)
	}

	// Start http server.
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create server socket: %v\n", err)
		os.Exit(1)
	}
	go func() {
		err := http.Serve(ln, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to start http server: %v\n", err)
			os.Exit(1)
		}
	}()

	// Open browser.
	if !startBrowser("http://" + ln.Addr().String()) {
		fmt.Fprintf(os.Stderr, "failed to start browser\n")
		os.Exit(1)
	}
	select {}
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
