// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lock

import (
	_core "runtime/internal/core"
	"unsafe"
)

//go:nosplit
func Findnull(s *byte) int {
	if s == nil {
		return 0
	}
	p := (*[_core.MaxMem/2 - 1]byte)(unsafe.Pointer(s))
	l := 0
	for p[l] != 0 {
		l++
	}
	return l
}

var Maxstring uintptr = 256 // a hint for print

//go:nosplit
func Gostringnocopy(str *byte) string {
	var s string
	sp := (*StringStruct)(unsafe.Pointer(&s))
	sp.Str = unsafe.Pointer(str)
	sp.Len = Findnull(str)
	for {
		ms := Maxstring
		if uintptr(len(s)) <= ms || _core.Casuintptr(&Maxstring, ms, uintptr(len(s))) {
			break
		}
	}
	return s
}
