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

func JSONTraceHandler(parsed parsedTrace) http.Handler {
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
			// TODO(fg): use log/slog?
			log.Printf("failed to generate trace: %v", err)
		}
	})
}

func generateTrace(parsed parsedTrace, c traceviewer.TraceConsumer) error {
	te := traceviewer.NewEmitter(c, 0, time.Duration(0), time.Duration(math.MaxInt64))
	defer te.Flush()

	gStates := map[tracev2.GoID]*gState{}

	var maxProcID tracev2.ProcID
	// first is used to calculate the relative time of each event.
	var first tracev2.Event
	var scratchFrames []*trace.Frame
	var startEvents = make(map[string]tracev2.Event)
	for _, ev := range parsed.events {
		// Find the first event to calculate the relative time of each event.
		if first.Kind() == tracev2.EventBad {
			first = ev
		}
		ts := ev.Time().Sub(first.Time())
		switch ev.Kind() {
		case tracev2.EventRangeBegin:
			startEvents[ev.Range().Name] = ev
		case tracev2.EventRangeActive:
			// Note: We're not handling any EventRangeActive events because we're
			// always processing all events from a trace right now, even when
			// only emitting the events for a goid/task/region/split.
		case tracev2.EventRangeEnd:
			startEv, ok := startEvents[ev.Range().Name]
			if !ok {
				return fmt.Errorf("missing start event for %q", ev.String())
			}
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
				viewerEv := traceviewer.InstantEvent{
					Ts:    ev.Time().Sub(first.Time()),
					Proc:  uint64(st.Resource.Proc()),
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
				}

				if viewerEv.Name != "" {
					// TODO(fg) Should we emit events for Undetermined -> Idle
					// proc transitions?
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
				// TODO(fg) how should we report syscall time with/without P?
				if (from == tracev2.GoSyscall && to == tracev2.GoWaiting) ||
					(from == tracev2.GoRunning && to != tracev2.GoSyscall) {
					te.Slice(traceviewer.SliceEvent{
						Name:     gs.label,
						Ts:       gs.running.Time().Sub(first.Time()),
						Dur:      ev.Time().Sub(gs.running.Time()),
						Proc:     uint64(ev.Proc()),
						Stack:    te.Stack(eventFrames(gs.running, &scratchFrames)),
						EndStack: te.Stack(eventFrames(ev, &scratchFrames)),
					})
				} else if to == tracev2.GoRunning {
					gStates[goID].running = ev
				}
				te.GoroutineTransition(ts, viewerGState(from), viewerGState(to))
			}
		}
	}
	return nil
}

func viewerProc(ev tracev2.Event) uint64 {
	if ev.Range().Name == "GC concurrent mark phase" {
		return trace.GCP
	}
	// TODO(fg) Should STW events get their own lane, or overlap on the proc
	// lanes like this: https://share.zight.com/9ZuLy6Nr ?
	return uint64(ev.Proc())
}

// eventFrames returns the frames of the stack of ev. The given frame slice is
// used to store the frames to reduce allocations.
//
// TODO(fg) I've not benchmarked this, might be premature optimization.
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

func viewerGState(state tracev2.GoState) traceviewer.GState {
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
		return traceviewer.GWaiting
	case tracev2.GoSyscall:
		// TODO(fg): maybe we should report this differently in the viewer? just
		// mimicking the old behavior for now.
		return traceviewer.GRunning
	default:
		panic(fmt.Sprintf("unknown GoState: %s", state.String()))
	}
}

type gState struct {
	label     string
	named     bool
	isSystemG bool
	// running is the most recent event that caused the goroutine to transition
	// to running.
	running tracev2.Event
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
