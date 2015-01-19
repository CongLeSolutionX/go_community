// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	_core "runtime/internal/core"
	_maps "runtime/internal/maps"
	_sched "runtime/internal/sched"
	_schedinit "runtime/internal/schedinit"
	_seq "runtime/internal/seq"
	_strings "runtime/internal/strings"
	"unsafe"
)

func stringtoslicebyte(s string) []byte {
	b := rawbyteslice(len(s))
	copy(b, s)
	return b
}

func stringtoslicerune(s string) []rune {
	// two passes.
	// unlike slicerunetostring, no race because strings are immutable.
	n := 0
	t := s
	for len(s) > 0 {
		_, k := _seq.Charntorune(s)
		s = s[k:]
		n++
	}
	a := _seq.Rawruneslice(n)
	n = 0
	for len(t) > 0 {
		r, k := _seq.Charntorune(t)
		t = t[k:]
		a[n] = r
		n++
	}
	return a
}

// rawbyteslice allocates a new byte slice. The byte slice is not zeroed.
func rawbyteslice(size int) (b []byte) {
	cap := _schedinit.Goroundupsize(uintptr(size))
	p := _maps.Mallocgc(cap, nil, _sched.FlagNoScan|_sched.FlagNoZero)
	if cap != uintptr(size) {
		_core.Memclr(_core.Add(p, uintptr(size)), cap-uintptr(size))
	}

	(*_core.Slice)(unsafe.Pointer(&b)).Array = (*uint8)(p)
	(*_core.Slice)(unsafe.Pointer(&b)).Len = uint(size)
	(*_core.Slice)(unsafe.Pointer(&b)).Cap = uint(cap)
	return
}

// used by cmd/cgo
func gobytes(p *byte, n int) []byte {
	if n == 0 {
		return make([]byte, 0)
	}
	x := make([]byte, n)
	_sched.Memmove(unsafe.Pointer(&x[0]), unsafe.Pointer(p), uintptr(n))
	return x
}

func gostringsize(n int) string {
	s, _ := _strings.Rawstring(n)
	return s
}

func gostringn(p *byte, l int) string {
	if l == 0 {
		return ""
	}
	s, b := _strings.Rawstring(l)
	_sched.Memmove(unsafe.Pointer(&b[0]), unsafe.Pointer(p), uintptr(l))
	return s
}
