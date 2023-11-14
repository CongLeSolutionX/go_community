// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package trace

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"runtime"
	"runtime/trace"
	"slices"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"internal/trace/traceviewer/format"
)

func TestJSONTraceHandler(t *testing.T) {
	parsed, done := exampleTrace(t)
	defer done()

	data := recordJSONTraceHandlerResponse(t, parsed)
	checkOneGoroutinePerProc(t, data)
	checkExecutionTimes(t, data)
	checkPlausibleHeapMetrics(t, data)
	// @TODO check for plausible thread and goroutine metrics
	checkMetaNamesEmitted(t, data, "process_name", []string{"STATS", "PROCS"})
	checkMetaNamesEmitted(t, data, "thread_name", []string{"GC", "Network", "Timers", "Syscalls", "Proc 0"})
	checkProcStartStop(t, data)
	checkSyscalls(t, data)
}

func checkSyscalls(t *testing.T, data format.Data) {
	data = filterViewerTrace(data,
		filterEventName("syscall"),
		filterStackRootFunc("cmd/trace/v2.blockingSyscall"))
	if len(data.Events) <= 1 {
		t.Errorf("got %d events, want > 1", len(data.Events))
	}
	data = filterViewerTrace(data, filterBlocked("yes"))
	if len(data.Events) != 1 {
		t.Errorf("got %d events, want 1", len(data.Events))
	}
}

type eventFilterFn func(*format.Event, *format.Data) bool

func filterEventName(name string) eventFilterFn {
	return func(e *format.Event, _ *format.Data) bool {
		return e.Name == name
	}
}

// filterGoRoutineName returns an event filter that returns true if the event's
// goroutine name is equal to name.
func filterGoRoutineName(name string) eventFilterFn {
	return func(e *format.Event, _ *format.Data) bool {
		return parseGoroutineName(e) == name
	}
}

// parseGoroutineName returns the goroutine name from the event's name field.
// E.g. if e.Name is "G42 cmd/trace/v2.cpu10", this returns "cmd/trace/v2.cpu10".
func parseGoroutineName(e *format.Event) string {
	parts := strings.SplitN(e.Name, " ", 2)
	if len(parts) != 2 || !strings.HasPrefix(parts[0], "G") {
		return ""
	}
	return parts[1]
}

// filterBlocked returns an event filter that returns true if the event's
// "blocked" argument is equal to blocked.
func filterBlocked(blocked string) eventFilterFn {
	return func(e *format.Event, _ *format.Data) bool {
		m, ok := e.Arg.(map[string]any)
		if !ok {
			return false
		}
		return m["blocked"] == blocked
	}
}

// filterStackRootFunc returns an event filter that returns true if the function
// at the root of the stack trace is named name.
func filterStackRootFunc(name string) eventFilterFn {
	return func(e *format.Event, data *format.Data) bool {
		frames := stackFrames(data, e.Stack)
		rootFrame := frames[len(frames)-1]
		return strings.HasPrefix(rootFrame, name+":")
	}
}

// filterViewerTrace returns a copy of data with only the events that pass all
// of the given filters.
func filterViewerTrace(data format.Data, fns ...eventFilterFn) (filtered format.Data) {
	filtered = data
	filtered.Events = nil
	for _, e := range data.Events {
		keep := true
		for _, fn := range fns {
			keep = keep && fn(e, &filtered)
		}
		if keep {
			filtered.Events = append(filtered.Events, e)
		}
	}
	return
}

func stackFrames(data *format.Data, stackID int) (frames []string) {
	for {
		frame, ok := data.Frames[strconv.Itoa(stackID)]
		if !ok {
			return
		}
		frames = append(frames, frame.Name)
		stackID = frame.Parent
	}
}

func checkProcStartStop(t *testing.T, data format.Data) {
	procStarted := map[uint64]bool{}
	for _, e := range data.Events {
		if e.Name == "proc start" {
			if procStarted[e.TID] == true {
				t.Errorf("proc started twice: %d", e.TID)
			}
			procStarted[e.TID] = true
		}
		if e.Name == "proc stop" {
			if procStarted[e.TID] == false {
				t.Errorf("proc stopped twice: %d", e.TID)
			}
			procStarted[e.TID] = false
		}
	}
	if got, want := len(procStarted), runtime.GOMAXPROCS(0); got != want {
		t.Errorf("wrong number of procs started/stopped got=%d want=%d", got, want)
	}
}

func checkExecutionTimes(t *testing.T, data format.Data) {
	cpu10 := sumExecutionTime(filterViewerTrace(data, filterGoRoutineName("cmd/trace/v2.cpu10")))
	cpu20 := sumExecutionTime(filterViewerTrace(data, filterGoRoutineName("cmd/trace/v2.cpu20")))
	if cpu10 <= 0 || cpu20 <= 0 || cpu10 >= cpu20 {
		t.Errorf("bad execution times: cpu10=%v, cpu20=%v", cpu10, cpu20)
	}
}

func checkMetaNamesEmitted(t *testing.T, data format.Data, category string, want []string) {
	t.Helper()
	names := metaEventNameArgs(category, data)
	for _, wantName := range want {
		if !slices.Contains(names, wantName) {
			t.Errorf("%s: names=%v, want %q", category, names, wantName)
		}
	}
}

func metaEventNameArgs(category string, data format.Data) (names []string) {
	for _, e := range data.Events {
		if e.Name == category && e.Phase == "M" {
			names = append(names, e.Arg.(map[string]any)["name"].(string))
		}
	}
	return
}

func checkPlausibleHeapMetrics(t *testing.T, data format.Data) {
	hms := heapMetrics(data)
	var nonZeroAllocated, nonZeroNextGC bool
	for _, hm := range hms {
		if hm.Allocated > 0 {
			nonZeroAllocated = true
		}
		if hm.NextGC > 0 {
			nonZeroNextGC = true
		}
	}

	if !nonZeroAllocated {
		t.Errorf("nonZeroAllocated=%v, want true", nonZeroAllocated)
	}
	if !nonZeroNextGC {
		t.Errorf("nonZeroNextGC=%v, want true", nonZeroNextGC)
	}
}

func heapMetrics(data format.Data) (metrics []format.HeapCountersArg) {
	for _, e := range data.Events {
		if e.Phase == "C" && e.Name == "Heap" {
			j, _ := json.Marshal(e.Arg)
			var metric format.HeapCountersArg
			json.Unmarshal(j, &metric)
			metrics = append(metrics, metric)
		}
	}
	return
}

func recordJSONTraceHandlerResponse(t *testing.T, parsed parsedTrace) format.Data {
	h := JSONTraceHandler(parsed)
	recorder := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/jsontrace", nil)
	h.ServeHTTP(recorder, r)

	var data format.Data
	if err := json.Unmarshal(recorder.Body.Bytes(), &data); err != nil {
		t.Fatal(err)
	}
	return data
}

func checkOneGoroutinePerProc(t *testing.T, data format.Data) {
	//TODO(fg) implement
}

func sumExecutionTime(data format.Data) (sum time.Duration) {
	for _, e := range data.Events {
		sum += time.Duration(e.Dur) * time.Microsecond
	}
	return
}

// TODO(fg) Generate this trace in a way that is not dependent on scheduling. Or
// maybe just check it into the tree?
func exampleTrace(t *testing.T) (parsedTrace, func()) {
	t.Helper()

	// If the trace file exists, use it. This allows us to debug the same trace
	// repeatedly. If it doesn't exist, generate a new one. The latter is the
	// normal case.
	traceFilename := "exampleTrace.trace"
	raw, err := os.ReadFile(traceFilename)
	if err != nil {
		raw, err = newExampleTrace()
		if err != nil {
			t.Fatal(err)
		}
	}

	writeTrace := func() {
		if t.Failed() || os.Getenv("TEST_WRITE_TRACE") == "1" {
			t.Logf("writing %s for manual inspection", traceFilename)
			os.WriteFile(traceFilename, raw, 0644)
		}
	}
	defer writeTrace()

	parsed, err := parseTrace(raw)
	if err != nil {
		t.Fatal(err)
	}
	return parsed, writeTrace
}

func newExampleTrace() ([]byte, error) {
	var buf bytes.Buffer
	if err := trace.Start(&buf); err != nil {
		return nil, err
	}

	// checkExecutionTimes relies on this.
	var wg sync.WaitGroup
	wg.Add(2)
	go cpu10(&wg)
	go cpu20(&wg)
	wg.Wait()

	// checkHeapMetrics relies on this.
	allocHog(25 * time.Millisecond)

	// checkProcStartStop relies on this.
	var wg2 sync.WaitGroup
	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		wg2.Add(1)
		go func() {
			defer wg2.Done()
			cpuHog(50 * time.Millisecond)
		}()
	}
	wg2.Wait()

	// checkSyscalls relies on this. (TODO: Implement)
	done := make(chan error)
	go blockingSyscall(50*time.Millisecond, done)
	if err := <-done; err != nil {
		return nil, err
	}

	trace.Stop()
	return buf.Bytes(), nil
}

// blockingSyscall blocks the current goroutine for duration d in a syscall and
// sends a message to done when it is done or if the syscall failed.
func blockingSyscall(d time.Duration, done chan<- error) {
	r, w, err := os.Pipe()
	if err != nil {
		done <- err
		return
	}
	start := time.Now()
	msg := []byte("hello")
	time.AfterFunc(d, func() { w.Write(msg) })
	_, err = syscall.Read(int(r.Fd()), make([]byte, len(msg)))
	if err == nil && time.Since(start) < d {
		err = fmt.Errorf("syscall returned too early: want=%s got=%s", d, time.Since(start))
	}
	done <- err
}

func cpu10(wg *sync.WaitGroup) {
	defer wg.Done()
	cpuHog(10 * time.Millisecond)
}

func cpu20(wg *sync.WaitGroup) {
	defer wg.Done()
	cpuHog(20 * time.Millisecond)
}

func cpuHog(dt time.Duration) {
	start := time.Now()
	for i := 0; ; i++ {
		if i%1000 == 0 && time.Since(start) > dt {
			return
		}
	}
}

func allocHog(dt time.Duration) {
	start := time.Now()
	var s [][]byte
	for i := 0; ; i++ {
		if i%1000 == 0 && time.Since(start) > dt {
			return
		}
		s = append(s, make([]byte, 1024))
	}
}
