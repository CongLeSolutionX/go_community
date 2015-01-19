// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sched

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
