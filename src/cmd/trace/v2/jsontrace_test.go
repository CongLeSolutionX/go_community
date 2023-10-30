package trace

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"runtime/trace"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"internal/trace/traceviewer/format"
)

func TestJSONTraceHandler(t *testing.T) {
	data := recordJSONTraceHandlerResponse(t, exampleTrace(t))

	checkOneGoroutinePerProc(t, data)

	cpu10 := sumExecutionTime(data, "cmd/trace/v2.cpu10")
	cpu20 := sumExecutionTime(data, "cmd/trace/v2.cpu20")
	if cpu10 <= 0 || cpu20 <= 0 || cpu10 > cpu20 {
		t.Errorf("cpu10=%v, cpu20=%v", cpu10, cpu20)
	}

	checkPlausibleHeapMetrics(t, data)
	// @TODO check for plausible thread and goroutine metrics
	checkMetaNamesEmitted(t, data, "process_name", []string{"STATS", "PROCS"})
	checkMetaNamesEmitted(t, data, "thread_name", []string{"GC", "Network", "Timers", "Syscalls", "Proc 0"})
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

func sumExecutionTime(data format.Data, name string) (sum time.Duration) {
	for _, e := range data.Events {
		if parseGoroutineName(e) == name {
			sum += time.Duration(e.Dur) * time.Microsecond
		}
	}
	return
}

func parseGoroutineName(e *format.Event) string {
	parts := strings.SplitN(e.Name, " ", 2)
	if len(parts) != 2 || !strings.HasPrefix(parts[0], "G") {
		return ""
	}
	return parts[1]
}

// TODO(fg) Generate this trace in a way that is not dependent on scheduling. Or
// maybe just check it into the tree?
func exampleTrace(t *testing.T) parsedTrace {
	t.Helper()
	var buf bytes.Buffer
	if err := trace.Start(&buf); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go cpu10(&wg)
	go cpu20(&wg)
	wg.Wait()

	allocHog(25 * time.Millisecond)

	trace.Stop()
	parsed, err := parseTrace(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	return parsed
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
