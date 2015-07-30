// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package base

import (
	"unsafe"
)

//go:nosplit
func Findnull(s *byte) int {
	if s == nil {
		return 0
	}
	p := (*[MaxMem/2 - 1]byte)(unsafe.Pointer(s))
	l := 0
	for p[l] != 0 {
		l++
	}
	return l
}

var Maxstring uintptr = 256 // a hint for print

//go:nosplit
func Gostringnocopy(str *byte) string {
	ss := StringStruct{Str: unsafe.Pointer(str), Len: Findnull(str)}
	s := *(*string)(unsafe.Pointer(&ss))
	for {
		ms := Maxstring
		if uintptr(len(s)) <= ms || Casuintptr(&Maxstring, ms, uintptr(len(s))) {
			break
		}
	}
	return s
}
