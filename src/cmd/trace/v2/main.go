// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package trace

import (
	"bytes"
	"fmt"
	"internal/trace"
	"internal/trace/traceviewer"
	tracev2 "internal/trace/v2"
	"io"
	"log"
	"net"
	"net/http"
	"os"

	"internal/trace/v2/raw"

	"cmd/internal/browser"
)

/*
TODO:

- Split traces by generation. Problem: Doesn't seem possible with the high level reader API.
- Support printing raw events (-d flag): Problem: Needs access to raw parsing API.
- ... everything else

*/

// Main is the main function for cmd/trace v2.
func Main(traceFile, httpAddr, pprof string, debug bool) error {
	data, err := os.ReadFile(traceFile)
	if err != nil {
		return fmt.Errorf("failed to read trace file: %w", err)
	}

	if debug {
		return debugEvents(data)
	}

	ln, err := net.Listen("tcp", httpAddr)
	if err != nil {
		return fmt.Errorf("failed to create server socket: %w", err)
	}

	log.Print("Preparing trace for viewer...")
	addr := "http://" + ln.Addr().String()
	parsed, err := parseTrace(data)
	if err != nil {
		return err
	}
	log.Print("Splitting trace for viewer...")
	ranges, err := splitTrace(parsed)
	if err != nil {
		return err
	}
	log.Printf("Analyzing goroutines...")
	gSummaries := trace.SummarizeGoroutines(parsed.events)

	log.Printf("Opening browser. Trace viewer is listening on %s", addr)
	browser.Open(addr)

	mux := http.NewServeMux()
	mux.Handle("/", traceviewer.MainHandler(ranges))
	mux.Handle("/trace", traceviewer.TraceHandler())
	mux.Handle("/jsontrace", JSONTraceHandler(parsed))
	mux.Handle("/static/", traceviewer.StaticHandler())
	mux.HandleFunc("/goroutines", GoroutinesHandlerFunc(gSummaries))
	mux.HandleFunc("/goroutine", GoroutineHandler(gSummaries))

	// Install MMU handlers.
	mutatorUtil := func(flags trace.UtilFlags) ([][]trace.MutatorUtil, error) {
		return trace.MutatorUtilizationV2(parsed.events, flags), nil
	}
	traceviewer.InstallMMUHandlers(mux, ranges, mutatorUtil)

	err = http.Serve(ln, mux)
	return fmt.Errorf("failed to start http server: %w", err)
}

type parsedTrace struct {
	events []tracev2.Event
}

func parseTrace(data []byte) (parsedTrace, error) {
	r, err := tracev2.NewReader(bytes.NewReader(data))
	if err != nil {
		return parsedTrace{}, fmt.Errorf("failed to create trace reader: %w", err)
	}
	var t parsedTrace
	for {
		ev, err := r.ReadEvent()
		if err == io.EOF {
			break
		} else if err != nil {
			return parsedTrace{}, fmt.Errorf("failed to read event: %w", err)
		}
		t.events = append(t.events, ev)
	}
	return t, nil
}

// splitTrace splits the trace into a number of ranges,
// each resulting in approx 100MB of json output
// (trace viewer can hardly handle more).
func splitTrace(parsed parsedTrace) ([]traceviewer.Range, error) {
	s, c := traceviewer.SplittingTraceConsumer(100 << 20) // 100M
	if err := generateTrace(parsed, c); err != nil {
		return nil, err
	}
	return s.Ranges, nil
}

// TODO(fg) Also offer a way to print/debug the high level events from the
// trace.NewReader API?
func debugEvents(traceData []byte) error {
	rr, err := raw.NewReader(bytes.NewReader(traceData))
	if err != nil {
		return err
	}

	for {
		ev, err := rr.ReadEvent()
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}
		fmt.Printf("%s\n", ev.String())
	}
}
