// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pprof_test

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof" // register pprof handlers in http.DefaultServeMux
)

func Example() {
	// use http.DefaultServeMux
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Sample!")
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
