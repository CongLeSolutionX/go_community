// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"internal/trace"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func init() {
	http.HandleFunc("/trace", httpTrace)
	http.HandleFunc("/jsontrace", httpJsonTrace)
	http.HandleFunc("/trace_viewer_html", httpTraceViewerHTML)
}

// httpTrace serves either whole trace (goid==0) or trace for goid goroutine.
func httpTrace(w http.ResponseWriter, r *http.Request) {
	_, err := parseEvents()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	params := ""
	if goids := r.FormValue("goid"); goids != "" {
		goid, err := strconv.ParseUint(goids, 10, 64)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to parse goid parameter '%v': %v", goids, err), http.StatusInternalServerError)
			return
		}
		params = fmt.Sprintf("?goid=%v", goid)
	}
	html := strings.Replace(templTrace, "{{PARAMS}}", params, -1)
	w.Write([]byte(html))

}

var templTrace = `
<html>
	<head>
		<link href="/trace_viewer_html" rel="import">
		<script>
			document.addEventListener("DOMContentLoaded", function(event) {
				var viewer = new tv.TraceViewer('/jsontrace{{PARAMS}}');
				document.body.appendChild(viewer);
			});
		</script>
	</head>
	<body>
	</body>
</html>
`

// httpTraceViewerHTML serves static part of trace-viewer.
// This URL is queried from templTrace HTML.
func httpTraceViewerHTML(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join(os.Getenv("GOROOT"), "misc", "trace", "trace_viewer_lean.html"))
}

// httpJsonTrace serves json trace, requested from within templTrace HTML.
func httpJsonTrace(w http.ResponseWriter, r *http.Request) {
	events, err := parseEvents()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	params := &traceParams{
		traceEvents: events,
		endTime:     int64(1<<63 - 1),
	}

	if goids := r.FormValue("goid"); goids != "" {
		goid, err := strconv.ParseUint(goids, 10, 64)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to parse goid parameter '%v': %v", goids, err), http.StatusInternalServerError)
			return
		}
		analyzeGoroutines(events)
		g := gs[goid]
		params.gtrace = true
		params.startTime = g.StartTime
		params.endTime = g.EndTime
		params.maing = goid
		params.gs = relatedGoroutines(events, goid)
	}

	err = json.NewEncoder(w).Encode(generateTrace(params))
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to serialize trace: %v", err), http.StatusInternalServerError)
		return
	}
}

type traceParams struct {
	traceEvents []*trace.Event
	gtrace      bool
	startTime   int64
	endTime     int64
	maing       uint64
	gs          map[uint64]bool
}

type traceContext struct {
	*traceParams
	events    []*ViewerEvent
	arrowSeq  uint64
	heapAlloc uint64
	nextGC    uint64
	gcount    uint64
	grunnable uint64
	grunning  uint64
	insyscall uint64
	prunning  uint64
}

type ViewerEvent struct {
	Name  string      `json:"name,omitempty"`
	Phase string      `json:"ph"`
	Time  int64       `json:"ts"`
	Pid   uint64      `json:"pid"` // can omitempty?
	Tid   uint64      `json:"tid"` // can omitempty?
	Scope string      `json:"s,omitempty"`
	ID    string      `json:"id,omitempty"`
	Arg   interface{} `json:"args,omitempty"`
}

type NameArg struct {
	Name string `json:"name"`
}

type SortIndexArg struct {
	Index int `json:"sort_index"`
}

// generateTrace generates json trace for trace-viewer:
// https://github.com/google/trace-viewer
// Trace format is described at:
// https://docs.google.com/document/d/1CvAClvFfyA5R-PhYUmn5OOQtYMH4h6I0nSsKchNAySU/view
// If gtrace=true, generate trace for goroutine goid, otherwise whole trace.
// startTime, endTime determine part of the trace that we are interested in.
// gset restricts goroutines that are included in the resulting trace.
func generateTrace(params *traceParams) []*ViewerEvent {
	ctx := &traceContext{traceParams: params}
	maxProc := 0
	gnames := make(map[uint64]string)
	for _, ev := range ctx.traceEvents {
		// Handle trace.EvGoStart separately, because we need the goroutine name
		// even if ignore the event otherwise.
		if ev.Type == trace.EvGoStart {
			if _, ok := gnames[ev.G]; !ok {
				if len(ev.Stk) > 0 {
					gnames[ev.G] = fmt.Sprintf("G%v %s", ev.G, ev.Stk[0].Fn)
				} else {
					gnames[ev.G] = fmt.Sprintf("G%v", ev.G)
				}
			}
		}

		// Ignore events that are from uninteresting goroutines
		// or outside of the interesting timeframe.
		if ctx.gs != nil && ev.P < trace.FakeP && !ctx.gs[ev.G] {
			continue
		}
		if ev.Ts < ctx.startTime || ev.Ts > ctx.endTime {
			continue
		}

		if ev.P < trace.FakeP && ev.P > maxProc {
			maxProc = ev.P
		}

		switch ev.Type {
		case trace.EvProcStart:
			if ctx.gtrace {
				continue
			}
			ctx.prunning++
			ctx.emitThreadCounters(ev)
			ctx.emitInstant(ev, "proc start")
		case trace.EvProcStop:
			if ctx.gtrace {
				continue
			}
			ctx.prunning--
			ctx.emitThreadCounters(ev)
			ctx.emitInstant(ev, "proc stop")
		case trace.EvGCStart:
			ctx.emitBegin(ev, "GC")
		case trace.EvGCDone:
			ctx.emitEnd(ev)
		case trace.EvGCScanStart:
			if ctx.gtrace {
				continue
			}
			ctx.emitBegin(ev, "MARK")
		case trace.EvGCScanDone:
			ctx.emitEnd(ev)
		case trace.EvGCSweepStart:
			ctx.emitBegin(ev, "SWEEP")
		case trace.EvGCSweepDone:
			ctx.emitEnd(ev)
		case trace.EvGoStart:
			ctx.grunnable--
			ctx.grunning++
			ctx.emitGoroutineCounters(ev)
			ctx.emitBegin(ev, gnames[ev.G])
		case trace.EvGoCreate:
			ctx.gcount++
			ctx.grunnable++
			ctx.emitGoroutineCounters(ev)
			ctx.emitArrow(ev, "go")
		case trace.EvGoEnd:
			ctx.gcount--
			ctx.grunning--
			ctx.emitGoroutineCounters(ev)
			ctx.emitEnd(ev)
		case trace.EvGoUnblock:
			ctx.grunnable++
			ctx.emitGoroutineCounters(ev)
			ctx.emitArrow(ev, "unblock")
		case trace.EvGoSysCall:
			ctx.emitInstant(ev, "syscall")
		case trace.EvGoSysExit:
			ctx.grunnable++
			ctx.emitGoroutineCounters(ev)
			ctx.insyscall--
			ctx.emitThreadCounters(ev)
			ctx.emitArrow(ev, "sysexit")
		case trace.EvGoSysBlock:
			ctx.grunning--
			ctx.emitGoroutineCounters(ev)
			ctx.insyscall++
			ctx.emitThreadCounters(ev)
			ctx.emitEnd(ev)
		case trace.EvGoSched, trace.EvGoPreempt:
			ctx.grunnable++
			ctx.grunning--
			ctx.emitGoroutineCounters(ev)
			ctx.emitEnd(ev)
		case trace.EvGoStop,
			trace.EvGoSleep, trace.EvGoBlock, trace.EvGoBlockSend, trace.EvGoBlockRecv,
			trace.EvGoBlockSelect, trace.EvGoBlockSync, trace.EvGoBlockCond, trace.EvGoBlockNet:
			ctx.grunning--
			ctx.emitGoroutineCounters(ev)
			ctx.emitEnd(ev)
		case trace.EvGoWaiting:
			ctx.grunnable--
			ctx.emitGoroutineCounters(ev)
		case trace.EvGoInSyscall:
			ctx.insyscall++
			ctx.emitThreadCounters(ev)
		case trace.EvHeapAlloc:
			ctx.heapAlloc = ev.Args[0]
			ctx.emitHeapCounters(ev)
		case trace.EvNextGC:
			ctx.nextGC = ev.Args[0]
			ctx.emitHeapCounters(ev)
		}
	}

	ctx.emit(&ViewerEvent{Name: "process_name", Phase: "M", Pid: 0, Arg: &NameArg{"PROCS"}})
	ctx.emit(&ViewerEvent{Name: "process_sort_index", Phase: "M", Pid: 0, Arg: &SortIndexArg{1}})

	ctx.emit(&ViewerEvent{Name: "process_name", Phase: "M", Pid: 1, Arg: &NameArg{"STATS"}})
	ctx.emit(&ViewerEvent{Name: "process_sort_index", Phase: "M", Pid: 1, Arg: &SortIndexArg{0}})

	ctx.emit(&ViewerEvent{Name: "thread_name", Phase: "M", Pid: 0, Tid: trace.NetpollP, Arg: &NameArg{"Network"}})
	ctx.emit(&ViewerEvent{Name: "thread_sort_index", Phase: "M", Pid: 0, Tid: trace.NetpollP, Arg: &SortIndexArg{-5}})

	ctx.emit(&ViewerEvent{Name: "thread_name", Phase: "M", Pid: 0, Tid: trace.TimerP, Arg: &NameArg{"Timers"}})
	ctx.emit(&ViewerEvent{Name: "thread_sort_index", Phase: "M", Pid: 0, Tid: trace.TimerP, Arg: &SortIndexArg{-4}})

	ctx.emit(&ViewerEvent{Name: "thread_name", Phase: "M", Pid: 0, Tid: trace.SyscallP, Arg: &NameArg{"Syscalls"}})
	ctx.emit(&ViewerEvent{Name: "thread_sort_index", Phase: "M", Pid: 0, Tid: trace.SyscallP, Arg: &SortIndexArg{-3}})

	if !ctx.gtrace {
		for i := 0; i <= maxProc; i++ {
			ctx.emit(&ViewerEvent{Name: "thread_name", Phase: "M", Pid: 0, Tid: uint64(i), Arg: &NameArg{fmt.Sprintf("Proc %v", i)}})
		}
	}

	if ctx.gtrace && ctx.gs != nil {
		for k, v := range gnames {
			if !ctx.gs[k] {
				continue
			}
			ctx.emit(&ViewerEvent{Name: "thread_name", Phase: "M", Pid: 0, Tid: k, Arg: &NameArg{v}})
		}
		ctx.emit(&ViewerEvent{Name: "thread_sort_index", Phase: "M", Pid: 0, Tid: ctx.maing, Arg: &SortIndexArg{-2}})
		ctx.emit(&ViewerEvent{Name: "thread_sort_index", Phase: "M", Pid: 0, Tid: 0, Arg: &SortIndexArg{-1}})
	}

	return ctx.events
}

func (ctx *traceContext) emit(e *ViewerEvent) {
	ctx.events = append(ctx.events, e)
}

func (ctx *traceContext) time(ev *trace.Event) int64 {
	if ev.Ts < ctx.startTime || ev.Ts > ctx.endTime {
		fmt.Printf("ts=%v startTime=%v endTime\n", ev.Ts, ctx.startTime, ctx.endTime)
		panic("timestamp is outside of trace range")
	}
	return ev.Ts - ctx.startTime
}

func (ctx *traceContext) proc(ev *trace.Event) uint64 {
	if ctx.gtrace && ev.P < trace.FakeP {
		return ev.G
	} else {
		return uint64(ev.P)
	}
}

func (ctx *traceContext) emitBegin(ev *trace.Event, name string) {
	ctx.emit(&ViewerEvent{Name: name, Phase: "B", Time: ctx.time(ev), Tid: ctx.proc(ev)})
}

func (ctx *traceContext) emitEnd(ev *trace.Event) {
	ctx.emit(&ViewerEvent{Phase: "E", Time: ctx.time(ev), Tid: ctx.proc(ev), Arg: formatStack(ev.Stk)})
}

func (ctx *traceContext) emitHeapCounters(ev *trace.Event) {
	type Arg struct {
		Allocated uint64
		NextGC    uint64
	}
	if ctx.gtrace {
		return
	}
	diff := uint64(0)
	if ctx.nextGC > ctx.heapAlloc {
		diff = ctx.nextGC - ctx.heapAlloc
	}
	ctx.emit(&ViewerEvent{Name: "Heap", Phase: "C", Time: ctx.time(ev), Pid: 1, Arg: &Arg{ctx.heapAlloc, diff}})
}

func (ctx *traceContext) emitGoroutineCounters(ev *trace.Event) {
	type Arg struct {
		Running  uint64
		Runnable uint64
	}
	if ctx.gtrace {
		return
	}
	ctx.emit(&ViewerEvent{Name: "Goroutines", Phase: "C", Time: ctx.time(ev), Pid: 1, Arg: &Arg{ctx.grunning, ctx.grunnable}})
}

func (ctx *traceContext) emitThreadCounters(ev *trace.Event) {
	type Arg struct {
		Running   uint64
		InSyscall uint64
	}
	if ctx.gtrace {
		return
	}
	ctx.emit(&ViewerEvent{Name: "Threads", Phase: "C", Time: ctx.time(ev), Pid: 1, Arg: &Arg{ctx.prunning, ctx.insyscall}})
}

func (ctx *traceContext) emitInstant(ev *trace.Event, name string) {
	ctx.emit(&ViewerEvent{Name: name, Phase: "I", Scope: "t", Time: ctx.time(ev), Tid: ctx.proc(ev), Arg: formatStack(ev.Stk)})
}

func (ctx *traceContext) emitArrow(ev *trace.Event, name string) {
	if ev.Link == nil {
		// The other end of the arrow is not captured in the trace.
		// For example, a goroutine was unblocked but was not scheduled before trace stop.
		return
	}
	if ctx.gtrace && (!ctx.gs[ev.Link.G] || ev.Link.Ts < ctx.startTime || ev.Link.Ts > ctx.endTime) {
		return
	}

	ctx.arrowSeq++
	id := fmt.Sprintf("%v", ctx.arrowSeq)
	ctx.emit(&ViewerEvent{Name: name, Phase: "s", Tid: ctx.proc(ev), ID: id, Time: ctx.time(ev), Arg: formatStack(ev.Stk)})
	ctx.emit(&ViewerEvent{Name: name, Phase: "t", Tid: ctx.proc(ev.Link), ID: id, Time: ctx.time(ev.Link)})
}

func formatStack(stk []*trace.Frame) interface{} {
	type Arg struct {
		Stack string `json:"stack"`
	}

	if len(stk) == 0 {
		return nil
	}
	var w bytes.Buffer
	for _, f := range stk {
		fmt.Fprintf(&w, "%s\n", f.Fn)
	}
	b := bytes.Replace(w.Bytes(), []byte(string('·')), []byte{'-'}, -1)
	return &Arg{string(b)}
}
