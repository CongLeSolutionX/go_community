// run

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"runtime"
)

var sink *[20]byte

func f() (x [20]byte) {
	// Force x to be allocated on the heap.
	sink = &x
	sink = nil

	// Go to deferreturn after the panic below.
	defer func() {
		recover()
	}()

	// This call collects the heap-allocated version of x (oops!)
	runtime.GC()

	// Allocate that same object again and clobber it.
	y := new([20]byte)
	for i := 0; i < 20; i++ {
		y[i] = 99
	}
	// Make sure y is heap allocated.
	sink = y

	panic(nil)

	// After the recover we reach the deferreturn, which
	// copies the heap version of x back to the stack.
	// It gets the pointer to x from a stack slot that was
	// not marked as live during the call to runtime.GC().
}
func main() {
	x := f()
	for _, v := range x {
		if v != 0 {
			fmt.Printf("%v\n", x)
			panic("bad")
		}
	}
}
