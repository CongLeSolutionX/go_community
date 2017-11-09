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
// to download a live trace:
//
//     import _ "net/http/pprof"
//
// See the net/http/pprof package for more details about all of the
// debug endpoints installed by this import.
//
// User annotation
//
// Package trace provides user annotation APIs that can be used to
// log interesting events during execution. There are three types of user
// annotations: log messages, tasks, and spans.
//
// Log emits a message to the execution trace along with additional
// information such as the category of the message and which goroutine
// called Log. Execution tracer provides UIs to filter and group goroutines
// using the log message category (key) and the message (value) supplied
// in Log.
//
// A task is a logical operation such as an RPC request, an HTTP request,
// or an interesting local operation which requires multiple goroutines
// working together. Also note that it is possible that no goroutines are
// working on the task at any given moment, for example, if the task is
// waiting in a queue. The tracer tool provides latency distributions
// for each task type found in the trace.
//
// The Log API provides a way to associate the log message with a task.
// That allows the tracer tool to provide task search capability based
// on the associcated log messages.
// For example, if we create a task for each HTTP server handler call
// and call Log with the response code with a user-determined key
// (e.g. 'status'), we can later ask the execution tracer to list
// the HTTP request tasks whose status matches only the specific status
// code and computes the latency distribution of those matching tasks.
// If we create a task for each RPC request and annotate each Task with
// its RPC ID using Log function, we can visualize all goroutines
// involved in serving the RPC request with a certain ID.
//
// A span is for logging with a time interval. It can be also used to
// annotate a time interval during a goroutine's execution with which
// Task the goroutine is processing or working on.
// In contrast with a task, which can encompass multiple goroutines, a
// span is local to a goroutine.
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
// The above code will produce a request Task that spans across two goroutines.
// The latency of the request is from when the request is created to
// when the processRequest goroutine finishes.
//
// It is possible to have a Task that has no Span associated. For example,
// a request may be sitting in the work queue waiting to be picked up
// by consumer goroutine in the following program. In this case, the total
// latency of the request includes the time in the work queue.
//
//   producer goroutine:
//      ctx, end := trace.NewContext(pctx, "request")
//      work <- req{ctx: ctx, req: req}
//
//   goroutine 2:
//      for w := range work {
//	   // While WithSpan runs, this goroutine will be associated with
//	   // the task created by the producer goroutine and stored in w.ctx.
//         trace.WithSpan(w.ctx, "work", processReq(w.req))
//      }
//
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
