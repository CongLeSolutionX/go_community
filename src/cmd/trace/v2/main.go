package trace

import (
	"errors"
	"fmt"
	"internal/trace"
	"log"
	"net"
	"net/http"
	"os"
)

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
	// browser.Open(addr)

	mux := http.NewServeMux()
	mux.Handle("/", trace.MainHandler(nil))
	mux.Handle("/trace", trace.TraceHandler())
	mux.Handle("/jsontrace", JSONTraceHandler(data))
	mux.Handle("/static/", trace.StaticHandler())

	// TODO(fg) Support trace splitting using generations?
	err = http.Serve(ln, mux)
	return fmt.Errorf("failed to start http server: %w", err)

}

// TODO(fg) Also offer a way to print/debug the high level events from the
// trace.NewReader API?
func debugEvents(data []byte) error {
	// Problem: Can't import internal/trace/v2/internal/raw from this package.
	// But if I move this code into internal/trace/v2 then I can't import
	// cmd/internal/browser.
	return errors.New("not implemented yet")
	// rr, err := raw.NewReader(r)
	// if err != nil {
	// 	return err
	// }
	// for {
	// 	ev, err := rr.ReadEvent()
	// 	if err == io.EOF {
	// 		return nil
	// 	} else if err != nil {
	// 		return err
	// 	}
	// 	fmt.Printf("%s\n", ev.String())
	// }
}
