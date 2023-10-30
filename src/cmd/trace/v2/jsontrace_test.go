package trace

import (
	"fmt"
	"net/http/httptest"
	"os"
	"testing"
	"time"
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
	data, err := os.ReadFile("../../../hello.trace")
	if err != nil {
		t.Fatal(err)
	}
	h := JSONTraceHandler(data)
	recorder := httptest.NewRecorder()
	h.ServeHTTP(recorder, nil)

	// TODO: Validate the output.
	fmt.Printf("recorder.Body.String(): %v\n", recorder.Body.String())
}
