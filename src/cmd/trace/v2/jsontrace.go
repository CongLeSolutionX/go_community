package trace

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"time"

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
		defer c.Flush()

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
			switch ev.Kind() {
			case tracev2.EventMetric:
				m := ev.Metric()
				fmt.Printf("m: %v\n", m)
			case tracev2.EventStateTransition:
				st := ev.StateTransition()
				if st.Resource.Kind == tracev2.ResourceGoroutine {
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
						}
					}

					// See assumed Go state machine: https://bit.ly/49iREv6
					// The one described in the design doc doesn't seem to be
					// up-to-date anymore.

					from, to := st.Goroutine()
					if (from == tracev2.GoSyscall && to == tracev2.GoWaiting) ||
						(from == tracev2.GoRunning && to != tracev2.GoSyscall) {
						c.ConsumeViewerEvent(&format.Event{
							Name:     gs.label,
							Phase:    "X",
							Time:     viewerTime(gs.running.Time().Sub(first.Time())),
							Dur:      viewerTime(ev.Time().Sub(gs.running.Time())),
							TID:      uint64(ev.Proc()),
							Stack:    0,
							EndStack: 0,
						}, true)
					} else if to == tracev2.GoRunning {
						gStates[goID].running = ev
					}
				}
			}
		}

		for i := 0; i <= int(maxProcID); i++ {
			c.ConsumeViewerEvent(&format.Event{
				Name:  "thread_name",
				Phase: "M",
				PID:   format.ProcsSection,
				TID:   uint64(i),
				Arg:   &format.NameArg{Name: fmt.Sprintf("Proc %v", i)},
			}, true)
			c.ConsumeViewerEvent(&format.Event{
				Name:  "thread_sort_index",
				Phase: "M",
				PID:   format.ProcsSection,
				TID:   uint64(i),
				Arg:   &format.SortIndexArg{Index: i},
			}, true)
		}
	})
}

type gState struct {
	label   string
	named   bool
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

type heapStats struct {
	heapAlloc uint64
	nextGC    uint64
}
