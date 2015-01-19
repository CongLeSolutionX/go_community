// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package seq

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	_maps "runtime/internal/maps"
	_sched "runtime/internal/sched"
	_schedinit "runtime/internal/schedinit"
	_strings "runtime/internal/strings"
	"unsafe"
)

func slicebytetostring(b []byte) string {
	if _sched.Raceenabled && len(b) > 0 {
		racereadrangepc(unsafe.Pointer(&b[0]),
			uintptr(len(b)),
			_lock.Getcallerpc(unsafe.Pointer(&b)),
			_lock.FuncPC(slicebytetostring))
	}
	s, c := _strings.Rawstring(len(b))
	copy(c, b)
	return s
}

func slicebytetostringtmp(b []byte) string {
	// Return a "string" referring to the actual []byte bytes.
	// This is only for use by internal compiler optimizations
	// that know that the string form will be discarded before
	// the calling goroutine could possibly modify the original
	// slice or synchronize with another goroutine.
	// Today, the only such case is a m[string(k)] lookup where
	// m is a string-keyed map and k is a []byte.

	if _sched.Raceenabled && len(b) > 0 {
		racereadrangepc(unsafe.Pointer(&b[0]),
			uintptr(len(b)),
			_lock.Getcallerpc(unsafe.Pointer(&b)),
			_lock.FuncPC(slicebytetostringtmp))
	}
	return *(*string)(unsafe.Pointer(&b))
}

func slicerunetostring(a []rune) string {
	if _sched.Raceenabled && len(a) > 0 {
		racereadrangepc(unsafe.Pointer(&a[0]),
			uintptr(len(a))*unsafe.Sizeof(a[0]),
			_lock.Getcallerpc(unsafe.Pointer(&a)),
			_lock.FuncPC(slicerunetostring))
	}
	var dum [4]byte
	size1 := 0
	for _, r := range a {
		size1 += runetochar(dum[:], r)
	}
	s, b := _strings.Rawstring(size1 + 3)
	size2 := 0
	for _, r := range a {
		// check for race
		if size2 >= size1 {
			break
		}
		size2 += runetochar(b[size2:], r)
	}
	return s[:size2]
}

func intstring(v int64) string {
	s, b := _strings.Rawstring(4)
	n := runetochar(b, rune(v))
	return s[:n]
}

// stringiter returns the index of the next
// rune after the rune that starts at s[k].
func stringiter(s string, k int) int {
	if k >= len(s) {
		// 0 is end of iteration
		return 0
	}

	c := s[k]
	if c < runeself {
		return k + 1
	}

	// multi-char rune
	_, n := Charntorune(s[k:])
	return k + n
}

// stringiter2 returns the rune that starts at s[k]
// and the index where the next rune starts.
func stringiter2(s string, k int) (int, rune) {
	if k >= len(s) {
		// 0 is end of iteration
		return 0, 0
	}

	c := s[k]
	if c < runeself {
		return k + 1, rune(c)
	}

	// multi-char rune
	r, n := Charntorune(s[k:])
	return k + n, r
}

// rawruneslice allocates a new rune slice. The rune slice is not zeroed.
func Rawruneslice(size int) (b []rune) {
	if uintptr(size) > _core.MaxMem/4 {
		_lock.Gothrow("out of memory")
	}
	mem := _schedinit.Goroundupsize(uintptr(size) * 4)
	p := _maps.Mallocgc(mem, nil, _sched.FlagNoScan|_sched.FlagNoZero)
	if mem != uintptr(size)*4 {
		_core.Memclr(_core.Add(p, uintptr(size)*4), mem-uintptr(size)*4)
	}

	(*_core.Slice)(unsafe.Pointer(&b)).Array = (*uint8)(p)
	(*_core.Slice)(unsafe.Pointer(&b)).Len = uint(size)
	(*_core.Slice)(unsafe.Pointer(&b)).Cap = uint(mem / 4)
	return
}
