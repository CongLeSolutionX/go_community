// +build amd64
// run

// Copyright 2016 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

var out int

func main() {
	wantPanic("test1", func() {
		v := f(0)
		out = 1 / v
	})
	wantPanic("test2", func() {
		v := f(0)
		_ = 1 / v
	})
	wantPanic("test3", func() {
		v := 0
		_ = 1 / v
	})
	wantPanic("test4", func() { divby(0) })
}

func wantPanic(test string, fn func()) {
	defer func() {
		if e := recover(); e == nil {
			panic(test + ": expected panic")
		}
	}()
	fn()
}

//gc:noinline
func f(n int) int { return n }

//gc:noinline
func divby(v int) {
	_ = 1 / v
}
