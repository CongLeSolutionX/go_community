// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package seq

import (
	_channels "runtime/internal/channels"
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	_maps "runtime/internal/maps"
	_sched "runtime/internal/sched"
	_schedinit "runtime/internal/schedinit"
	_strings "runtime/internal/strings"
	"unsafe"
)

// TODO: take uintptrs instead of int64s?
func makeslice(t *slicetype, len64 int64, cap64 int64) _sched.SliceStruct {
	// NOTE: The len > MaxMem/elemsize check here is not strictly necessary,
	// but it produces a 'len out of range' error instead of a 'cap out of range' error
	// when someone does make([]T, bignumber). 'cap out of range' is true too,
	// but since the cap is only being supplied implicitly, saying len is clearer.
	// See issue 4085.
	len := int(len64)
	if len64 < 0 || int64(len) != len64 || t.elem.Size > 0 && uintptr(len) > _core.MaxMem/uintptr(t.elem.Size) {
		panic(_sched.ErrorString("makeslice: len out of range"))
	}
	cap := int(cap64)
	if cap < len || int64(cap) != cap64 || t.elem.Size > 0 && uintptr(cap) > _core.MaxMem/uintptr(t.elem.Size) {
		panic(_sched.ErrorString("makeslice: cap out of range"))
	}
	p := _maps.Newarray(t.elem, uintptr(cap))
	return _sched.SliceStruct{p, len, cap}
}

// TODO: take uintptr instead of int64?
func growslice(t *slicetype, old _sched.SliceStruct, n int64) _sched.SliceStruct {
	if n < 1 {
		panic(_sched.ErrorString("growslice: invalid n"))
	}

	cap64 := int64(old.Cap) + n
	cap := int(cap64)

	if int64(cap) != cap64 || cap < old.Cap || t.elem.Size > 0 && uintptr(cap) > _core.MaxMem/uintptr(t.elem.Size) {
		panic(_sched.ErrorString("growslice: cap out of range"))
	}

	if _sched.Raceenabled {
		callerpc := _lock.Getcallerpc(unsafe.Pointer(&t))
		_strings.Racereadrangepc(old.Array, uintptr(old.Len*int(t.elem.Size)), callerpc, _lock.FuncPC(growslice))
	}

	et := t.elem
	if et.Size == 0 {
		return _sched.SliceStruct{old.Array, old.Len, cap}
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

	if uintptr(newcap) >= _core.MaxMem/uintptr(et.Size) {
		panic(_sched.ErrorString("growslice: cap out of range"))
	}
	lenmem := uintptr(old.Len) * uintptr(et.Size)
	capmem := _schedinit.Roundupsize(uintptr(newcap) * uintptr(et.Size))
	newcap = int(capmem / uintptr(et.Size))
	var p unsafe.Pointer
	if et.Kind&_channels.KindNoPointers != 0 {
		p = rawmem(capmem)
		_sched.Memmove(p, old.Array, lenmem)
		_core.Memclr(_core.Add(p, lenmem), capmem-lenmem)
	} else {
		// Note: can't use rawmem (which avoids zeroing of memory), because then GC can scan unitialized memory.
		// TODO(rsc): Use memmove when !needwb().
		p = _maps.Newarray(et, uintptr(newcap))
		for i := 0; i < old.Len; i++ {
			_channels.Typedmemmove(et, _core.Add(p, uintptr(i)*et.Size), _core.Add(old.Array, uintptr(i)*et.Size))
		}
	}

	return _sched.SliceStruct{p, old.Len, newcap}
}

func slicecopy(to _sched.SliceStruct, fm _sched.SliceStruct, width uintptr) int {
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

	if _sched.Raceenabled {
		callerpc := _lock.Getcallerpc(unsafe.Pointer(&to))
		pc := _lock.FuncPC(slicecopy)
		racewriterangepc(to.Array, uintptr(n*int(width)), callerpc, pc)
		_strings.Racereadrangepc(fm.Array, uintptr(n*int(width)), callerpc, pc)
	}

	size := uintptr(n) * width
	if size == 1 { // common case worth about 2x to do here
		// TODO: is this still worth it with new memmove impl?
		*(*byte)(to.Array) = *(*byte)(fm.Array) // known to be a byte pointer
	} else {
		_sched.Memmove(to.Array, fm.Array, size)
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

	if _sched.Raceenabled {
		callerpc := _lock.Getcallerpc(unsafe.Pointer(&to))
		pc := _lock.FuncPC(slicestringcopy)
		racewriterangepc(unsafe.Pointer(&to[0]), uintptr(n), callerpc, pc)
	}

	_sched.Memmove(unsafe.Pointer(&to[0]), unsafe.Pointer((*_lock.StringStruct)(unsafe.Pointer(&fm)).Str), uintptr(n))
	return n
}
