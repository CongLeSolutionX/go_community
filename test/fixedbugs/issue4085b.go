// run

// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"strings"
	"unsafe"
)

type T []int

var x T
var c = make(T, 8)

func main() {
	n := -1
	shouldPanic("len out of range", func() { x = make(T, n) })
	shouldPanic("cap out of range", func() { x = make(T, 0, n) })
	shouldPanic("len out of range", func() { x = make(T, int64(n)) })
	shouldPanic("cap out of range", func() { x = make(T, 0, int64(n)) })

	// Test make+copy panics since the gc compiler optimizes these
	// to runtime.makeslicecopy calls.
	shouldPanic("len out of range", func() { x = make(T, n); copy(x, c) })
	shouldPanic("cap out of range", func() { x = make(T, 0, n); copy(x, c) })
	shouldPanic("len out of range", func() { x = make(T, int64(n)); copy(x, c) })
	shouldPanic("cap out of range", func() { x = make(T, 0, int64(n)); copy(x, c) })

	var t *byte
	if unsafe.Sizeof(t) == 8 {
		// Test mem > maxAlloc
		var n2 int64 = 1 << 59
		shouldPanic("len out of range", func() { x = make(T, int(n2)) })
		shouldPanic("cap out of range", func() { x = make(T, 0, int(n2)) })
		shouldPanic("len out of range", func() { x = make(T, int(n2)); copy(x, c) })
		shouldPanic("cap out of range", func() { x = make(T, 0, int(n2)); copy(x, c) })
		// Test elem.size*cap overflow
		n2 = 1<<63 - 1
		shouldPanic("len out of range", func() { x = make(T, int(n2)) })
		shouldPanic("cap out of range", func() { x = make(T, 0, int(n2)) })
		shouldPanic("len out of range", func() { x = make(T, int(n2)); copy(x, c) })
		shouldPanic("cap out of range", func() { x = make(T, 0, int(n2)); copy(x, c) })
	} else {
		n = 1<<31 - 1
		shouldPanic("len out of range", func() { x = make(T, n) })
		shouldPanic("cap out of range", func() { x = make(T, 0, n) })
		shouldPanic("len out of range", func() { x = make(T, int64(n)) })
		shouldPanic("cap out of range", func() { x = make(T, 0, int64(n)) })

		shouldPanic("len out of range", func() { x = make(T, n); copy(x, c) })
		shouldPanic("cap out of range", func() { x = make(T, 0, n); copy(x, c) })
		shouldPanic("len out of range", func() { x = make(T, int64(n)); copy(x, c) })
		shouldPanic("cap out of range", func() { x = make(T, 0, int64(n)); copy(x, c) })
	}

	// Test make in append panics since the gc compiler optimizes makes in appends.
	shouldPanic("len out of range", func() { x = append(T{}, make(T, n)...) })
	shouldPanic("cap out of range", func() { x = append(T{}, make(T, 0, n)...) })
	shouldPanic("len out of range", func() { x = append(T{}, make(T, int64(n))...) })
	shouldPanic("cap out of range", func() { x = append(T{}, make(T, 0, int64(n))...) })
}

func shouldPanic(str string, f func()) {
	defer func() {
		err := recover()
		if err == nil {
			panic("did not panic")
		}
		s := err.(error).Error()
		if !strings.Contains(s, str) {
			panic("got panic " + s + ", want " + str)
		}
	}()

	f()
}
