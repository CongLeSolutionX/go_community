// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http_test

import (
	"fmt"
	"log"
	"net/http"
	"sync"
)

type exampleHandler struct {
	n  int
	mu sync.Mutex
}

func (e *exampleHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e.mu.Lock()
	e.n++
	e.mu.Unlock()
	fmt.Fprintf(w, "count is %d\n", e.n)
}

func ExampleHandle() {
	h := exampleHandler{n: 0, mu: sync.Mutex{}}
	http.Handle("/count", &h)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
