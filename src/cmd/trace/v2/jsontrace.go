// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package trace

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"internal/trace"
	"internal/trace/traceviewer"
	"internal/trace/traceviewer/format"
	tracev2 "internal/trace/v2"
)

func JSONTraceHandler(parsed *parsedTrace) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := int64(0)
		end := int64(math.MaxInt64)
		var err error
		if startStr, endStr := r.FormValue("start"), r.FormValue("end"); startStr != "" && endStr != "" {
			// If start/end arguments are present, we are rendering a range of the trace.
			start, err = strconv.ParseInt(startStr, 10, 64)
			if err != nil {
				log.Printf("failed to parse start parameter %q: %v", startStr, err)
				return
			}
			end, err = strconv.ParseInt(endStr, 10, 64)
			if err != nil {
				log.Printf("failed to parse end parameter %q: %v", endStr, err)
				return
			}
		}

		c := traceviewer.ViewerDataTraceConsumer(w, start, end)
		if err := generateTrace(parsed, c); err != nil {
			log.Printf("failed to generate trace: %v", err)
		}
	})
}

// traceContext is a wrapper around a traceviewer.Emitter with some additional
// information that's useful to most parts of trace viewer JSON emission.
type traceContext struct {
	*traceviewer.Emitter
	startTime tracev2.Time
}

// elapsed returns the elapsed time between the trace time and the start time
// of the trace.
func (ctx *traceContext) elapsed(now tracev2.Time) time.Duration {
	return now.Sub(ctx.startTime)
}

func generateTrace(parsed *parsedTrace, c traceviewer.TraceConsumer) error {
	ctx := &traceContext{
		Emitter:   traceviewer.NewEmitter(c, 0, time.Duration(0), time.Duration(math.MaxInt64)),
		startTime: parsed.events[0].Time(),
	}
	defer ctx.Flush()

	gStates := map[tracev2.GoID]*gState{}
	inSyscall := map[tracev2.ProcID]*gState{}
	globalRanges := map[string]activeRange{}
	pRanges := map[tracev2.Range]activeRange{}
	seenSync := false

	// startTime is used to calculate the relative time of each event.
	for _, ev := range parsed.events {
		switch ev.Kind() {
		case tracev2.EventSync:
			seenSync = true
		case tracev2.EventStackSample:
			if ev.Proc() == tracev2.NoProc {
				// We have nowhere to put this in the UI.
				// TODO(mknyszek): A per-M mode could place all of these.
				break
			}
			ctx.Instant(traceviewer.InstantEvent{
				Name:  "CPU profile sample",
				Ts:    ctx.elapsed(ev.Time()),
				Proc:  uint64(ev.Proc()),
				Stack: ctx.Stack(viewerFrames(ev.Stack())),
			})
		case tracev2.EventRangeBegin:
			r := ev.Range()
			switch r.Scope.Kind {
			case tracev2.ResourceGoroutine:
				gStates[r.Scope.Goroutine()].rangeBegin(ev.Time(), r.Name, ev.Stack())
			case tracev2.ResourceProc:
				pRanges[r] = activeRange{ev.Time(), ev.Stack()}
			case tracev2.ResourceNone:
				globalRanges[r.Name] = activeRange{ev.Time(), ev.Stack()}
			}
		case tracev2.EventRangeActive:
			r := ev.Range()
			switch r.Scope.Kind {
			case tracev2.ResourceGoroutine:
				gStates[r.Scope.Goroutine()].rangeActive(r.Name)
			case tracev2.ResourceProc:
				// If we've seen a Sync event, then Active events are always redundant.
				if !seenSync {
					// Otherwise, they extend back to the start of the trace.
					pRanges[r] = activeRange{ctx.startTime, ev.Stack()}
				}
			case tracev2.ResourceNone:
				// If we've seen a Sync event, then Active events are always redundant.
				if !seenSync {
					// Otherwise, they extend back to the start of the trace.
					globalRanges[r.Name] = activeRange{ctx.startTime, ev.Stack()}
				}
			}
		case tracev2.EventRangeEnd:
			r := ev.Range()
			switch r.Scope.Kind {
			case tracev2.ResourceGoroutine:
				gStates[r.Scope.Goroutine()].rangeEnd(ev.Time(), r.Name, ev.Stack(), ctx)
			case tracev2.ResourceProc:
				// Emit proc-based ranges.
				ar := pRanges[r]
				ctx.Slice(traceviewer.SliceEvent{
					Name:     r.Name,
					Ts:       ctx.elapsed(ar.time),
					Dur:      ev.Time().Sub(ar.time),
					Proc:     uint64(r.Scope.Proc()),
					Stack:    ctx.Stack(viewerFrames(ar.stack)),
					EndStack: ctx.Stack(viewerFrames(ev.Stack())),
				})
				delete(pRanges, r)
			case tracev2.ResourceNone:
				// Only emit GC events, because we have nowhere to
				// put other events.
				ar := globalRanges[r.Name]
				if strings.Contains(r.Name, "GC") {
					ctx.Slice(traceviewer.SliceEvent{
						Name:     r.Name,
						Ts:       ctx.elapsed(ar.time),
						Dur:      ev.Time().Sub(ar.time),
						Proc:     trace.GCP,
						Stack:    ctx.Stack(viewerFrames(ar.stack)),
						EndStack: ctx.Stack(viewerFrames(ev.Stack())),
					})
				}
				delete(globalRanges, r.Name)
			}
		case tracev2.EventMetric:
			m := ev.Metric()
			switch m.Name {
			case "/memory/classes/heap/objects:bytes":
				ctx.HeapAlloc(ctx.elapsed(ev.Time()), m.Value.Uint64())
			case "/gc/heap/goal:bytes":
				ctx.HeapGoal(ctx.elapsed(ev.Time()), m.Value.Uint64())
			case "/sched/gomaxprocs:threads":
				ctx.Gomaxprocs(m.Value.Uint64())
			}
		case tracev2.EventLabel:
			l := ev.Label()
			if l.Resource.Kind == tracev2.ResourceGoroutine {
				gStates[l.Resource.Goroutine()].extraLabel = l.Label
			}
		case tracev2.EventStateTransition:
			st := ev.StateTransition()
			switch st.Resource.Kind {
			case tracev2.ResourceProc:
				proc := st.Resource.Proc()
				viewerEv := traceviewer.InstantEvent{
					Ts:    ctx.elapsed(ev.Time()),
					Proc:  uint64(proc),
					Stack: ctx.Stack(viewerFrames(ev.Stack())),
				}

				from, to := st.Proc()
				if from == to {
					// Filter out no-op events.
					break
				}
				if to.Executing() {
					viewerEv.Name = "proc start"
					viewerEv.Arg = format.ThreadIDArg{ThreadID: uint64(ev.Thread())}
					ctx.IncThreadStateCount(ctx.elapsed(ev.Time()), traceviewer.ThreadStateRunning, 1)
				}
				if from.Executing() {
					viewerEv.Name = "proc stop"
					ctx.IncThreadStateCount(ctx.elapsed(ev.Time()), traceviewer.ThreadStateRunning, -1)

					// Check if this proc was in a syscall before it stopped.
					// This means the syscall blocked. We need to emit it to the
					// viewer at this point because we only display the time the
					// syscall occupied a P when the viewer is in per-P mode.
					//
					// TODO(mknyszek): We could do better in a per-M mode because
					// all events have to happen on *some* thread, and in v2 traces
					// we know what that thread is.
					gs, ok := inSyscall[proc]
					if ok {
						// Emit syscall slice for blocked syscall.
						gs.syscallEnd(ev.Time(), true, ctx)
						gs.stop(ev.Time(), ev.Stack(), ctx)
						delete(inSyscall, proc)
					}
				}
				// TODO(mknyszek): Consider modeling procs differently and have them be
				// transition to and from NotExist when GOMAXPROCS changes. We can emit
				// events for this to clearly delineate GOMAXPROCS changes.

				if viewerEv.Name != "" {
					ctx.Instant(viewerEv)
				}
			case tracev2.ResourceGoroutine:
				goID := st.Resource.Goroutine()

				// If we haven't seen this goroutine before, create a new
				// gState for it.
				gs, ok := gStates[goID]
				if !ok {
					gs = newGState(goID)
					gStates[goID] = gs
				}

				// If we haven't already named this goroutine, try to name it.
				if !gs.named {
					if name := goroutineName(ev, st); name != "" {
						gs.label += fmt.Sprintf(" %s", name)
						gs.named = true
						gs.isSystemG = trace.IsSystemGoroutine(name)
					}
				}

				// Goroutine state machine in dot notation:
				//
				//	digraph G {
				//	    # Guaranteed by the internal/trace/v2 API.
				//	    Undetermined -> Running;
				//	    Undetermined -> Waiting;
				//	    Undetermined -> Runnable;
				//	    Undetermined -> Syscall;
				//
				//	    NotExist -> Runnable;
				//	    NotExist -> Syscall;
				//	    Runnable -> Running;
				//	    Running -> Syscall;
				//	    Running -> Waiting;
				//	    Syscall -> Running;
				//	    Syscall -> Runnable;
				//	    Syscall -> NotExist;
				//	    Waiting -> Runnable;
				//	    Running -> Runnable;
				//	    Running -> NotExist;
				//	}
				//
				// Note: although this state machine describes valid transitions,
				// we should generally be robust to any transition. The general
				// parsing strategy here is thus to use Executing to identify the
				// the general high-level slices of goroutines executing and to check
				// for specific transitions to create more detailed annotations.
				//
				// TODO(mknyszek): Include this in the design doc.
				from, to := st.Goroutine()
				if from == to {
					// Filter out no-op events.
					break
				}
				if from == tracev2.GoRunning && !to.Executing() {
					if to == tracev2.GoWaiting {
						// Goroutine started blocking.
						gs.block(ev.Time(), ev.Stack(), st.Reason, ctx)
					} else {
						gs.stop(ev.Time(), ev.Stack(), ctx)
					}
				}
				if !from.Executing() && to == tracev2.GoRunning {
					start := ev.Time()
					if from == tracev2.GoUndetermined {
						// Back-date the event to the start of the trace.
						start = ctx.startTime
					}
					gs.start(start, ev.Proc(), ctx)
				}

				// Handle unblock events.
				if from == tracev2.GoWaiting {
					// Unblocking goroutine.
					name := "unblock"
					proc := uint64(ev.Proc())
					if strings.Contains(gs.startBlockReason, "network") {
						// Emit an unblock instant event for the "Network" lane.
						ctx.Instant(traceviewer.InstantEvent{
							Name:  "unblock",
							Ts:    ctx.elapsed(ev.Time()),
							Proc:  trace.NetpollP,
							Stack: ctx.Stack(viewerFrames(ev.Stack())),
						})
						gs.startBlockReason = ""
						proc = trace.NetpollP
					}
					gs.setStartCause(ev.Time(), name, proc, ev.Stack())
				}
				if from == tracev2.GoNotExist && to == tracev2.GoRunnable {
					// Goroutine was created.
					gs.setStartCause(ev.Time(), "go", uint64(ev.Proc()), ev.Stack())
				}
				if from == tracev2.GoSyscall && to != tracev2.GoRunning {
					// Exiting blocked syscall.
					name := "exit blocked syscall"
					gs.setStartCause(ev.Time(), name, trace.SyscallP, ev.Stack())

					// Emit an syscall exit instant event for the "Syscall" lane.
					ctx.Instant(traceviewer.InstantEvent{
						Name:  name,
						Ts:    ctx.elapsed(ev.Time()),
						Proc:  trace.SyscallP,
						Stack: ctx.Stack(viewerFrames(ev.Stack())),
					})
				}

				// Handle syscalls.
				if to == tracev2.GoSyscall && ev.Proc() != tracev2.NoProc {
					// Write down that we've entered a syscall. Note: we might have no P here
					// if we're in a cgo callback or this is a transition from GoUndetermined
					// (i.e. the G has been blocked in a syscall).
					gs.syscallBegin(ev.Time(), ev.Stack())
					inSyscall[ev.Proc()] = gs
				}
				// Check if we're exiting a non-blocking syscall.
				_, didNotBlock := inSyscall[ev.Proc()]
				if from == tracev2.GoSyscall && didNotBlock {
					gs.syscallEnd(ev.Time(), false, ctx)
					delete(inSyscall, ev.Proc())
				}

				// Note down the goroutine transition.
				_, inMarkAssist := gs.activeRanges["GC mark assist"]
				ctx.GoroutineTransition(ctx.elapsed(ev.Time()), viewerGState(from, inMarkAssist), viewerGState(to, inMarkAssist))
			}
		}
	}

	// Finish off the JSON trace.
	lastTime := parsed.events[len(parsed.events)-1].Time()
	for _, gs := range gStates {
		gs.finalize(lastTime, ctx)
	}
	for r, ar := range pRanges {
		ctx.Slice(traceviewer.SliceEvent{
			Name:  r.Name,
			Ts:    ctx.elapsed(ar.time),
			Dur:   lastTime.Sub(ar.time),
			Proc:  uint64(r.Scope.Proc()),
			Stack: ctx.Stack(viewerFrames(ar.stack)),
		})
	}
	for name, ar := range globalRanges {
		if !strings.Contains(name, "GC") {
			continue
		}
		ctx.Slice(traceviewer.SliceEvent{
			Name:  name,
			Ts:    ctx.elapsed(ar.time),
			Dur:   lastTime.Sub(ar.time),
			Proc:  trace.GCP,
			Stack: ctx.Stack(viewerFrames(ar.stack)),
		})
	}
	return nil
}

// viewerFrames returns the frames of the stack of ev. The given frame slice is
// used to store the frames to reduce allocations.
func viewerFrames(stk tracev2.Stack) []*trace.Frame {
	var frames []*trace.Frame
	stk.Frames(func(f tracev2.StackFrame) bool {
		frames = append(frames, &trace.Frame{
			PC:   f.PC,
			Fn:   f.Func,
			File: f.File,
			Line: int(f.Line),
		})
		return true
	})
	return frames
}

func viewerGState(state tracev2.GoState, inMarkAssist bool) traceviewer.GState {
	switch state {
	case tracev2.GoUndetermined:
		return traceviewer.GDead
	case tracev2.GoNotExist:
		return traceviewer.GDead
	case tracev2.GoRunnable:
		return traceviewer.GRunnable
	case tracev2.GoRunning:
		return traceviewer.GRunning
	case tracev2.GoWaiting:
		if inMarkAssist {
			return traceviewer.GWaitingGC
		}
		return traceviewer.GWaiting
	case tracev2.GoSyscall:
		// N.B. A goroutine in a syscall is considered "executing" (state.Executing() == true).
		return traceviewer.GRunning
	default:
		panic(fmt.Sprintf("unknown GoState: %s", state.String()))
	}
}

type gState struct {
	label      string // Basic name.
	named      bool   // Whether label has been set.
	extraLabel string // EventLabel extension.
	isSystemG  bool
	executing  tracev2.ProcID // The proc this goroutine is executing on.

	// lastStopStack is the stack trace at the point of the last
	// call to the stop method. This tends to be a more reliable way
	// of picking up stack traces, since the parser doesn't provide
	// a stack for every state transition event.
	lastStopStack tracev2.Stack

	// activeRanges is the set of all active ranges on the goroutine.
	activeRanges map[string]activeRange

	// startRunning is the most recent event that caused a goroutine to
	// transition to GoRunning.
	startRunningTime tracev2.Time

	// startSyscall is the most recent event that caused a goroutine to
	// transition to GoSyscall.
	syscall struct {
		time   tracev2.Time
		stack  tracev2.Stack
		active bool
	}

	// startBlockReason is the StateTransition.Reason of the most recent
	// event that caused a gorotuine to transition to GoWaiting.
	startBlockReason string

	// startCause is the event that allowed this goroutine to start running.
	// It's used to generate flow events. This is typically something like
	// an unblock event or a goroutine creation event.
	//
	// startCauseProc is the proc on which startCause happened, but is
	// listed separately because the cause may have happened without a P,
	// or we may want to attribute it to an event that appears in a special
	// lane (e.g. trace.NetpollP).
	startCause struct {
		time  tracev2.Time
		name  string
		proc  uint64
		stack tracev2.Stack
	}
}

// activeRange represents an active EventRange* range.
type activeRange struct {
	time  tracev2.Time
	stack tracev2.Stack
}

// newGState constructs a new goroutine state for the goroutine
// identified by the provided ID.
func newGState(goID tracev2.GoID) *gState {
	return &gState{
		label:        fmt.Sprintf("G%d", goID),
		activeRanges: make(map[string]activeRange),
		executing:    tracev2.NoProc,
	}
}

// start indicates that a goroutine has started running on a proc.
func (gs *gState) start(ts tracev2.Time, proc tracev2.ProcID, ctx *traceContext) {
	// Set the time for all the active ranges.
	for name := range gs.activeRanges {
		gs.activeRanges[name] = activeRange{ts, tracev2.NoStack}
	}
	if gs.startCause.name != "" {
		// It has a start cause. Emit a flow event.
		ctx.Arrow(traceviewer.ArrowEvent{
			Name:      gs.startCause.name,
			Start:     ctx.elapsed(gs.startCause.time),
			End:       ctx.elapsed(ts),
			FromProc:  gs.startCause.proc,
			ToProc:    uint64(proc),
			FromStack: ctx.Stack(viewerFrames(gs.startCause.stack)),
		})
		gs.startCause.time = 0
		gs.startCause.name = ""
		gs.startCause.proc = 0
		gs.startCause.stack = tracev2.NoStack
	}
	gs.executing = proc
	gs.startRunningTime = ts
}

// rangeBegin indicates the start of a special range of time.
func (gs *gState) rangeBegin(ts tracev2.Time, name string, stack tracev2.Stack) {
	if gs.executing != tracev2.NoProc {
		// If we're executing, start the slice from here.
		gs.activeRanges[name] = activeRange{ts, stack}
	} else {
		// If the goroutine isn't executing, there's no place for
		// us to create a slice from. Wait until it starts executing.
		gs.activeRanges[name] = activeRange{0, stack}
	}
}

// rangeActive indicates that a special range of time has been in progress.
func (gs *gState) rangeActive(name string) {
	if gs.executing != tracev2.NoProc {
		// If we're executing, and the range is active, then start
		// from wherever the goroutine started running from.
		gs.activeRanges[name] = activeRange{gs.startRunningTime, tracev2.NoStack}
	} else {
		// If the goroutine isn't executing, there's no place for
		// us to create a slice from. Wait until it starts executing.
		gs.activeRanges[name] = activeRange{0, tracev2.NoStack}
	}
}

// rangeEnd indicates the end of a special range of time.
func (gs *gState) rangeEnd(ts tracev2.Time, name string, stack tracev2.Stack, ctx *traceContext) {
	if gs.executing != tracev2.NoProc {
		r := gs.activeRanges[name]
		ctx.Slice(traceviewer.SliceEvent{
			Name:     name,
			Ts:       ctx.elapsed(r.time),
			Dur:      ts.Sub(r.time),
			Proc:     uint64(gs.executing),
			Stack:    ctx.Stack(viewerFrames(r.stack)),
			EndStack: ctx.Stack(viewerFrames(stack)),
		})
	}
	delete(gs.activeRanges, name)
}

// setStartCause sets the reason a goroutine will be allowed to start soon.
// For example, via unblocking or exiting a blocked syscall.
func (gs *gState) setStartCause(ts tracev2.Time, name string, proc uint64, stack tracev2.Stack) {
	gs.startCause.time = ts
	gs.startCause.name = name
	gs.startCause.proc = proc
	gs.startCause.stack = stack
}

// syscallBegin indicates that the goroutine entered a syscall on a proc.
func (gs *gState) syscallBegin(ts tracev2.Time, stack tracev2.Stack) {
	gs.syscall.time = ts
	gs.syscall.stack = stack
	gs.syscall.active = true
}

// syscallEnd indicates that the goroutine has exited the part of the syscall
// where it's executing on a proc.
func (gs *gState) syscallEnd(ts tracev2.Time, blocked bool, ctx *traceContext) {
	if !gs.syscall.active {
		return
	}
	blockString := "no"
	if blocked {
		blockString = "yes"
	}
	ctx.Slice(traceviewer.SliceEvent{
		Name:  "syscall",
		Ts:    ctx.elapsed(gs.syscall.time),
		Dur:   ts.Sub(gs.syscall.time),
		Proc:  uint64(gs.executing),
		Stack: ctx.Stack(viewerFrames(gs.syscall.stack)),
		Arg:   format.BlockedArg{Blocked: blockString},
	})
	gs.syscall.active = false
	gs.syscall.time = 0
	gs.syscall.stack = tracev2.NoStack
}

// block indicates that the goroutine has stopped executing on a proc -- specifically,
// it blocked for some reason.
func (gs *gState) block(ts tracev2.Time, stack tracev2.Stack, reason string, ctx *traceContext) {
	gs.startBlockReason = reason
	gs.stop(ts, stack, ctx)
}

// stop indicates that the goroutine has stopped executing on a proc.
func (gs *gState) stop(ts tracev2.Time, stack tracev2.Stack, ctx *traceContext) {
	// Emit range slices.
	for name, r := range gs.activeRanges {
		// Check invariant.
		if r.time == 0 {
			panic("silently broken trace or generator invariant (activeRanges time != 0)) not held")
		}
		ctx.Slice(traceviewer.SliceEvent{
			Name:  name,
			Ts:    ctx.elapsed(r.time),
			Dur:   ts.Sub(r.time),
			Proc:  uint64(gs.executing),
			Stack: ctx.Stack(viewerFrames(r.stack)),
		})
	}
	// Clear the range info.
	for name := range gs.activeRanges {
		gs.activeRanges[name] = activeRange{-1, tracev2.NoStack}
	}

	// Emit the execution time slice.
	var stk int
	if gs.lastStopStack != tracev2.NoStack {
		stk = ctx.Stack(viewerFrames(gs.lastStopStack))
	}
	// Check invariant.
	if gs.startRunningTime == 0 {
		panic("silently broken trace or generator invariant (startRunningTime != 0) not held")
	}
	ctx.Slice(traceviewer.SliceEvent{
		Name:  gs.name(),
		Ts:    ctx.elapsed(gs.startRunningTime),
		Dur:   ts.Sub(gs.startRunningTime),
		Proc:  uint64(gs.executing),
		Stack: stk,
	})
	gs.startRunningTime = 0
	gs.lastStopStack = stack
	gs.executing = tracev2.NoProc
}

// finalize writes out any in-progress slices as if the goroutine stopped.
// This must only be used once the trace has been fully processed and no
// further events will be processed. This method may leave the gState in
// an inconsistent state.
func (gs *gState) finalize(ts tracev2.Time, ctx *traceContext) {
	if gs.executing != tracev2.NoProc {
		gs.syscallEnd(ts, false, ctx)
		gs.stop(ts, tracev2.NoStack, ctx)
	}
}

func (gs *gState) name() string {
	name := gs.label
	if gs.extraLabel != "" {
		name += " (" + gs.extraLabel + ")"
	}
	return name
}

func viewerTime(t time.Duration) float64 {
	return float64(t) / float64(time.Microsecond)
}

func goroutineName(ev tracev2.Event, st tracev2.StateTransition) string {
	if st.Stack != tracev2.NoStack {
		return lastFunc(st.Stack)
	} else if ev.Stack() != tracev2.NoStack {
		return lastFunc(ev.Stack())
	}
	return ""
}

func lastFunc(s tracev2.Stack) string {
	var last tracev2.StackFrame
	s.Frames(func(f tracev2.StackFrame) bool {
		last = f
		return true
	})
	return last.Func
}
