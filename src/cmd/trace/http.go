// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"internal/trace"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

func init() {
	http.Handle("/", http.HandlerFunc(httpMain))
	http.Handle("/trace", http.HandlerFunc(httpTrace))
	http.Handle("/io", http.HandlerFunc(httpIO))
	http.Handle("/block", http.HandlerFunc(httpBlock))
	http.Handle("/syscall", http.HandlerFunc(httpSyscall))
	http.Handle("/sched", http.HandlerFunc(httpSched))
}

// httpMain serves the starting page.
func httpMain(w http.ResponseWriter, r *http.Request) {
	templMain.Execute(w, nil)
}

var templMain = template.Must(template.New("").Parse(`
<html>
<body>
<a href="/trace">View trace</a><br>
<a href="/goroutines">Goroutine analysis</a><br>
<a href="/io">IO blocking profile</a><br>
<a href="/block">Synchronization blocking profile</a><br>
<a href="/syscall">Syscall blocking profile</a><br>
<a href="/sched">Scheduler latency profile</a><br>
</body>
</html>
`))

// httpTrace serves the whole trace.
func httpTrace(w http.ResponseWriter, r *http.Request) {
	httpTraceImpl(w, r, false, 0, 1<<63-1, 0, nil)
}

// httpTraceImpl serves either whole trace or trace for a particular goroutine.
// If gtrace=true, serve trace for goroutine goid, otherwise whole trace.
// startTime, endTime determine part of the trace that we are interested in.
// gset restricts goroutines that are included in the resulting trace.
func httpTraceImpl(w http.ResponseWriter, r *http.Request, gtrace bool, startTime, endTime int64, goid uint64, gset map[uint64]bool) {
	tracef, err := ioutil.TempFile("", "trace")
	if err != nil {
		fmt.Fprintf(w, "failed to create temp file: %v\n", err)
		return
	}
	traceb := bufio.NewWriter(tracef)
	generateTrace(traceb, events, gtrace, startTime, endTime, goid, gset)
	traceb.Flush()
	tracef.Close()
	defer func() {
		os.Remove(tracef.Name())
	}()

	htmlFilename := tracef.Name() + ".html"
	_, err = exec.Command("trace2html", "--output="+htmlFilename, tracef.Name()).CombinedOutput()
	if err != nil {
		fmt.Fprintf(w, "failed to execute trace2html: %v\n", err)
		return
	}

	htmlf, err := os.Open(htmlFilename)
	if err != nil {
		fmt.Fprintf(w, "failed to open html file: %v\n", err)
		return
	}
	io.Copy(w, htmlf)
	htmlf.Close()
	os.Remove(htmlf.Name())
}

// Record represents one entry in pprof-like profiles.
type Record struct {
	stk  []*trace.Frame
	n    uint64
	time int64
}

// httpIO serves IO pprof-like profile (time spent in IO wait).
func httpIO(w http.ResponseWriter, r *http.Request) {
	prof := make(map[uint64]Record)
	for _, ev := range events {
		if ev.Type != trace.EvGoBlockNet || ev.Link == nil || ev.StkID == 0 || len(ev.Stk) == 0 {
			continue
		}
		rec := prof[ev.StkID]
		rec.stk = ev.Stk
		rec.n++
		rec.time += ev.Link.Ts - ev.Ts
		prof[ev.StkID] = rec
	}
	serveSVGProfile(w, r, prof)
}

// httpBlock serves blocking pprof-like profile (time spent blocked on synchronization primitives).
func httpBlock(w http.ResponseWriter, r *http.Request) {
	prof := make(map[uint64]Record)
	for _, ev := range events {
		switch ev.Type {
		case trace.EvGoBlockSend, trace.EvGoBlockRecv, trace.EvGoBlockSelect,
			trace.EvGoBlockSync, trace.EvGoBlockCond:
		default:
			continue
		}
		if ev.Link == nil || ev.StkID == 0 || len(ev.Stk) == 0 {
			continue
		}
		rec := prof[ev.StkID]
		rec.stk = ev.Stk
		rec.n++
		rec.time += ev.Link.Ts - ev.Ts
		prof[ev.StkID] = rec
	}
	serveSVGProfile(w, r, prof)
}

// httpSyscall serves syscall pprof-like profile (time spent blocked in syscalls).
func httpSyscall(w http.ResponseWriter, r *http.Request) {
	prof := make(map[uint64]Record)
	for _, ev := range events {
		if ev.Type != trace.EvGoSysCall || ev.Link == nil || ev.StkID == 0 || len(ev.Stk) == 0 {
			continue
		}
		rec := prof[ev.StkID]
		rec.stk = ev.Stk
		rec.n++
		rec.time += ev.Link.Ts - ev.Ts
		prof[ev.StkID] = rec
	}
	serveSVGProfile(w, r, prof)
}

// httpSched serves scheduler latency pprof-like profile
// (time between a goroutine become runnable and actually scheduled for execution).
func httpSched(w http.ResponseWriter, r *http.Request) {
	prof := make(map[uint64]Record)
	for _, ev := range events {
		if (ev.Type != trace.EvGoUnblock && ev.Type != trace.EvGoCreate) ||
			ev.Link == nil || ev.StkID == 0 || len(ev.Stk) == 0 {
			continue
		}
		rec := prof[ev.StkID]
		rec.stk = ev.Stk
		rec.n++
		rec.time += ev.Link.Ts - ev.Ts
		prof[ev.StkID] = rec
	}
	serveSVGProfile(w, r, prof)
}

// serveSVGProfile serves pprof-like profile stored in prof.
func serveSVGProfile(w http.ResponseWriter, r *http.Request, prof map[uint64]Record) {
	blockf, err := ioutil.TempFile("", "block")
	if err != nil {
		fmt.Fprintf(w, "failed to create temp file: %v\n", err)
		return
	}
	blockb := bufio.NewWriter(blockf)
	fmt.Fprintf(blockb, "--- contention:\ncycles/second=1000000000\n")
	for _, rec := range prof {
		fmt.Fprintf(blockb, "%v %v @", rec.time, rec.n)
		for _, f := range rec.stk {
			fmt.Fprintf(blockb, " 0x%x", f.PC)
		}
		fmt.Fprintf(blockb, "\n")
	}
	blockb.Flush()
	blockf.Close()
	defer func() {
		os.Remove(blockf.Name())
	}()

	svgFilename := blockf.Name() + ".svg"
	_, err = exec.Command("go", "tool", "pprof", "-svg", "-output", svgFilename, bin, blockf.Name()).CombinedOutput()
	if err != nil {
		fmt.Fprintf(w, "failed to execute go tool pprof: %v\n", err)
		return
	}

	svgf, err := os.Open(svgFilename)
	if err != nil {
		fmt.Fprintf(w, "failed to open svg file: %v\n", err)
		return
	}
	w.Header().Set("Content-Type", "image/svg+xml")
	io.Copy(w, svgf)
	svgf.Close()
	os.Remove(svgf.Name())
}

type traceContext struct {
	w         io.Writer
	gtrace    bool
	startTime int64
	endTime   int64
	maing     uint64
	gs        map[uint64]bool
	seq       uint64
}

// generateTrace generates json trace for trace2html program.
// If gtrace=true, serve trace for goroutine goid, otherwise whole trace.
// startTime, endTime determine part of the trace that we are interested in.
// gset restricts goroutines that are included in the resulting trace.
func generateTrace(w io.Writer, events []*trace.Event, gtrace bool, startTime, endTime int64, maing uint64, gs map[uint64]bool) {
	ctx := &traceContext{
		w:         w,
		gtrace:    gtrace,
		startTime: startTime,
		endTime:   endTime,
		maing:     maing,
		gs:        gs,
	}
	w.Write([]byte("{\"trace.Events\": [\n"))

	var gcount, grunnable, grunning, insyscall, prunning uint64
	gnames := make(map[uint64]string)
	for _, ev := range events {
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
		if gs != nil && ev.P < trace.NetpollP && !gs[ev.G] {
			continue
		}
		if ev.Ts < startTime || ev.Ts > endTime {
			continue
		}

		switch ev.Type {
		case trace.EvProcStart:
			if gtrace {
				continue
			}
			prunning++
			emitCounter(ctx, ev, "procs running", prunning)
			emitInstant(ctx, ev, "proc start")
		case trace.EvProcStop:
			if gtrace {
				continue
			}
			prunning--
			emitCounter(ctx, ev, "procs running", prunning)
			emitInstant(ctx, ev, "proc stop")
		case trace.EvGCStart:
			emitBegin(ctx, ev, "GC")
		case trace.EvGCDone:
			emitEnd(ctx, ev)
		case trace.EvGCScanStart:
			if gtrace {
				continue
			}
			emitBegin(ctx, ev, "MARK")
		case trace.EvGCScanDone:
			emitEnd(ctx, ev)
		case trace.EvGCSweepStart:
			emitBegin(ctx, ev, "SWEEP")
		case trace.EvGCSweepDone:
			emitEnd(ctx, ev)
		case trace.EvGoStart:
			grunnable--
			emitCounter(ctx, ev, "runqueue", grunnable)
			grunning++
			emitCounter(ctx, ev, "running", grunning)
			emitBegin(ctx, ev, gnames[ev.G])
		case trace.EvGoCreate:
			gcount++
			emitCounter(ctx, ev, "goroutines", gcount)
			grunnable++
			emitCounter(ctx, ev, "runqueue", grunnable)
			emitArrow(ctx, ev, "go")
		case trace.EvGoEnd:
			gcount--
			emitCounter(ctx, ev, "goroutines", gcount)
			grunning--
			emitCounter(ctx, ev, "running", grunning)
			emitEnd(ctx, ev)
		case trace.EvGoUnblock:
			grunnable++
			emitCounter(ctx, ev, "runqueue", grunnable)
			emitArrow(ctx, ev, "unblock")
		case trace.EvGoSysCall:
			emitInstant(ctx, ev, "syscall")
		case trace.EvGoSysExit:
			grunnable++
			emitCounter(ctx, ev, "runqueue", grunnable)
			insyscall--
			emitCounter(ctx, ev, "threads in syscalls", insyscall)
			emitArrow(ctx, ev, "sysexit")
		case trace.EvGoSysBlock:
			grunning--
			emitCounter(ctx, ev, "running", grunning)
			insyscall++
			emitCounter(ctx, ev, "threads in syscalls", insyscall)
			emitEnd(ctx, ev)
		case trace.EvGoSched, trace.EvGoPreempt:
			grunnable++
			emitCounter(ctx, ev, "runqueue", grunnable)
			grunning--
			emitCounter(ctx, ev, "running", grunning)
			emitEnd(ctx, ev)
		case trace.EvGoStop,
			trace.EvGoSleep, trace.EvGoBlock, trace.EvGoBlockSend, trace.EvGoBlockRecv,
			trace.EvGoBlockSelect, trace.EvGoBlockSync, trace.EvGoBlockCond, trace.EvGoBlockNet:
			grunning--
			emitCounter(ctx, ev, "running", grunning)
			emitEnd(ctx, ev)
		case trace.EvGoWaiting:
			grunnable--
			emitCounter(ctx, ev, "running", grunning)
		case trace.EvGoInSyscall:
			insyscall++
			emitCounter(ctx, ev, "threads in syscalls", insyscall)
		case trace.EvHeapAlloc:
			emitCounter(ctx, ev, "heap_alloc", ev.Args[0])
		case trace.EvNextGC:
			emitCounter(ctx, ev, "next_gc", ev.Args[0])
		}
	}
	fmt.Fprintf(w, `{"name": "process_name", "ph": "M", "pid": 0, "args": { "name" : "PROCS" }},`)
	fmt.Fprintf(w, `{"name": "process_sort_index", "ph": "M", "pid": 0, "args": { "sort_index" : 1 }},%v`, "\n")
	fmt.Fprintf(w, `{"name": "process_name", "ph": "M", "pid": 1, "args": { "name" : "STATS" }},`)
	fmt.Fprintf(w, `{"name": "process_sort_index", "ph": "M", "pid": 1, "args": { "sort_index" : 0 }},%v`, "\n")

	fmt.Fprintf(w, `{"name": "thread_name", "ph": "M", "pid": 0, "tid": %v, "args": { "name" : "network" }},%v`, trace.NetpollP, "\n")
	fmt.Fprintf(w, `{"name": "thread_sort_index", "ph": "M", "pid": 0, "tid": %v, "args": { "sort_index" : -5 }},%v`, trace.NetpollP, "\n")

	fmt.Fprintf(w, `{"name": "thread_name", "ph": "M", "pid": 0, "tid": %v, "args": { "name" : "timers" }},%v`, trace.TimerP, "\n")
	fmt.Fprintf(w, `{"name": "thread_sort_index", "ph": "M", "pid": 0, "tid": %v, "args": { "sort_index" : -4 }},%v`, trace.TimerP, "\n")

	fmt.Fprintf(w, `{"name": "thread_name", "ph": "M", "pid": 0, "tid": %v, "args": { "name" : "syscalls" }},%v`, trace.SyscallP, "\n")
	fmt.Fprintf(w, `{"name": "thread_sort_index", "ph": "M", "pid": 0, "tid": %v, "args": { "sort_index" : -3 }},%v`, trace.SyscallP, "\n")

	if gtrace && gs != nil {
		for k, v := range gnames {
			if !gs[k] {
				continue
			}
			fmt.Fprintf(w, `{"name": "thread_name", "ph": "M", "pid": 0, "tid": %v, "args": { "name" : "%v" }},%v`, k, v, "\n")
		}
		fmt.Fprintf(w, `{"name": "thread_sort_index", "ph": "M", "pid": 0, "tid": %v, "args": { "sort_index" : -2 }},%v`, maing, "\n")
		fmt.Fprintf(w, `{"name": "thread_sort_index", "ph": "M", "pid": 0, "tid": 0, "args": { "sort_index" : -1 }},%v`, "\n")
	}

	w.Write([]byte("{}]}\n"))
}

func emitBegin(ctx *traceContext, ev *trace.Event, name string) {
	fmt.Fprintf(ctx.w, `{"name": "%v", "ph": "B", "ts": %v, "pid": 0, "tid": %v},%v`,
		name, traceTime(ctx, ev), traceProc(ctx, ev), "\n")
}

func emitEnd(ctx *traceContext, ev *trace.Event) {
	fmt.Fprintf(ctx.w, `{"ph": "E", "ts": %v, "pid": 0, "tid": %v%v},%v`,
		traceTime(ctx, ev), traceProc(ctx, ev), formatStack("end stack", ev.Stk), "\n")
}

func emitCounter(ctx *traceContext, ev *trace.Event, name string, val uint64) {
	if ctx.gtrace {
		return
	}
	fmt.Fprintf(ctx.w, `{"name": "%v", "ph": "C", "ts": %v, "pid": 1, "args": {"v": %v}},%v`,
		name, traceTime(ctx, ev), val, "\n")
}

func emitInstant(ctx *traceContext, ev *trace.Event, name string) {
	fmt.Fprintf(ctx.w, `{"name": "%v", "ph": "I", "s": "t", "ts": %v, "pid": 0, "tid": %v%v},%v`,
		name, traceTime(ctx, ev), traceProc(ctx, ev), formatStack("stack", ev.Stk), "\n")
}

func emitArrow(ctx *traceContext, ev *trace.Event, name string) {
	if ev.Link == nil {
		// The other end of the arrow is not captured in the trace.
		// For example, a goroutine was unblocked but was not scheduled before trace stop.
		return
	}
	if ctx.gtrace && (!ctx.gs[ev.Link.G] || ev.Link.Ts < ctx.startTime || ev.Link.Ts > ctx.endTime) {
		return
	}

	ctx.seq++
	id := fmt.Sprintf("%v", ctx.seq)
	fmt.Fprintf(ctx.w, `{"name": "%v", "ph": "s", "pid": 0, "tid": %v, "id": "%v", "ts": %v%v},%v`,
		name, traceProc(ctx, ev), id, traceTime(ctx, ev), formatStack("stack", ev.Stk), "\n")
	fmt.Fprintf(ctx.w, `{"name": "%v", "ph": "t", "pid": 0, "tid": %v, "id": "%v", "ts": %v},%v`,
		name, traceProc(ctx, ev.Link), id, traceTime(ctx, ev.Link), "\n")
}

func traceTime(ctx *traceContext, ev *trace.Event) int64 {
	if ev.Ts < ctx.startTime || ev.Ts > ctx.endTime {
		fmt.Printf("ts=%v startTime=%v endTime\n", ev.Ts, ctx.startTime, ctx.endTime)
		panic("timestamp is outside of trace range")
	}
	return ev.Ts - ctx.startTime
}

func traceProc(ctx *traceContext, ev *trace.Event) uint64 {
	if ctx.gtrace && ev.P < trace.NetpollP {
		return ev.G
	} else {
		return uint64(ev.P)
	}
}

func formatStack(name string, stk []*trace.Frame) string {
	if len(stk) <= 1 {
		return ""
	}
	var w bytes.Buffer
	for i, f := range stk {
		if i == len(stk)-1 {
			break
		}
		fmt.Fprintf(&w, "%s\n", f.Fn)
	}
	str := w.String()
	str = strings.Replace(str, "·", "-", -1)
	b, _ := json.Marshal(str[:len(str)-1])
	return fmt.Sprintf(`, "args": { "%v": %v} `, name, string(b))
}
