// errorcheck -0 -m -l

// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test escape analysis for function parameters.

// In this test almost everything is BAD except the simplest cases
// where input directly flows to output.

package foo

import "unsafe"

const (
	_MHeapMap_TotalBits = 31
	_MaxMem             = uintptr(1<<_MHeapMap_TotalBits - 1)
)

var maxstring uintptr = 256 // a hint for print

func casuintptr(ptr *uintptr, old, new uintptr) bool { // ERROR "casuintptr ptr does not escape$"
	return true
}

func exits(msg *byte) { // ERROR "exits msg does not escape$"
}

type stringStruct struct {
	str unsafe.Pointer
	len int
}

//go:nosplit
func findnull(s *byte) int { // ERROR "findnull s does not escape$"
	l := 0
	return l
}

//go:nosplit
func itoa(buf []byte, val uint64) []byte { // ERROR "leaking param: buf to result ~r2 level=0$"
	i := len(buf) - 1
	return buf[i:]
}

var Sink *stringStruct

//go:nosplit
func gostringnocopy(str *byte) string { // ERROR "leaking param: str$"
	var s string                              // ERROR "moved to heap: s$"
	sp := (*stringStruct)(unsafe.Pointer(&s)) // ERROR "&s escapes to heap$"
	Sink = sp                                 // expose heap storage.  Should cause tmp below to be heap allocated
	sp.str = unsafe.Pointer(str)
	sp.len = findnull(str)
	return s
}

//go:nosplit
func exit(e int) {
	var status []byte
	if e == 0 {
		status = []byte("\x00") // ERROR "exit \(\[\]byte\)\(.\\x00.\) does not escape$"
	} else {
		var tmp [32]byte                                                       // ERROR "moved to heap: tmp$"
		status = []byte(gostringnocopy(&itoa(tmp[:len(tmp)-1], uint64(e))[0])) // ERROR "&itoa\(tmp\[:len\(tmp\) - 1\], uint64\(e\)\)\[0\] escapes to heap$" "exit \(\[\]byte\)\(gostringnocopy\(&itoa\(tmp\[:len\(tmp\) - 1\], uint64\(e\)\)\[0\]\)\) does not escape$" "tmp escapes to heap$"
	}
	exits(&status[0]) // ERROR "exit &status\[0\] does not escape$"
}
