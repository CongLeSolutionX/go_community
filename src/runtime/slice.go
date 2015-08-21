// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	_base "runtime/internal/base"
	_iface "runtime/internal/iface"
	_race "runtime/internal/race"
	"unsafe"
)

// TODO: take uintptrs instead of int64s?
func makeslice(t *slicetype, len64, cap64 int64) _base.Slice {
	// NOTE: The len > MaxMem/elemsize check here is not strictly necessary,
	// but it produces a 'len out of range' error instead of a 'cap out of range' error
	// when someone does make([]T, bignumber). 'cap out of range' is true too,
	// but since the cap is only being supplied implicitly, saying len is clearer.
	// See issue 4085.
	len := int(len64)
	if len64 < 0 || int64(len) != len64 || t.elem.Size > 0 && uintptr(len) > _base.MaxMem/uintptr(t.elem.Size) {
		panic(_base.ErrorString("makeslice: len out of range"))
	}
	cap := int(cap64)
	if cap < len || int64(cap) != cap64 || t.elem.Size > 0 && uintptr(cap) > _base.MaxMem/uintptr(t.elem.Size) {
		panic(_base.ErrorString("makeslice: cap out of range"))
	}
	p := newarray(t.elem, uintptr(cap))
	return _base.Slice{p, len, cap}
}

// growslice_n is a variant of growslice that takes the number of new elements
// instead of the new minimum capacity.
// TODO(rsc): This is used by append(slice, slice...).
// The compiler should change that code to use growslice directly (issue #11419).
func growslice_n(t *slicetype, old _base.Slice, n int) _base.Slice {
	if n < 1 {
		panic(_base.ErrorString("growslice: invalid n"))
	}
	return growslice(t, old, old.Cap+n)
}

// growslice handles slice growth during append.
// It is passed the slice type, the old slice, and the desired new minimum capacity,
// and it returns a new slice with at least that capacity, with the old data
// copied into it.
func growslice(t *slicetype, old _base.Slice, cap int) _base.Slice {
	if cap < old.Cap || t.elem.Size > 0 && uintptr(cap) > _base.MaxMem/uintptr(t.elem.Size) {
		panic(_base.ErrorString("growslice: cap out of range"))
	}

	if _base.Raceenabled {
		callerpc := _base.Getcallerpc(unsafe.Pointer(&t))
		_race.Racereadrangepc(old.Array, uintptr(old.Len*int(t.elem.Size)), callerpc, _base.FuncPC(growslice))
	}

	et := t.elem
	if et.Size == 0 {
		// append should not create a slice with nil pointer but non-zero len.
		// We assume that append doesn't need to preserve old.array in this case.
		return _base.Slice{unsafe.Pointer(&_iface.Zerobase), old.Len, cap}
	}

	newcap := old.Cap
	if newcap+newcap < cap {
		newcap = cap
	} else {
		for {
			if old.Len < 1024 {
				newcap += newcap
			} else {
				newcap += newcap / 4
			}
			if newcap >= cap {
				break
			}
		}
	}

	if uintptr(newcap) >= _base.MaxMem/uintptr(et.Size) {
		panic(_base.ErrorString("growslice: cap out of range"))
	}
	lenmem := uintptr(old.Len) * uintptr(et.Size)
	capmem := roundupsize(uintptr(newcap) * uintptr(et.Size))
	newcap = int(capmem / uintptr(et.Size))
	var p unsafe.Pointer
	if et.Kind&_iface.KindNoPointers != 0 {
		p = rawmem(capmem)
		_base.Memmove(p, old.Array, lenmem)
		_base.Memclr(_base.Add(p, lenmem), capmem-lenmem)
	} else {
		// Note: can't use rawmem (which avoids zeroing of memory), because then GC can scan uninitialized memory.
		p = newarray(et, uintptr(newcap))
		if !_base.WriteBarrierEnabled {
			_base.Memmove(p, old.Array, lenmem)
		} else {
			for i := uintptr(0); i < lenmem; i += et.Size {
				_iface.Typedmemmove(et, _base.Add(p, i), _base.Add(old.Array, i))
			}
		}
	}

	return _base.Slice{p, old.Len, newcap}
}

func slicecopy(to, fm _base.Slice, width uintptr) int {
	if fm.Len == 0 || to.Len == 0 {
		return 0
	}

	n := fm.Len
	if to.Len < n {
		n = to.Len
	}

	if width == 0 {
		return n
	}

	if _base.Raceenabled {
		callerpc := _base.Getcallerpc(unsafe.Pointer(&to))
		pc := _base.FuncPC(slicecopy)
		_race.Racewriterangepc(to.Array, uintptr(n*int(width)), callerpc, pc)
		_race.Racereadrangepc(fm.Array, uintptr(n*int(width)), callerpc, pc)
	}

	size := uintptr(n) * width
	if size == 1 { // common case worth about 2x to do here
		// TODO: is this still worth it with new memmove impl?
		*(*byte)(to.Array) = *(*byte)(fm.Array) // known to be a byte pointer
	} else {
		_base.Memmove(to.Array, fm.Array, size)
	}
	return int(n)
}

func slicestringcopy(to []byte, fm string) int {
	if len(fm) == 0 || len(to) == 0 {
		return 0
	}

	n := len(fm)
	if len(to) < n {
		n = len(to)
	}

	if _base.Raceenabled {
		callerpc := _base.Getcallerpc(unsafe.Pointer(&to))
		pc := _base.FuncPC(slicestringcopy)
		_race.Racewriterangepc(unsafe.Pointer(&to[0]), uintptr(n), callerpc, pc)
	}

	_base.Memmove(unsafe.Pointer(&to[0]), unsafe.Pointer((*_base.StringStruct)(unsafe.Pointer(&fm)).Str), uintptr(n))
	return n
}
