package trace

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"time"

	"internal/trace"
	"internal/trace/traceviewer"
	tracev2 "internal/trace/v2"

	"internal/trace/traceviewer/format"
)

func JSONTraceHandler(traceData []byte) http.Handler {
	// TODO:
	// - CPU Samples
	// - Syscalls
	// - Goroutine Metric
	// - Heap Metric
	// - Thread Metric

	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		start := int64(0)
		end := int64(math.MaxInt64)
		c := traceviewer.ViewerDataTraceConsumer(w, start, end)
		c.ConsumeTimeUnit("ns")
		te := traceviewer.NewEmitter(c, 0)
		defer te.Flush()

		gStates := map[tracev2.GoID]*gState{}

		r, err := tracev2.NewReader(bytes.NewReader(traceData))
		if err != nil {
			log.Printf("failed to create trace reader: %v", err)
			return
		}

		var maxProcID tracev2.ProcID
		// first is used to calculate the relative time of each event.
		var first tracev2.Event
		for {
			ev, err := r.ReadEvent()
			if err == io.EOF {
				break
			} else if err != nil {
				log.Printf("failed to read event: %v", err)
				return
			}

			// Find the first event to calculate the relative time of each event.
			if first.Kind() == tracev2.EventBad {
				first = ev
			}
			ts := ev.Time().Sub(first.Time())
			switch ev.Kind() {
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
					from, to := st.Proc()
					if to == tracev2.ProcRunning {
						te.IncThreadStateCount(ts, traceviewer.ThreadStateRunning, 1)
					}
					if from == tracev2.ProcRunning {
						te.IncThreadStateCount(ts, traceviewer.ThreadStateRunning, -1)
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

					// See assumed Go state machine: https://bit.ly/49iREv6
					// The one described in the design doc doesn't seem to be
					// up-to-date anymore.

					from, to := st.Goroutine()
					// TODO(fg) how should we report syscall time with/without P?
					if (from == tracev2.GoSyscall && to == tracev2.GoWaiting) ||
						(from == tracev2.GoRunning && to != tracev2.GoSyscall) {
						te.Event(&format.Event{
							Name:     gs.label,
							Phase:    "X",
							Time:     viewerTime(gs.running.Time().Sub(first.Time())),
							Dur:      viewerTime(ev.Time().Sub(gs.running.Time())),
							TID:      uint64(ev.Proc()),
							Stack:    0,
							EndStack: 0,
						})
					} else if to == tracev2.GoRunning {
						gStates[goID].running = ev
					}
					te.GoroutineTransition(ts, viewerGState(from), viewerGState(to))
				}
			}
		}
	})
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
	running   tracev2.Event
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
