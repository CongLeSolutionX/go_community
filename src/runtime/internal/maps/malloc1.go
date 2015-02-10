// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// See malloc.h for overview.
//
// TODO(rsc): double-check stats.

package maps

import (
	_core "runtime/internal/core"
	_gc "runtime/internal/gc"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
)

func largeAlloc(size uintptr, flag uint32) *_core.Mspan {
	// print("largeAlloc size=", size, "\n")

	if size+_core.PageSize < size {
		_lock.Throw("out of memory")
	}
	npages := size >> _core.PageShift
	if size&_core.PageMask != 0 {
		npages++
	}
	s := mHeap_Alloc(&_lock.Mheap_, npages, 0, true, flag&_sched.FlagNoZero == 0)
	if s == nil {
		_lock.Throw("out of memory")
	}
	s.Limit = uintptr(s.Start)<<_core.PageShift + size
	_gc.HeapBitsForSpan(s.Base()).InitSpan(s.Layout())
	return s
}
