// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package trace contains facilities for programs to generate trace
// for Go execution tracer.
//
// The execution trace captures a wide range of execution events such as
// goroutine creation/blocking/unblocking, syscall enter/exit/block,
// GC-related events, changes of heap size, processor stop/stop, etc.
// A precise nanosecond-precision timestamp and a stack trace is
// captured for most events. The generated trace can be interpreted
// using `go tool trace`.
//
// Tracing a Go program
//
// Support for tracing tests and benchmarks built with the standard
// testing package is built into `go test`. For example, the following
// command runs the test in the current directory and writes the trace
// file (trace.out).
//
//    go test -trace=test.out
//
// This runtime/trace package provides APIs to add equivalent tracing
// support to a standalone program. Add code like the following to
// your main function:
//
//    var traceFile = flag.String("trace", "", "write Go execution trace")
//    func main() {
//       flag.Parse()
//       if *traceFile != "" {
//           f, err := os.Create(*traceFile)
//           if err != nil {
//              log.Fatal("could not create execution trace output: ", err)
//           }
//           if err := trace.Start(f); err != nil {
//              log.Fatal("could not start execution trace: ", err)
//           }
//           defer func() {
//               trace.Stop()
//               if err := f.Close(); err != nil {
//                  log.Fatal("could not write execution trace: ", err)
//               }
//           }()
//        }
//
//        // ... rest of the program ...
//    }
//
// There is also a standard HTTP interface to profiling data. Adding the
// following line will install handlers under the /debug/pprof/trace URL
// to download live profiles:
//
//     import _ "net/http/pprof"
//
// See the net/http/pprof package for more details.
package trace

import (
	"io"
	"runtime"
)

// Start enables tracing for the current program.
// While tracing, the trace will be buffered and written to w.
// Start returns an error if tracing is already enabled.
func Start(w io.Writer) error {
	if err := runtime.StartTrace(); err != nil {
		return err
	}
	go func() {
		for {
			data := runtime.ReadTrace()
			if data == nil {
				break
			}
			w.Write(data)
		}
	}()
	return nil
}

// Stop stops the current tracing, if any.
// Stop only returns after all the writes for the trace have completed.
func Stop() {
	runtime.StopTrace()
}
