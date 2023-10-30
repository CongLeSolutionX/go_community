package trace

import (
	"fmt"
	"net/http/httptest"
	"os"
	"testing"
)

func TestHelloWorld(t *testing.T) {
	// TODO: Remove this. Just used to show multiple Gs running on one P issue.
	fmt.Println("Hello, world!")
}

func TestJSONTraceHandler(t *testing.T) {
	data, err := os.ReadFile("../../../hello-world.v2.trace")
	if err != nil {
		t.Fatal(err)
	}
	h := JSONTraceHandler(data)
	recorder := httptest.NewRecorder()
	h.ServeHTTP(recorder, nil)

	// TODO: Validate the output.
	fmt.Printf("recorder.Body.String(): %v\n", recorder.Body.String())
}
