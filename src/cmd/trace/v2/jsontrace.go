package trace

import (
	"bytes"
	"fmt"
	tracev2 "internal/trace/v2"
	"io"
	"log"
	"net/http"
)

func JSONTraceHandler(data []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		r, err := tracev2.NewReader(bytes.NewReader(data))
		if err != nil {
			log.Printf("failed to create trace reader: %v", err)
			return
		}
		n := 0
		for {
			ev, err := r.ReadEvent()
			if err == io.EOF {
				break
			} else if err != nil {
				log.Printf("failed to read event: %v", err)
				return
			}

			if ev.Kind() == tracev2.EventSync {
				fmt.Printf("ev: %v\n", ev)
			}
			n++
			// fmt.Printf("%s\n", ev.String())
		}
		fmt.Printf("n: %v\n", n)
	})
}
