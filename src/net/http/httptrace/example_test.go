// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package httptrace_test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptrace"
	rtrace "runtime/trace"
)

func Example() {

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	trace := &httptrace.ClientTrace{
		GotConn: func(connInfo httptrace.GotConnInfo) {
			fmt.Printf("Got Conn: %+v\n", connInfo)
		},
		DNSDone: func(dnsInfo httptrace.DNSDoneInfo) {
			fmt.Printf("DNS Info: %+v\n", dnsInfo)
		},
	}

	rtrace.Do(req.Context(), "Example", "one", func(ctx context.Context) string {
		req = req.WithContext(httptrace.WithClientTrace(ctx, trace))
		_, err := http.DefaultTransport.RoundTrip(req)
		if err != nil {
			log.Fatal(err)
		}
		return err.Error()
	})
}
