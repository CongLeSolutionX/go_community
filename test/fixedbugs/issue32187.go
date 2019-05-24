// run

// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// short-circuiting interface-to-concrete comparisons
// will not miss panics

package main

import "log"

func main() {
	var x interface{}
	var p *int
	var s []int
	tests := []struct {
		name string
		f    func()
	}{
		{"switch-case", func() {
			switch x {
			case x.(*int):
			}
		}},
		{"interface convertion", func() { _ = x == x.(*int) }},
		{"nil pointer dereference", func() { _ = x == *p }},
		{"out of bound", func() { _ = x == s[1] }},
	}

	for _, tc := range tests {
		testFuncShouldPanic(tc.name, tc.f)
	}
}

func testFuncShouldPanic(name string, f func()) {
	defer func() {
		if e := recover(); e == nil {
			log.Fatalf("%s: comparison did not panic\n", name)
		}
	}()
	f()
}
