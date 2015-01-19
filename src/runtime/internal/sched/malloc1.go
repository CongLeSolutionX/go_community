// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// See malloc.h for overview.
//
// TODO(rsc): double-check stats.

package sched

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	"unsafe"
)

const MaxArena32 = 2 << 30

func mHeap_SysAlloc(h *_lock.Mheap, n uintptr) unsafe.Pointer {
	if n > uintptr(h.Arena_end)-uintptr(h.Arena_used) {
		// We are in 32-bit mode, maybe we didn't use all possible address space yet.
		// Reserve some more space.
		p_size := Round(n+_core.PageSize, 256<<20)
		new_end := h.Arena_end + p_size
		if new_end <= h.Arena_start+MaxArena32 {
			// TODO: It would be bad if part of the arena
			// is reserved and part is not.
			var reserved bool
			p := uintptr(SysReserve((unsafe.Pointer)(h.Arena_end), p_size, &reserved))
			if p == h.Arena_end {
				h.Arena_end = new_end
				h.Arena_reserved = reserved
			} else if p+p_size <= h.Arena_start+MaxArena32 {
				// Keep everything page-aligned.
				// Our pages are bigger than hardware pages.
				h.Arena_end = p + p_size
				h.Arena_used = p + (-uintptr(p) & (_core.PageSize - 1))
				h.Arena_reserved = reserved
			} else {
				var stat uint64
				SysFree((unsafe.Pointer)(p), p_size, &stat)
			}
		}
	}

	if n <= uintptr(h.Arena_end)-uintptr(h.Arena_used) {
		// Keep taking from our reservation.
		p := h.Arena_used
		sysMap((unsafe.Pointer)(p), n, h.Arena_reserved, &_lock.Memstats.Heap_sys)
		h.Arena_used += n
		mHeap_MapBits(h)
		mHeap_MapSpans(h)
		if Raceenabled {
			racemapshadow((unsafe.Pointer)(p), n)
		}

		if uintptr(p)&(_core.PageSize-1) != 0 {
			_lock.Gothrow("misrounded allocation in MHeap_SysAlloc")
		}
		return (unsafe.Pointer)(p)
	}

	// If using 64-bit, our reservation is all we have.
	if uintptr(h.Arena_end)-uintptr(h.Arena_start) >= MaxArena32 {
		return nil
	}

	// On 32-bit, once the reservation is gone we can
	// try to get memory at a location chosen by the OS
	// and hope that it is in the range we allocated bitmap for.
	p_size := Round(n, _core.PageSize) + _core.PageSize
	p := uintptr(_lock.SysAlloc(p_size, &_lock.Memstats.Heap_sys))
	if p == 0 {
		return nil
	}

	if p < h.Arena_start || uintptr(p)+p_size-uintptr(h.Arena_start) >= MaxArena32 {
		print("runtime: memory allocated by OS (", p, ") not in usable range [", _core.Hex(h.Arena_start), ",", _core.Hex(h.Arena_start+MaxArena32), ")\n")
		SysFree((unsafe.Pointer)(p), p_size, &_lock.Memstats.Heap_sys)
		return nil
	}

	p_end := p + p_size
	p += -p & (_core.PageSize - 1)
	if uintptr(p)+n > uintptr(h.Arena_used) {
		h.Arena_used = p + n
		if p_end > h.Arena_end {
			h.Arena_end = p_end
		}
		mHeap_MapBits(h)
		mHeap_MapSpans(h)
		if Raceenabled {
			racemapshadow((unsafe.Pointer)(p), n)
		}
	}

	if uintptr(p)&(_core.PageSize-1) != 0 {
		_lock.Gothrow("misrounded allocation in MHeap_SysAlloc")
	}
	return (unsafe.Pointer)(p)
}
