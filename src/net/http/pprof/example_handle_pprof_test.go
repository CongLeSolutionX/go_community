// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pprof_test

import (
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"
)

func Example_handle_pprof() {
	// import (
	//     "net/http"
	//     "net/http/pprof"
	// )

	r := http.NewServeMux() // Custom Mux

	// register all pprof handlers in Custom Mux
	r.HandleFunc("/debug/pprof/", pprof.Index)
	r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	r.HandleFunc("/debug/pprof/profile", pprof.Profile)
	r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Sample!")
	})

	log.Fatal(http.ListenAndServe(":8080", r))
}
