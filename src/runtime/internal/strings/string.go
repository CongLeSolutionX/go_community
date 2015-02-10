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

// The constant is known to the compiler.
// There is no fundamental theory behind this number.
const tmpStringBufSize = 32

type TmpBuf [tmpStringBufSize]byte

// concatstrings implements a Go string concatenation x+y+z+...
// The operands are passed in the slice a.
// If buf != nil, the compiler has determined that the result does not
// escape the calling function, so the string data can be stored in buf
// if small enough.
func concatstrings(buf *TmpBuf, a []string) string {
	idx := 0
	l := 0
	count := 0
	for i, x := range a {
		n := len(x)
		if n == 0 {
			continue
		}
		if l+n < l {
			_lock.Throw("string concatenation too long")
		}
		l += n
		count++
		idx = i
	}
	if count == 0 {
		return ""
	}

	// If there is just one string and either it is not on the stack
	// or our result does not escape the calling frame (buf != nil),
	// then we can return that string directly.
	if count == 1 && (buf != nil || !stringDataOnStack(a[idx])) {
		return a[idx]
	}
	s, b := Rawstringtmp(buf, l)
	l = 0
	for _, x := range a {
		copy(b[l:], x)
		l += len(x)
	}
	return s
}

func concatstring2(buf *TmpBuf, a [2]string) string {
	return concatstrings(buf, a[:])
}

func concatstring3(buf *TmpBuf, a [3]string) string {
	return concatstrings(buf, a[:])
}

func concatstring4(buf *TmpBuf, a [4]string) string {
	return concatstrings(buf, a[:])
}

func concatstring5(buf *TmpBuf, a [5]string) string {
	return concatstrings(buf, a[:])
}

// stringDataOnStack reports whether the string's data is
// stored on the current goroutine's stack.
func stringDataOnStack(s string) bool {
	ptr := uintptr((*_lock.StringStruct)(unsafe.Pointer(&s)).Str)
	stk := _core.Getg().Stack
	return stk.Lo <= ptr && ptr < stk.Hi
}

func Rawstringtmp(buf *TmpBuf, l int) (s string, b []byte) {
	if buf != nil && l <= len(buf) {
		b = buf[:l]
		s = Slicebytetostringtmp(b)
	} else {
		s, b = Rawstring(l)
	}
	return
}

func Slicebytetostringtmp(b []byte) string {
	// Return a "string" referring to the actual []byte bytes.
	// This is only for use by internal compiler optimizations
	// that know that the string form will be discarded before
	// the calling goroutine could possibly modify the original
	// slice or synchronize with another goroutine.
	// First such case is a m[string(k)] lookup where
	// m is a string-keyed map and k is a []byte.
	// Second such case is "<"+string(b)+">" concatenation where b is []byte.
	// Third such case is string(b)=="foo" comparison where b is []byte.

	if _sched.Raceenabled && len(b) > 0 {
		Racereadrangepc(unsafe.Pointer(&b[0]),
			uintptr(len(b)),
			_lock.Getcallerpc(unsafe.Pointer(&b)),
			_lock.FuncPC(Slicebytetostringtmp))
	}
	return *(*string)(unsafe.Pointer(&b))
}

// rawstring allocates storage for a new string. The returned
// string and byte slice both refer to the same storage.
// The storage is not zeroed. Callers should use
// b to set the string contents and then drop b.
func Rawstring(size int) (s string, b []byte) {
	p := _maps.Mallocgc(uintptr(size), nil, _sched.XFlagNoScan|_sched.XFlagNoZero)

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
