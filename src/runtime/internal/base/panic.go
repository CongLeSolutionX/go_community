// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package base

import (
	"unsafe"
)

var divideError = error(ErrorString("integer divide by zero"))

func panicdivide() {
	panic(divideError)
}

var overflowError = error(ErrorString("integer overflow"))

func panicoverflow() {
	panic(overflowError)
}

var floatError = error(ErrorString("floating point error"))

func panicfloat() {
	panic(floatError)
}

var memoryError = error(ErrorString("invalid memory address or nil pointer dereference"))

func panicmem() {
	panic(memoryError)
}

//go:nosplit
func Startpanic() {
	Systemstack(startpanic_m)
}

//go:nosplit
func Dopanic(unused int) {
	pc := Getcallerpc(unsafe.Pointer(&unused))
	sp := Getcallersp(unsafe.Pointer(&unused))
	gp := Getg()
	Systemstack(func() {
		dopanic_m(gp, pc, sp) // should never return
	})
	*(*int)(nil) = 0
}

//go:nosplit
func Throw(s string) {
	print("fatal error: ", s, "\n")
	gp := Getg()
	if gp.M.Throwing == 0 {
		gp.M.Throwing = 1
	}
	Startpanic()
	Dopanic(0)
	*(*int)(nil) = 0 // not reached
}
