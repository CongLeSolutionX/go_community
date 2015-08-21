// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Garbage collector: sweeping

package iface

import (
	_base "runtime/internal/base"
	_gc "runtime/internal/gc"
)

// deductSweepCredit deducts sweep credit for allocating a span of
// size spanBytes. This must be performed *before* the span is
// allocated to ensure the system has enough credit. If necessary, it
// performs sweeping to prevent going in to debt. If the caller will
// also sweep pages (e.g., for a large allocation), it can pass a
// non-zero callerSweepPages to leave that many pages unswept.
//
// deductSweepCredit is the core of the "proportional sweep" system.
// It uses statistics gathered by the garbage collector to perform
// enough sweeping so that all pages are swept during the concurrent
// sweep phase between GC cycles.
//
// mheap_ must NOT be locked.
func deductSweepCredit(spanBytes uintptr, callerSweepPages uintptr) {
	if _base.Mheap_.SweepPagesPerByte == 0 {
		// Proportional sweep is done or disabled.
		return
	}

	// Account for this span allocation.
	spanBytesAlloc := _base.Xadd64(&_base.Mheap_.SpanBytesAlloc, int64(spanBytes))

	// Fix debt if necessary.
	pagesOwed := int64(_base.Mheap_.SweepPagesPerByte * float64(spanBytesAlloc))
	for pagesOwed-int64(_base.Atomicload64(&_base.Mheap_.PagesSwept)) > int64(callerSweepPages) {
		if _gc.Gosweepone() == ^uintptr(0) {
			_base.Mheap_.SweepPagesPerByte = 0
			break
		}
	}
}
