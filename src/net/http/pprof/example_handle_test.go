// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pprof_test

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
)

func Example_handle() {
	// register pprof handlers in http.DefaultServeMux
	// import _ "net/http/pprof"

	r := http.NewServeMux() // Custom Mux

	// register http.DefaultServeMux in Custom Mux only for urls related to /debug/pprof/
	r.Handle("/debug/pprof/", http.DefaultServeMux)

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Sample!")
	})

	log.Fatal(http.ListenAndServe(":8080", r))
}
