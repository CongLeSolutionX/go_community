package trace

import (
	"bytes"
	"fmt"
	"internal/trace/traceviewer"
	"io"
	"log"
	"net"
	"net/http"
	"os"

	"internal/trace/v2/raw"

	"cmd/internal/browser"
)

/*
TODO:

- Split traces by generation. Problem: Doesn't seem possible with the high level reader API.
- Support printing raw events (-d flag): Problem: Needs access to raw parsing API.
- ... everything else

*/

// Main is the main function for cmd/trace v2.
func Main(traceFile, httpAddr, pprof string, debug bool) error {
	data, err := os.ReadFile(traceFile)
	if err != nil {
		return fmt.Errorf("failed to read trace file: %w", err)
	}

	if debug {
		return debugEvents(data)
	}

	ln, err := net.Listen("tcp", httpAddr)
	if err != nil {
		return fmt.Errorf("failed to create server socket: %w", err)
	}

	addr := "http://" + ln.Addr().String()
	log.Printf("Opening browser. Trace viewer is listening on %s", addr)
	browser.Open(addr)

	log.Print("Splitting trace...")
	ranges, err := splitTrace(data)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.Handle("/", traceviewer.MainHandler(ranges))
	mux.Handle("/trace", traceviewer.TraceHandler())
	mux.Handle("/jsontrace", JSONTraceHandler(data))
	mux.Handle("/static/", traceviewer.StaticHandler())

	// TODO(fg) Support trace splitting using generations?
	err = http.Serve(ln, mux)
	return fmt.Errorf("failed to start http server: %w", err)

}

// splitTrace splits the trace into a number of ranges,
// each resulting in approx 100MB of json output
// (trace viewer can hardly handle more).
func splitTrace(traceData []byte) ([]traceviewer.Range, error) {
	s, c := traceviewer.SplittingTraceConsumer(100 << 20) // 100M
	if err := generateTrace(traceData, c); err != nil {
		return nil, err
	}
	return s.Ranges, nil
}

// TODO(fg) Also offer a way to print/debug the high level events from the
// trace.NewReader API?
func debugEvents(traceData []byte) error {
	rr, err := raw.NewReader(bytes.NewReader(traceData))
	if err != nil {
		return err
	}

	for {
		ev, err := rr.ReadEvent()
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}
		fmt.Printf("%s\n", ev.String())
	}
}
