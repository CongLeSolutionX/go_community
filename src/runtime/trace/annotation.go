package trace

import (
	"context"
	"fmt"
	"sync/atomic"
	_ "unsafe" // for linkname
)

// WithSpan returns a new Context containing a span of the type and
// id. The type is used by the tracer UI to classify spans, and the
// id is a user-defined id to distinguish span instances. There is an
// internally generated hidden id to avoid unintended collision from
// use of duplicate user-provided ids.
// If tracing is enabled, this emits a span creation event. The current
// goroutine is associated with the span in the trace until the goroutine
// ends, or Attach with a different context is called, whichever occurs
// first.
//
// Calling the returned end function explicitly marks the span end.
// It takes a status string that the tracer UI uses to classify spans
// in conjunction with the span type. The end function does not detach
// the span from the current goroutine.
//
// The end function can run multiple times, but status from duplicate
// calls will be ignored.
func WithSpan(ctx context.Context, typ, id string) (newCtx context.Context, end func(status string)) {
	// retrieve internalPID, pType, pID from ctx.
	// runtime_traceUserSpanStart(internalID, typ, id)
	return ctx, func(_ string) {}
}

// Attach associates the span in the context with the current goroutine
// and newly created goroutines. A goroutine can have at most one span
// attached to it at any given time.
func Attach(ctx context.Context) (detach func()) {
	// retrieve internalID, typ, id from ctx.
	// runtime_traceUserAttach(internalID, status)
	return func() {} // set back to the previous context
}

// Log emits a one-off event with the given message
// if tracing is enabled. If the goroutine has an attached span,
// the msg is associated with the span.
func Log(msg string) {
	runtime_traceUserLog(msg)
}

// Logf is similar to Log.
func Logf(format string, args ...interface{}) {
	if IsEnabled() {
		Log(fmt.Sprintf(format, args...))
	}
}

// IsEnabled returns whether tracing is enabled.
// The information is advisory only. The tracing status
// may have changed by the time this function returns.
func IsEnabled() bool {
	enabled := atomic.LoadInt32(&tracing.enabled)
	return enabled == 1
}

// Function bodies are defined in runtime/trace.go

func runtime_traceUserSpanStart(internalID uint64, typ, id string)

func runtime_traceUserSpanEnd(internalID uint64, status string)

func runtime_traceUserAttach(internalID uint64, typ, id string)

func runtime_traceUserLog(msg string)

// Do calls f with a newly created context with the span of
// the given typ and id. The span is attached to the goroutine
// during the execution of f, and ends when f returns.
// The return value of f is used as the end status message of the span.
func Do(ctx context.Context, typ, id string, f func(context.Context) string) {
	nctx, end := WithSpan(ctx, typ, id)
	var status string
	defer func() {
		end(status)
		Attach(ctx)
	}()

	status = f(nctx)
}
