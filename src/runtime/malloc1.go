// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// See malloc.h for overview.
//
// TODO(rsc): double-check stats.

package runtime

import (
	_core "runtime/internal/core"
	_gc "runtime/internal/gc"
	_lock "runtime/internal/lock"
	_maps "runtime/internal/maps"
	"unsafe"
)

// For use by Go. If it were a C enum it would be made available automatically,
// but the value of MaxMem is too large for enum.
// XXX - uintptr runtime·maxmem = MaxMem;

func mlookup(v uintptr, base *uintptr, size *uintptr, sp **_core.Mspan) int32 {
	_g_ := _core.Getg()

	_g_.M.Mcache.Local_nlookup++
	if _core.PtrSize == 4 && _g_.M.Mcache.Local_nlookup >= 1<<30 {
		// purge cache stats to prevent overflow
		_lock.Lock(&_lock.Mheap_.Lock)
		_gc.Purgecachedstats(_g_.M.Mcache)
		_lock.Unlock(&_lock.Mheap_.Lock)
	}

	s := _maps.MHeap_LookupMaybe(&_lock.Mheap_, unsafe.Pointer(v))
	if sp != nil {
		*sp = s
	}
	if s == nil {
		if base != nil {
			*base = 0
		}
		if size != nil {
			*size = 0
		}
		return 0
	}

	p := uintptr(s.Start) << _core.PageShift
	if s.Sizeclass == 0 {
		// Large object.
		if base != nil {
			*base = p
		}
		if size != nil {
			*size = s.Npages << _core.PageShift
		}
		return 1
	}

	n := s.Elemsize
	if base != nil {
		i := (uintptr(v) - uintptr(p)) / n
		*base = p + i*n
	}
	if size != nil {
		*size = n
	}

	return 1
}
