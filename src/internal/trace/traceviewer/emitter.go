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
	c                            TraceConsumer // TODO: rename this?
	mode                         Mode
	heapStats, prevHeapStats     heapStats
	gstates, prevGstates         [gStateCount]int64
	threadStats, prevThreadStats [threadStateCount]int64
	gomaxprocs                   uint64
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
	t.emitHeapCounters(ts)
}

// TODO(fg) only emit if in time range. Also update cmd/trace.
func (e *Emitter) GoroutineTransition(ts time.Duration, from, to GState) {
	e.gstates[from]--
	e.gstates[to]++

	if e.prevGstates == e.gstates {
		return
	}
	// TODO: if tsWithinRange(ev.Ts, ctx.startTime, ctx.endTime) {
	e.Event(&format.Event{
		Name:  "Goroutines",
		Phase: "C",
		Time:  viewerTime(ts),
		PID:   1,
		Arg: &format.GoroutineCountersArg{
			Running:   uint64(e.gstates[GRunning]),
			Runnable:  uint64(e.gstates[GRunnable]),
			GCWaiting: uint64(e.gstates[GWaitingGC]),
		},
	})
	e.prevGstates = e.gstates
}

func (e *Emitter) IncThreadStateCount(ts time.Duration, state ThreadState, delta int64) {
	e.threadStats[state] += delta

	if e.prevThreadStats == e.threadStats {
		return
	}

	// TODO: if tsWithinRange(ev.Ts, ctx.startTime, ctx.endTime) {
	e.Event(&format.Event{
		Name:  "Threads",
		Phase: "C",
		Time:  viewerTime(ts),
		PID:   1,
		Arg: &format.ThreadCountersArg{
			Running:   int64(e.threadStats[ThreadStateRunning]),
			InSyscall: int64(e.threadStats[ThreadStateInSyscall]),
			// TODO(fg) Why is InSyscallRuntime not included here?
		},
	})
	e.prevThreadStats = e.threadStats
}

func (t *Emitter) HeapGoal(ts time.Duration, v uint64) {
	// Ignore non-sensical values. Workaround for
	// https://github.com/golang/go/issues/63864.
	// TODO(fg) Remove this once the problem has been fixed.
	const PB = 1 << 50
	if v > PB {
		return
	}

	t.heapStats.nextGC = v
	t.emitHeapCounters(ts)
}

func (t *Emitter) emitHeapCounters(ts time.Duration) {
	if t.prevHeapStats == t.heapStats {
		return
	}
	diff := uint64(0)
	if t.heapStats.nextGC > t.heapStats.heapAlloc {
		diff = t.heapStats.nextGC - t.heapStats.heapAlloc
	}
	// TODO: if tsWithinRange(ev.Ts, ctx.startTime, ctx.endTime) {
	t.c.ConsumeViewerEvent(&format.Event{
		Name:  "Heap",
		Phase: "C",
		Time:  viewerTime(ts),
		PID:   1,
		Arg:   &format.HeapCountersArg{Allocated: t.heapStats.heapAlloc, NextGC: diff},
	}, false)
	t.prevHeapStats = t.heapStats
}

func (e *Emitter) Flush() {
	e.processMeta(format.StatsSection, "STATS", 0)
	if e.mode&ModeTaskOriented != 0 {
		e.processMeta(format.TasksSection, "TASKS", 1)
	}

	if e.mode&ModeGoroutineOriented != 0 {
		e.processMeta(format.ProcsSection, "G", 2)
	} else {
		e.processMeta(format.ProcsSection, "PROCS", 2)
	}

	e.threadMeta(format.ProcsSection, trace.GCP, "GC", -6)
	e.threadMeta(format.ProcsSection, trace.NetpollP, "Network", -5)
	e.threadMeta(format.ProcsSection, trace.TimerP, "Timers", -4)
	e.threadMeta(format.ProcsSection, trace.SyscallP, "Syscalls", -3)

	// Display rows for Ps if we are in the default trace view mode (not goroutine-oriented presentation)
	if e.mode&ModeGoroutineOriented == 0 {
		for i := uint64(0); i <= e.gomaxprocs; i++ {
			e.threadMeta(format.ProcsSection, i, fmt.Sprintf("Proc %v", i), int(i))
		}
	}

	e.c.Flush()
}

func (e *Emitter) threadMeta(sectionID, tid uint64, name string, priority int) {
	e.c.ConsumeViewerEvent(&format.Event{
		Name:  "thread_name",
		Phase: "M",
		PID:   sectionID,
		TID:   tid,
		Arg:   &format.NameArg{Name: name},
	}, true)
	e.c.ConsumeViewerEvent(&format.Event{
		Name:  "thread_sort_index",
		Phase: "M",
		PID:   sectionID,
		TID:   tid,
		Arg:   &format.SortIndexArg{Index: priority},
	}, true)
}

func (e *Emitter) processMeta(sectionID uint64, name string, priority int) {
	e.c.ConsumeViewerEvent(&format.Event{
		Name:  "process_name",
		Phase: "M",
		PID:   sectionID,
		Arg:   &format.NameArg{Name: name},
	}, true)
	e.c.ConsumeViewerEvent(&format.Event{
		Name:  "process_sort_index",
		Phase: "M",
		PID:   sectionID,
		Arg:   &format.SortIndexArg{Index: priority},
	}, true)
}

type heapStats struct {
	heapAlloc uint64
	nextGC    uint64
}

func viewerTime(t time.Duration) float64 {
	return float64(t) / float64(time.Microsecond)
}

type GState int

const (
	GDead GState = iota
	GRunnable
	GRunning
	GWaiting
	GWaitingGC

	gStateCount
)

type ThreadState int

const (
	ThreadStateInSyscall ThreadState = iota
	ThreadStateInSyscallRuntime
	ThreadStateRunning

	threadStateCount
)
