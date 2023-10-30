package trace

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"runtime/trace"
	"strings"
	"sync"
	"testing"
	"time"

	"internal/trace/traceviewer/format"
)

// TestHelloWorld is used for producing simple traces for debugging the
// generated json trace.
// TODO: Remove this before merging.
func TestHelloWorld(t *testing.T) {
	go helloWorld()
	time.Sleep(10 * time.Millisecond)
}

func helloWorld() {
	fmt.Println("Hello, world!")
}

func TestJSONTraceHandler(t *testing.T) {
	data := recordJSONTraceHandlerResponse(t, exampleTrace(t))
	assertOneGoroutinePerProc(t, data)

	cpu10 := sumExecutionTime(data, "cmd/trace/v2.cpu10")
	cpu20 := sumExecutionTime(data, "cmd/trace/v2.cpu20")
	if cpu10 <= 0 || cpu20 <= 0 || cpu10 > cpu20 {
		t.Errorf("cpu10=%v, cpu20=%v", cpu10, cpu20)
	}
}

func recordJSONTraceHandlerResponse(t *testing.T, traceData []byte) format.Data {
	h := JSONTraceHandler(traceData)
	recorder := httptest.NewRecorder()
	h.ServeHTTP(recorder, nil)

	var data format.Data
	if err := json.Unmarshal(recorder.Body.Bytes(), &data); err != nil {
		t.Fatal(err)
	}
	return data
}

func assertOneGoroutinePerProc(t *testing.T, data format.Data) {
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
func exampleTrace(t *testing.T) []byte {
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
	return buf.Bytes()
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
