package traceviewer

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	"internal/trace/traceviewer/format"

	"internal/trace"
)

type TraceConsumer struct {
	ConsumeTimeUnit    func(unit string)
	ConsumeViewerEvent func(v *format.Event, required bool)
	ConsumeViewerFrame func(key string, f format.Frame)
	Flush              func()
}

func ViewerDataTraceConsumer(w io.Writer, start, end int64) TraceConsumer {
	allFrames := make(map[string]format.Frame)
	requiredFrames := make(map[string]format.Frame)
	enc := json.NewEncoder(w)
	written := 0
	index := int64(-1)

	io.WriteString(w, "{")
	return TraceConsumer{
		ConsumeTimeUnit: func(unit string) {
			io.WriteString(w, `"displayTimeUnit":`)
			enc.Encode(unit)
			io.WriteString(w, ",")
		},
		ConsumeViewerEvent: func(v *format.Event, required bool) {
			index++
			if !required && (index < start || index > end) {
				// not in the range. Skip!
				return
			}
			WalkStackFrames(allFrames, v.Stack, func(id int) {
				s := strconv.Itoa(id)
				requiredFrames[s] = allFrames[s]
			})
			WalkStackFrames(allFrames, v.EndStack, func(id int) {
				s := strconv.Itoa(id)
				requiredFrames[s] = allFrames[s]
			})
			if written == 0 {
				io.WriteString(w, `"traceEvents": [`)
			}
			if written > 0 {
				io.WriteString(w, ",")
			}
			enc.Encode(v)
			// TODO: get rid of the extra \n inserted by enc.Encode.
			// Same should be applied to splittingTraceConsumer.
			written++
		},
		ConsumeViewerFrame: func(k string, v format.Frame) {
			allFrames[k] = v
		},
		Flush: func() {
			io.WriteString(w, `], "stackFrames":`)
			enc.Encode(requiredFrames)
			io.WriteString(w, `}`)
		},
	}
}

// WalkStackFrames calls fn for id and all of its parent frames from allFrames.
func WalkStackFrames(allFrames map[string]format.Frame, id int, fn func(id int)) {
	for id != 0 {
		f, ok := allFrames[strconv.Itoa(id)]
		if !ok {
			break
		}
		fn(id)
		id = f.Parent
	}
}

type Mode int

const (
	ModeGoroutineOriented Mode = 1 << iota
	ModeTaskOriented
)

func NewEmitter(c TraceConsumer, mode Mode) *Emitter {
	return &Emitter{c: c, mode: mode}
}

type Emitter struct {
	c                        TraceConsumer // TODO: rename this?
	mode                     Mode
	heapStats, prevHeapStats heapStats
	gomaxprocs               uint64
}

func (t *Emitter) Gomaxprocs(v uint64) {
	if v > t.gomaxprocs {
		t.gomaxprocs = v
	}
}

// TODO(fg) create a higher level abstraction, e.g. for running on a proc?
func (t *Emitter) Event(e *format.Event) {
	t.c.ConsumeViewerEvent(e, true)
}

func (t *Emitter) HeapAlloc(ts time.Duration, v uint64) {
	t.heapStats.heapAlloc = v
	t.flushHeapCounters(ts)
}

func (t *Emitter) NextGC(ts time.Duration, v uint64) {
	// Ignore non-sensical values. Workaround for
	// https://github.com/golang/go/issues/63864.
	// TODO(fg) Remove this once the problem has been fixed.
	const PB = 1 << 50
	if v > PB {
		return
	}

	t.heapStats.nextGC = v
	t.flushHeapCounters(ts)
}

func (t *Emitter) flushHeapCounters(ts time.Duration) {
	if t.prevHeapStats == t.heapStats {
		return
	}
	diff := uint64(0)
	if t.heapStats.nextGC > t.heapStats.heapAlloc {
		diff = t.heapStats.nextGC - t.heapStats.heapAlloc
	}
	t.c.ConsumeViewerEvent(&format.Event{
		Name:  "Heap",
		Phase: "C",
		Time:  viewerTime(ts),
		PID:   1,
		Arg:   &format.HeapCountersArg{Allocated: t.heapStats.heapAlloc, NextGC: diff},
	}, false)
	t.prevHeapStats = t.heapStats
}

func (t *Emitter) Flush() {
	t.c.ConsumeViewerEvent(&format.Event{Name: "thread_name", Phase: "M", PID: format.ProcsSection, TID: trace.GCP, Arg: &format.NameArg{Name: "GC"}}, true)
	t.c.ConsumeViewerEvent(&format.Event{Name: "thread_sort_index", Phase: "M", PID: format.ProcsSection, TID: trace.GCP, Arg: &format.SortIndexArg{Index: -6}}, true)

	t.c.ConsumeViewerEvent(&format.Event{Name: "thread_name", Phase: "M", PID: format.ProcsSection, TID: trace.NetpollP, Arg: &format.NameArg{Name: "Network"}}, true)
	t.c.ConsumeViewerEvent(&format.Event{Name: "thread_sort_index", Phase: "M", PID: format.ProcsSection, TID: trace.NetpollP, Arg: &format.SortIndexArg{Index: -5}}, true)

	t.c.ConsumeViewerEvent(&format.Event{Name: "thread_name", Phase: "M", PID: format.ProcsSection, TID: trace.TimerP, Arg: &format.NameArg{Name: "Timers"}}, true)
	t.c.ConsumeViewerEvent(&format.Event{Name: "thread_sort_index", Phase: "M", PID: format.ProcsSection, TID: trace.TimerP, Arg: &format.SortIndexArg{Index: -4}}, true)

	t.c.ConsumeViewerEvent(&format.Event{Name: "thread_name", Phase: "M", PID: format.ProcsSection, TID: trace.SyscallP, Arg: &format.NameArg{Name: "Syscalls"}}, true)
	t.c.ConsumeViewerEvent(&format.Event{Name: "thread_sort_index", Phase: "M", PID: format.ProcsSection, TID: trace.SyscallP, Arg: &format.SortIndexArg{Index: -3}}, true)

	if t.mode&ModeGoroutineOriented == 0 {
		for i := uint64(0); i <= t.gomaxprocs; i++ {
			t.c.ConsumeViewerEvent(&format.Event{
				Name:  "thread_name",
				Phase: "M",
				PID:   format.ProcsSection,
				TID:   uint64(i),
				Arg:   &format.NameArg{Name: fmt.Sprintf("Proc %v", i)},
			}, true)
			t.c.ConsumeViewerEvent(&format.Event{
				Name:  "thread_sort_index",
				Phase: "M",
				PID:   format.ProcsSection,
				TID:   uint64(i),
				Arg:   &format.SortIndexArg{Index: int(i)},
			}, true)
		}
	}

	t.c.Flush()
}

type heapStats struct {
	heapAlloc uint64
	nextGC    uint64
}

func viewerTime(t time.Duration) float64 {
	return float64(t) / float64(time.Microsecond)
}
