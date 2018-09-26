// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http_test

import (
	"fmt"
	"log"
	"net/http"
)

type exampleHandler struct{}

func (e exampleHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "/" {
		fmt.Fprintln(w, "This is the example handler")
		return
	}
	w.WriteHeader(http.StatusBadRequest)
}

func ExampleHandle() {
	http.Handle("*", exampleHandler{})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
