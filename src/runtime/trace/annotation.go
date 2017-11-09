package trace

import (
	"context"
	"fmt"
	"sync/atomic"
	_ "unsafe"
)

type traceContextKey struct{}

// NewContext creates a child context with a new task instance with
// the type taskType. If the input context contains a task, the
// new task is its subtask.
//
// The taskType is used to classify task instances and analysis tools
// like the go execution tracer may assume there are only a bounded
// number of unique task types in the system.
//
// The returned end function is used to mark the task's end.
// Tools like the go execution trace measure the latency of each task
// using the task creation time and the end time determined by the task
// end mark, and provide the latency distribution per task type. Thus,
// the lack of the end mark can result in inaccuracy in analysis so the
// end fuction must be called at least once for accurate results.
// In case the end function is called multiple times, only the first
// call should be used in the latency measurement.
//
//   ctx, taskEnd := trace.NewContext(ctx, "awesome task")
//   trace.WithSpan(ctx, prepWork)
//   // preparation of the task
//   go func() {  // continue processing the task in a separate goroutine.
//       defer taskEnd()
//       trace.WithSpan(ctx, remainingWork)
//   }
//
func NewContext(pctx context.Context, taskType string) (ctx context.Context, end func()) {
	// TODO: should we limit the taskType length?
	// the string should fit in the buffer :-) (64K - some overhead bytes)
	pid := fromContext(pctx).id
	id := newID()
	userTaskCreate(id, pid, taskType)
	s := &task{id: id}
	return context.WithValue(pctx, traceContextKey{}, s), func() {
		userTaskEnd(id)
	}

	// We allocate a new task and the end function even when
	// the tracing is disabled because the context and the detach
	// function can be used across trace enable/disable boundaries,
	// which complicates the problem.
	//
	// For example, consider the following scenario:
	//   - trace is enabled.
	//   - trace.WithSpan is called, so a new context ctx
	//     with a new span is created.
	//   - trace is disabled.
	//   - trace is enabled again.
	//   - trace APIs with the ctx is called. Is the ID in the task
	//   a valid one to use?
	//
	// TODO(hyangah): reduce the overhead at least when
	// tracing is disabled. Maybe the id can embed a tracing
	// round number and ignore ids generated from previous
	// tracing round.
}

func fromContext(ctx context.Context) *task {
	if s, ok := ctx.Value(traceContextKey{}).(*task); ok {
		return s
	}
	return &bgTask
}

type task struct {
	id uint64
	// TODO(hyangah): record parent id?
}

func newID() uint64 {
	// TODO(hyangah): implement
	return 0
}

var bgTask = task{id: uint64(0)}

// Log emits a one-off event with the given key-value message. The execution
// tracer allows analysis by filtering and grouping goroutines or spans.
// Keys can be empty and the API assumes there are only handful number of
// unique keys in the system.
func Log(ctx context.Context, key, value string) {
	id := fromContext(ctx).id
	userLog(id, key, value)
}

// Logf is like Log, but the value is formatted using the specified format spec.
func Logf(ctx context.Context, key, format string, args ...interface{}) {
	if IsEnabled() {
		Log(ctx, key, fmt.Sprintf(format, args...))
	}
}

// WithSpan starts a span, the time interval during which the calling goroutine
// is working on behalf of the task in the given context, and runs the given
// function.
//
// If the context doesn't carry a task, the span is considered
// a span for the background task.
func WithSpan(ctx context.Context, name string, fn func(context.Context)) {
	// TODO: Consider exposing StartSpan and deferred endSpan as well.
	//    end := trace.StartSpan(ctx, name)
	//    defer end()
	// WithSpan helps avoiding misuse of the API but in practice,
	// this is very restrictive:
	// - Use of WithSpan makes the stack traces captured from
	//   span start and end are identical.
	// - Refactoring the existing code to use WithSpan is sometimes
	//   hard and makes the code less readable.
	//     e.g. code block nested deep in the loop with various
	//          exit point with return values
	// - Refactoring the code to use this API with closure can
	//   cause different GC behavior such as retaining some parameters
	//   longer.
	// This causes more churns in code than I hoped, and sometimes
	// makes the code less readable.

	const start = uint64(0)
	const end = uint64(1)
	id := fromContext(ctx).id
	userSpan(id, start, name)
	defer userSpan(id, end, name) // name may not be recorded for span end event
	fn(ctx)
}

// IsEnabled returns whether tracing is enabled.
// The information is advisory only. The tracing status
// may have changed by the time this function returns.
func IsEnabled() bool {
	enabled := atomic.LoadInt32(&tracing.enabled)
	return enabled == 1
}

//
// Function bodies are defined in runtime/trace.go
//

// emits UserTaskCreate event.
func userTaskCreate(internalID, internalParentID uint64, name string)

// emits UserTaskEnd event.
func userTaskEnd(internalID uint64)

// emits UserSpan event.
func userSpan(internalID, mode uint64, name string)

// emits UserLog event.
func userLog(internalID uint64, key, val string)
