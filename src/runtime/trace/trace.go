// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package trace contains facilities for programs to generate trace
// for Go execution tracer.
//
// Tracing runtime activities
//
// The execution trace captures a wide range of execution events such as
// goroutine creation/blocking/unblocking, syscall enter/exit/block,
// GC-related events, changes of heap size, processor start/stop, etc.
// A precise nanosecond-precision timestamp and a stack trace is
// captured for most events. The generated trace can be interpreted
// using `go tool trace`.
//
// Support for tracing tests and benchmarks built with the standard
// testing package is built into `go test`. For example, the following
// command runs the test in the current directory and writes the trace
// file (trace.out).
//
//    go test -trace=test.out
//
// This runtime/trace package provides APIs to add equivalent tracing
// support to a standalone program. See the Example that demonstrates
// how to use this API to enable tracing.
//
// There is also a standard HTTP interface to trace data. Adding the
// following line will install a handler under the /debug/pprof/trace URL
// to download live trace (and also profile data):
//
//     import _ "net/http/pprof"
//
// See the net/http/pprof package for more details about the
// debug endpoints.
//
// User annotation
//
// Package trace provides user annotation APIs that allows to
// log interesting events during execution. trace.Log emits the key-value
// message to the execution trace along with additional information
// such as when and in which goroutine the logging occurred.
//
// Task is a logical operation such as an RPC request, an HTTP request,
// or an interesting function call which requires one or more goroutines
// to work together to accomplish. Execution tracer tool provides latency
// distributions for each task type found in the trace.
//
// Span is a time interval during which a goroutine is working on a Task.
// Span is defined locally in goroutine and can be nested. There can be
// zero or more spans working for a Task.
//
//   ctx, end := trace.NewContext(pctx, "request")
//   trace.WithSpan(ctx, "readRequest", func(ctx context.Context) {
//      req := readRequest(ctx)
//      go trace.WithSpan(ctx, "processRequest", func(ctx context.Context) {
//           /* do something with req */
//          end()
//      })
//   })
//
// The above code will produce a request Task that spans across two goroutines
// and the latency of the request is from when the request was created to
// when the processRequest goroutine finished.
//
// It is possible to have a Task that has no Span associated. For example,
// a request may be sitting in the work queue waiting for being picked up
// by consumer goroutine in the following program. In this case, the total
// latency of the request includes the time in the work queue.
//
//   producer goroutine:
//      ctx, end := trace.NewContext(pctx, "request")
//      work <- req{ctx, req}
//
//   goroutine 2:
//      for w := range work {
//         trace.WithSpan(w.ctx, "work", processReq(w.req))
//      }
//
// Logging API allows adding a log entry associated with a task. Log entries
// are timestamped key-value pairs and contains the calling goroutine.
// Execution tracer provides UIs to filter and group tasks or goroutines
// using the keys and values supplied with the Log calls.
// For example, if we create a Task for each HTTP server handler call
// and call Log with the response code with a user-determined key
// (e.g. 'status'), we can later ask the execution tracer to display
// the latency distribution of the HTTP request Tasks whose status matches
// the specific status code. If we create a Task for each RPC request and
// annotate each Task with its RPC ID using Log function, we can visualize all
// goroutines involved in serving the RPC request with a certain ID.
package trace

import (
	"io"
	"runtime"
	"sync"
	"sync/atomic"
)

// Start enables tracing for the current program.
// While tracing, the trace will be buffered and written to w.
// Start returns an error if tracing is already enabled.
func Start(w io.Writer) error {
	tracing.Lock()
	defer tracing.Unlock()

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
	atomic.StoreInt32(&tracing.enabled, 1)
	return nil
}

// Stop stops the current tracing, if any.
// Stop only returns after all the writes for the trace have completed.
func Stop() {
	tracing.Lock()
	defer tracing.Unlock()
	atomic.StoreInt32(&tracing.enabled, 0)

	runtime.StopTrace()
}

var tracing struct {
	sync.Mutex       // gate mutators (Start, Stop)
	enabled    int32 // accessed via atomic
}
