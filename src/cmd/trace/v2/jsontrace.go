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

func generateTrace(parsed *parsedTrace, c traceviewer.TraceConsumer) error {
	te := traceviewer.NewEmitter(c, 0, time.Duration(0), time.Duration(math.MaxInt64))
	defer te.Flush()

	gStates := map[tracev2.GoID]*gState{}
	inSyscall := map[tracev2.ProcID]*gState{}

	var maxProcID tracev2.ProcID
	// first is used to calculate the relative time of each event.
	var first tracev2.Event
	var scratchFrames []*trace.Frame
	type rangeStartKey struct {
		Name  string
		Scope tracev2.ResourceID
	}
	var startEvents = make(map[rangeStartKey]tracev2.Event)
	for _, ev := range parsed.events {
		// Find the first event to calculate the relative time of each event.
		if first.Kind() == tracev2.EventBad {
			first = ev
		}
		ts := ev.Time().Sub(first.Time())
		switch ev.Kind() {
		case tracev2.EventRangeBegin:
			key := rangeStartKey{Name: ev.Range().Name, Scope: ev.Range().Scope}
			startEvents[key] = ev
		case tracev2.EventRangeActive:
			key := rangeStartKey{Name: ev.Range().Name, Scope: ev.Range().Scope}
			if _, ok := startEvents[key]; !ok {
				startEvents[key] = ev
			}
		case tracev2.EventRangeEnd:
			key := rangeStartKey{Name: ev.Range().Name, Scope: ev.Range().Scope}
			startEv, ok := startEvents[key]
			if !ok {
				return fmt.Errorf("missing EventRangeBegin or EventRangeActive event for %q", ev.String())
			}
			delete(startEvents, key)
			te.Slice(traceviewer.SliceEvent{
				Name:     ev.Range().Name,
				Ts:       ev.Time().Sub(first.Time()),
				Dur:      ev.Time().Sub(startEv.Time()),
				Proc:     viewerProc(ev),
				Stack:    te.Stack(eventFrames(startEv, &scratchFrames)),
				EndStack: te.Stack(eventFrames(ev, &scratchFrames)),
			})
		case tracev2.EventMetric:
			m := ev.Metric()
			switch m.Name {
			case "/memory/classes/heap/objects:bytes":
				te.HeapAlloc(ts, m.Value.Uint64())
			case "/gc/heap/goal:bytes":
				te.HeapGoal(ts, m.Value.Uint64())
			case "/sched/gomaxprocs:threads":
				te.Gomaxprocs(m.Value.Uint64())
			}
		case tracev2.EventStateTransition:
			st := ev.StateTransition()
			switch st.Resource.Kind {
			case tracev2.ResourceProc:
				proc := st.Resource.Proc()
				viewerEv := traceviewer.InstantEvent{
					Ts:    ev.Time().Sub(first.Time()),
					Proc:  uint64(proc),
					Stack: te.Stack(eventFrames(ev, &scratchFrames)),
				}

				from, to := st.Proc()
				if to == tracev2.ProcRunning {
					viewerEv.Name = "proc start"
					viewerEv.Arg = format.ThreadIDArg{ThreadID: uint64(ev.Thread())}
					te.IncThreadStateCount(ts, traceviewer.ThreadStateRunning, 1)
				}
				if from == tracev2.ProcRunning {
					viewerEv.Name = "proc stop"
					te.IncThreadStateCount(ts, traceviewer.ThreadStateRunning, -1)

					// Check if this proc was in a syscall before it stopped.
					// This means the syscall blocked. We need to emit it to the
					// viewer at this point because we only display the time the
					// syscall occupied a P when the viewer is in P-mode.
					gs, ok := inSyscall[proc]
					if ok {
						te.Slice(traceviewer.SliceEvent{
							Name:  "syscall",
							Ts:    gs.startSyscall.Time().Sub(first.Time()),
							Dur:   ev.Time().Sub(gs.startSyscall.Time()),
							Proc:  uint64(proc),
							Stack: te.Stack(eventFrames(gs.startSyscall, &scratchFrames)),
							Arg:   format.BlockedArg{Blocked: "yes"},
						})
						delete(inSyscall, proc)

						te.Slice(traceviewer.SliceEvent{
							Name:  gs.label,
							Ts:    gs.startRunning.Time().Sub(first.Time()),
							Dur:   ev.Time().Sub(gs.startRunning.Time()),
							Proc:  uint64(proc),
							Stack: te.Stack(eventFrames(gs.startRunning, &scratchFrames)),
						})
						gs.startRunning = tracev2.Event{}
					}
				}
				// TODO(mknyszek): Consider modeling procs differently and have them be
				// transition to and from NotExist when GOMAXPROCS changes. We can emit
				// events for this to clearly delineate GOMAXPROCS changes.

				if viewerEv.Name != "" {
					te.Instant(viewerEv)
				}

			case tracev2.ResourceGoroutine:
				if ev.Proc() > maxProcID {
					maxProcID = ev.Proc()
				}
				goID := st.Resource.Goroutine()

				// If we haven't seen this goroutine before, create a new
				// gState for it.
				gs, ok := gStates[goID]
				if !ok {
					gs = &gState{label: fmt.Sprintf("G%d", goID)}
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

				// See assumed Go state machine: https://bit.ly/49quCCy
				// The one described in the design doc doesn't seem to be
				// up-to-date anymore.
				from, to := st.Goroutine()

				if (from == tracev2.GoSyscall && to == tracev2.GoWaiting) ||
					(from == tracev2.GoRunning && to != tracev2.GoSyscall) {
					if gs.startRunning.Kind() != tracev2.EventBad {
						// For Syscall -> Waiting and Running -> !Syscall
						// transitions, emit a slice event for the time the
						// goroutine was running unless.
						te.Slice(traceviewer.SliceEvent{
							Name:     gs.label,
							Ts:       gs.startRunning.Time().Sub(first.Time()),
							Dur:      ev.Time().Sub(gs.startRunning.Time()),
							Proc:     uint64(ev.Proc()),
							Stack:    te.Stack(eventFrames(gs.startRunning, &scratchFrames)),
							EndStack: te.Stack(eventFrames(ev, &scratchFrames)),
						})
						gs.startRunning = tracev2.Event{}
					} else if from != tracev2.GoSyscall {
						// Sanity check: We should not end up here unless we're
						// coming out of a blocked syscall.
						return fmt.Errorf("unexpected state transition from %s to %s", from.String(), to.String())
					}
				}

				_, didNotBlock := inSyscall[ev.Proc()]
				if from == tracev2.GoSyscall && didNotBlock {
					te.Slice(traceviewer.SliceEvent{
						Name:     "syscall",
						Ts:       gs.startSyscall.Time().Sub(first.Time()),
						Dur:      ev.Time().Sub(gs.startSyscall.Time()),
						Proc:     uint64(ev.Proc()),
						Stack:    te.Stack(eventFrames(gs.startSyscall, &scratchFrames)),
						EndStack: te.Stack(eventFrames(ev, &scratchFrames)),
						Arg:      format.BlockedArg{Blocked: "no"},
					})
					delete(inSyscall, ev.Proc())
				}

				if to == tracev2.GoRunning && gs.startRunning.Kind() == tracev2.EventBad {
					gStates[goID].startRunning = ev
				} else if to == tracev2.GoSyscall {
					gStates[goID].startSyscall = ev
					inSyscall[ev.Proc()] = gs
				}

				key := rangeStartKey{Name: "GC mark assist", Scope: tracev2.MakeResourceID(goID)}
				_, inMarkAssist := startEvents[key]
				te.GoroutineTransition(ts, viewerGState(from, inMarkAssist), viewerGState(to, inMarkAssist))
			}
		}
	}
	return nil
}

func viewerProc(ev tracev2.Event) uint64 {
	if ev.Range().Name == "GC concurrent mark phase" {
		return trace.GCP
	}
	// N.B. STW events end up in the lane of the proc emitting the event.
	//
	// This may seem weird at first because STW is a global property of the program,
	// but in reality the STW must always be triggered by *something*, and it's not
	// possible for a bare thread to STW currently. We may want this to end up in
	// a separate lane, but for now ending up in a P's lane is fine.
	return uint64(ev.Proc())
}

// eventFrames returns the frames of the stack of ev. The given frame slice is
// used to store the frames to reduce allocations.
//
// TODO(mknyszek): Passing the slice in might be a premature optimization.
func eventFrames(ev tracev2.Event, frames *[]*trace.Frame) []*trace.Frame {
	n := 0
	ev.Stack().Frames(func(f tracev2.StackFrame) bool {
		frame := trace.Frame{
			PC:   f.PC,
			Fn:   f.Func,
			File: f.File,
			Line: int(f.Line),
		}
		if len(*frames) <= n {
			*frames = append(*frames, &frame)
		} else {
			*(*frames)[n] = frame
		}
		n++
		return true
	})
	return (*frames)[0:n]
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
	label     string
	named     bool
	isSystemG bool

	// startRunning is the most recent event that caused a goroutine to
	// transition to GoRunning.
	startRunning tracev2.Event
	// startSyscall is the most recent event that caused a goroutine to
	// transition to GoSyscall.
	startSyscall tracev2.Event
}

func (g gState) GoroutineName() string {
	return ""
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
