// run

// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This test makes sure unsafe-uintptr arguments are not
// kept alive longer than expected.

package main

import (
	"context"
	"runtime"
	"time"
	"unsafe"
)

var done = make(chan bool)

func setup() unsafe.Pointer {
	s := "ok"
	runtime.SetFinalizer(&s, func(p *string) { close(done) })
	return unsafe.Pointer(&s)
}

//go:noinline
//go:uintptrescapes
func before(p uintptr) int {
	runtime.GC()
	select {
	case <-done:
		panic("GC early")
	default:
	}
	return 0
}

func after() int {
	runtime.GC()
	ctx, cancelFn := context.WithTimeout(context.Background(), time.Minute)
	defer cancelFn()
	select {
	case <-done:
	case <-ctx.Done():
		panic("GC late")
	}
	<-done
	return 0
}

func main() {
	_ = before(uintptr(setup())) + after()
}
