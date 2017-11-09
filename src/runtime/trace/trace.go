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
// User annotation (EXPERIMENTAL)
//
// Package trace provides user annotation APIs that can be used to
// log interesting events during execution. This API is EXPERIMENTAL
// and may change in the future.
//
// There are three types of user annotations: log messages, spans,
// and tasks.
//
// Log emits a message to the execution trace along with additional
// information such as the category of the message and which goroutine
// called Log. Execution tracer provides UIs to filter and group goroutines
// using the log message category (key) and the message (value) supplied
// in Log.
//
// A span is for logging a time interval during a goroutine's execution.
// By definition, a span starts and ends in the same goroutine.
// Spans can be nested to represent subintervals.
// For example, the following code records four spans in the execution
// trace to trace the durations of sequential steps in a cappuccino making
// operation.
//
//   trace.WithSpan(ctx, "Cappuccino", func(ctx context.Context) {
//      trace.WithSpan(ctx, "Milk", steamMilk)
//      trace.WithSpan(ctx, "Espresso", extractCoffee)
//      trace.WithSpan(ctx, "Assemble", mixMilkCoffee)
//   })
//
// A task is a higher-level component that aids tracing of logical
// operations such as an RPC request, an HTTP request, or an
// interesting local operation which may require multiple goroutines
// working together. NewContext creates a new task and embeds it in
// the the returned context.Context object. The Log and WithSpan
// APIs recognize the task information embedded in the context.Context
// and record it in the execution trace. With the information, the
// tracer tool can associate spans and log messages related to
// processing the task.
//
// For example, assume that we decided to froth
// milk in a separate goroutine so we can parallelize steaming milk
// and extracting coffee. With a task, the execution tracer can
// identify the milk steaming goroutine and the main cappuccino
// making goroutine.
//
//   trace.WithSpan(ctx, "Cappuccino", func(ctx context.Context) {
//     ctx, end := trace.NewContext(pctx, "Cappuccino")
//     defer end()  // mark cappuccino making is complete.
//
//     trace.Log(ctx, "order", orderID)
//
//     var milk chan bool
//     go trace.WithSpan(ctx, "Milk", func(ctx context.Context) {
//        steamMilk(ctx)
//        milk<-true
//     })
//     trace.WithSpan(ctx, "Espresso", extraceCoffee)
//
//     <-milk
//     trace.WithSpan(ctx, "Assemble", mixMilkCoffee)
//   })
//
// The tracer tool computes the latency of a task by measuring the
// time between the task creation and the task end and provides
// latency distributions for each task type found in the trace.
//
// Note that it is possible that a task is not associated with
// any span at a certain point of time.
//
//   producer goroutine:
//      ctx, end := trace.NewContext(pctx, "request")
//      work <- req{ctx: ctx, req: req, end: end}
//      // Until the req is picked up by the consumer goroutine,
//      // the task exists in the system but has no Span associated.
//
//   consumer goroutine:
//      for w := range work {
//	   // While WithSpan runs, this goroutine will be associated with
//	   // the task created by the producer goroutine and stored in w.ctx.
//         trace.WithSpan(w.ctx, "work", processReq(w.req))
//
//         w.end()  // mark the end of the request processing
//      }
//
// In the above example, a request task created by the producer goroutine
// will not have any span associated until the consumer goroutine
// receives the task from the work channel and starts processing.
// The total latency of the request task however should include
// the time in the work queue as well.
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
