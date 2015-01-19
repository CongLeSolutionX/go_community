// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Page heap.
//
// See malloc.h for overview.
//
// When a MSpan is in the heap free list, state == MSpanFree
// and heapmap(s->start) == span, heapmap(s->start+s->npages-1) == span.
//
// When a MSpan is allocated, state == MSpanInUse or MSpanStack
// and heapmap(i) == span for all s->start <= i < s->start+s->npages.

package runtime

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
	"unsafe"
)

func scavengelist(list *_core.Mspan, now, limit uint64) uintptr {
	if _sched.MSpanList_IsEmpty(list) {
		return 0
	}

	var sumreleased uintptr
	for s := list.Next; s != list; s = s.Next {
		if (now-uint64(s.Unusedsince)) > limit && s.Npreleased != s.Npages {
			released := (s.Npages - s.Npreleased) << _core.PageShift
			_lock.Memstats.Heap_released += uint64(released)
			sumreleased += released
			s.Npreleased = s.Npages
			sysUnused((unsafe.Pointer)(s.Start<<_core.PageShift), s.Npages<<_core.PageShift)
		}
	}
	return sumreleased
}

func mHeap_Scavenge(k int32, now, limit uint64) {
	h := &_lock.Mheap_
	_lock.Lock(&h.Lock)
	var sumreleased uintptr
	for i := 0; i < len(h.Free); i++ {
		sumreleased += scavengelist(&h.Free[i], now, limit)
	}
	sumreleased += scavengelist(&h.Freelarge, now, limit)
	_lock.Unlock(&h.Lock)

	if _lock.Debug.Gctrace > 0 {
		if sumreleased > 0 {
			print("scvg", k, ": ", sumreleased>>20, " MB released\n")
		}
		// TODO(dvyukov): these stats are incorrect as we don't subtract stack usage from heap.
		// But we can't call ReadMemStats on g0 holding locks.
		print("scvg", k, ": inuse: ", _lock.Memstats.Heap_inuse>>20, ", idle: ", _lock.Memstats.Heap_idle>>20, ", sys: ", _lock.Memstats.Heap_sys>>20, ", released: ", _lock.Memstats.Heap_released>>20, ", consumed: ", (_lock.Memstats.Heap_sys-_lock.Memstats.Heap_released)>>20, " (MB)\n")
	}
}

func scavenge_m() {
	mHeap_Scavenge(-1, ^uint64(0), 0)
}
