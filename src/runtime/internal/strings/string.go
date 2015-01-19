// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strings

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	_maps "runtime/internal/maps"
	_sched "runtime/internal/sched"
	"unsafe"
)

func concatstrings(a []string) string {
	idx := 0
	l := 0
	count := 0
	for i, x := range a {
		n := len(x)
		if n == 0 {
			continue
		}
		if l+n < l {
			_lock.Gothrow("string concatenation too long")
		}
		l += n
		count++
		idx = i
	}
	if count == 0 {
		return ""
	}
	if count == 1 {
		return a[idx]
	}
	s, b := Rawstring(l)
	l = 0
	for _, x := range a {
		copy(b[l:], x)
		l += len(x)
	}
	return s
}

func concatstring2(a [2]string) string {
	return concatstrings(a[:])
}

func concatstring3(a [3]string) string {
	return concatstrings(a[:])
}

func concatstring4(a [4]string) string {
	return concatstrings(a[:])
}

func concatstring5(a [5]string) string {
	return concatstrings(a[:])
}

// rawstring allocates storage for a new string. The returned
// string and byte slice both refer to the same storage.
// The storage is not zeroed. Callers should use
// b to set the string contents and then drop b.
func Rawstring(size int) (s string, b []byte) {
	p := _maps.Mallocgc(uintptr(size), nil, _sched.FlagNoScan|_sched.FlagNoZero)

	(*_lock.StringStruct)(unsafe.Pointer(&s)).Str = p
	(*_lock.StringStruct)(unsafe.Pointer(&s)).Len = size

	(*_core.Slice)(unsafe.Pointer(&b)).Array = (*uint8)(p)
	(*_core.Slice)(unsafe.Pointer(&b)).Len = uint(size)
	(*_core.Slice)(unsafe.Pointer(&b)).Cap = uint(size)

	for {
		ms := _lock.Maxstring
		if uintptr(size) <= uintptr(ms) || _core.Casuintptr((*uintptr)(unsafe.Pointer(&_lock.Maxstring)), uintptr(ms), uintptr(size)) {
			return
		}
	}
}

func gostring(p *byte) string {
	l := _lock.Findnull(p)
	if l == 0 {
		return ""
	}
	s, b := Rawstring(l)
	_sched.Memmove(unsafe.Pointer(&b[0]), unsafe.Pointer(p), uintptr(l))
	return s
}
