// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	_base "runtime/internal/base"
	_iface "runtime/internal/iface"
	_race "runtime/internal/race"
	"unsafe"
)

// The constant is known to the compiler.
// There is no fundamental theory behind this number.
const tmpStringBufSize = 32

type tmpBuf [tmpStringBufSize]byte

// concatstrings implements a Go string concatenation x+y+z+...
// The operands are passed in the slice a.
// If buf != nil, the compiler has determined that the result does not
// escape the calling function, so the string data can be stored in buf
// if small enough.
func concatstrings(buf *tmpBuf, a []string) string {
	idx := 0
	l := 0
	count := 0
	for i, x := range a {
		n := len(x)
		if n == 0 {
			continue
		}
		if l+n < l {
			_base.Throw("string concatenation too long")
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
	s, b := rawstringtmp(buf, l)
	l = 0
	for _, x := range a {
		copy(b[l:], x)
		l += len(x)
	}
	return s
}

func concatstring2(buf *tmpBuf, a [2]string) string {
	return concatstrings(buf, a[:])
}

func concatstring3(buf *tmpBuf, a [3]string) string {
	return concatstrings(buf, a[:])
}

func concatstring4(buf *tmpBuf, a [4]string) string {
	return concatstrings(buf, a[:])
}

func concatstring5(buf *tmpBuf, a [5]string) string {
	return concatstrings(buf, a[:])
}

// Buf is a fixed-size buffer for the result,
// it is not nil if the result does not escape.
func slicebytetostring(buf *tmpBuf, b []byte) string {
	l := len(b)
	if l == 0 {
		// Turns out to be a relatively common case.
		// Consider that you want to parse out data between parens in "foo()bar",
		// you find the indices and convert the subslice to string.
		return ""
	}
	if _base.Raceenabled && l > 0 {
		_race.Racereadrangepc(unsafe.Pointer(&b[0]),
			uintptr(l),
			_base.Getcallerpc(unsafe.Pointer(&b)),
			_base.FuncPC(slicebytetostring))
	}
	s, c := rawstringtmp(buf, l)
	copy(c, b)
	return s
}

// stringDataOnStack reports whether the string's data is
// stored on the current goroutine's stack.
func stringDataOnStack(s string) bool {
	ptr := uintptr((*_base.StringStruct)(unsafe.Pointer(&s)).Str)
	stk := _base.Getg().Stack
	return stk.Lo <= ptr && ptr < stk.Hi
}

func rawstringtmp(buf *tmpBuf, l int) (s string, b []byte) {
	if buf != nil && l <= len(buf) {
		b = buf[:l]
		s = slicebytetostringtmp(b)
	} else {
		s, b = rawstring(l)
	}
	return
}

func slicebytetostringtmp(b []byte) string {
	// Return a "string" referring to the actual []byte bytes.
	// This is only for use by internal compiler optimizations
	// that know that the string form will be discarded before
	// the calling goroutine could possibly modify the original
	// slice or synchronize with another goroutine.
	// First such case is a m[string(k)] lookup where
	// m is a string-keyed map and k is a []byte.
	// Second such case is "<"+string(b)+">" concatenation where b is []byte.
	// Third such case is string(b)=="foo" comparison where b is []byte.

	if _base.Raceenabled && len(b) > 0 {
		_race.Racereadrangepc(unsafe.Pointer(&b[0]),
			uintptr(len(b)),
			_base.Getcallerpc(unsafe.Pointer(&b)),
			_base.FuncPC(slicebytetostringtmp))
	}
	return *(*string)(unsafe.Pointer(&b))
}

func stringtoslicebyte(buf *tmpBuf, s string) []byte {
	var b []byte
	if buf != nil && len(s) <= len(buf) {
		b = buf[:len(s)]
	} else {
		b = rawbyteslice(len(s))
	}
	copy(b, s)
	return b
}

func stringtoslicebytetmp(s string) []byte {
	// Return a slice referring to the actual string bytes.
	// This is only for use by internal compiler optimizations
	// that know that the slice won't be mutated.
	// The only such case today is:
	// for i, c := range []byte(str)

	str := (*_base.StringStruct)(unsafe.Pointer(&s))
	ret := _base.Slice{Array: unsafe.Pointer(str.Str), Len: str.Len, Cap: str.Len}
	return *(*[]byte)(unsafe.Pointer(&ret))
}

func stringtoslicerune(buf *[tmpStringBufSize]rune, s string) []rune {
	// two passes.
	// unlike slicerunetostring, no race because strings are immutable.
	n := 0
	t := s
	for len(s) > 0 {
		_, k := charntorune(s)
		s = s[k:]
		n++
	}
	var a []rune
	if buf != nil && n <= len(buf) {
		a = buf[:n]
	} else {
		a = rawruneslice(n)
	}
	n = 0
	for len(t) > 0 {
		r, k := charntorune(t)
		t = t[k:]
		a[n] = r
		n++
	}
	return a
}

func slicerunetostring(buf *tmpBuf, a []rune) string {
	if _base.Raceenabled && len(a) > 0 {
		_race.Racereadrangepc(unsafe.Pointer(&a[0]),
			uintptr(len(a))*unsafe.Sizeof(a[0]),
			_base.Getcallerpc(unsafe.Pointer(&a)),
			_base.FuncPC(slicerunetostring))
	}
	var dum [4]byte
	size1 := 0
	for _, r := range a {
		size1 += runetochar(dum[:], r)
	}
	s, b := rawstringtmp(buf, size1+3)
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

func intstring(buf *[4]byte, v int64) string {
	var s string
	var b []byte
	if buf != nil {
		b = buf[:]
		s = slicebytetostringtmp(b)
	} else {
		s, b = rawstring(4)
	}
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
	_, n := charntorune(s[k:])
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
	r, n := charntorune(s[k:])
	return k + n, r
}

// rawstring allocates storage for a new string. The returned
// string and byte slice both refer to the same storage.
// The storage is not zeroed. Callers should use
// b to set the string contents and then drop b.
func rawstring(size int) (s string, b []byte) {
	p := _iface.Mallocgc(uintptr(size), nil, _base.XFlagNoScan|_base.XFlagNoZero)

	(*_base.StringStruct)(unsafe.Pointer(&s)).Str = p
	(*_base.StringStruct)(unsafe.Pointer(&s)).Len = size

	*(*_base.Slice)(unsafe.Pointer(&b)) = _base.Slice{p, size, size}

	for {
		ms := _base.Maxstring
		if uintptr(size) <= uintptr(ms) || _base.Casuintptr((*uintptr)(unsafe.Pointer(&_base.Maxstring)), uintptr(ms), uintptr(size)) {
			return
		}
	}
}

// rawbyteslice allocates a new byte slice. The byte slice is not zeroed.
func rawbyteslice(size int) (b []byte) {
	cap := roundupsize(uintptr(size))
	p := _iface.Mallocgc(cap, nil, _base.XFlagNoScan|_base.XFlagNoZero)
	if cap != uintptr(size) {
		_base.Memclr(_base.Add(p, uintptr(size)), cap-uintptr(size))
	}

	*(*_base.Slice)(unsafe.Pointer(&b)) = _base.Slice{p, size, int(cap)}
	return
}

// rawruneslice allocates a new rune slice. The rune slice is not zeroed.
func rawruneslice(size int) (b []rune) {
	if uintptr(size) > _base.MaxMem/4 {
		_base.Throw("out of memory")
	}
	mem := roundupsize(uintptr(size) * 4)
	p := _iface.Mallocgc(mem, nil, _base.XFlagNoScan|_base.XFlagNoZero)
	if mem != uintptr(size)*4 {
		_base.Memclr(_base.Add(p, uintptr(size)*4), mem-uintptr(size)*4)
	}

	*(*_base.Slice)(unsafe.Pointer(&b)) = _base.Slice{p, size, int(mem / 4)}
	return
}

// used by cmd/cgo
func gobytes(p *byte, n int) []byte {
	if n == 0 {
		return make([]byte, 0)
	}
	x := make([]byte, n)
	_base.Memmove(unsafe.Pointer(&x[0]), unsafe.Pointer(p), uintptr(n))
	return x
}

func gostring(p *byte) string {
	l := _base.Findnull(p)
	if l == 0 {
		return ""
	}
	s, b := rawstring(l)
	_base.Memmove(unsafe.Pointer(&b[0]), unsafe.Pointer(p), uintptr(l))
	return s
}

func gostringn(p *byte, l int) string {
	if l == 0 {
		return ""
	}
	s, b := rawstring(l)
	_base.Memmove(unsafe.Pointer(&b[0]), unsafe.Pointer(p), uintptr(l))
	return s
}
