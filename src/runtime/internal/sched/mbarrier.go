// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Garbage collector: write barriers.
//
// For the concurrent garbage collector, the Go compiler implements
// updates to pointer-valued fields that may be in heap objects by
// emitting calls to write barriers. This file contains the actual write barrier
// implementation, markwb, and the various wrappers called by the
// compiler to implement pointer assignment, slice assignment,
// typed memmove, and so on.
//
// To check for missed write barriers, the GODEBUG=wbshadow debugging
// mode allocates a second copy of the heap. Write barrier-based pointer
// updates make changes to both the real heap and the shadow, and both
// the pointer updates and the GC look for inconsistencies between the two,
// indicating pointer writes that bypassed the barrier.

package sched

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	"unsafe"
)

// markwb is the mark-phase write barrier, the only barrier we have.
// The rest of this file exists only to make calls to this function.
//
// This is the Dijkstra barrier coarsened to always shade the ptr (dst) object.
// The original Dijkstra barrier only shaded ptrs being placed in black slots.
//
// Shade indicates that it has seen a white pointer by adding the referent
// to wbuf as well as marking it.
//
// slot is the destination (dst) in go code
// ptr is the value that goes into the slot (src) in the go code
//
// Dijkstra pointed out that maintaining the no black to white
// pointers means that white to white pointers not need
// to be noted by the write barrier. Furthermore if either
// white object dies before it is reached by the
// GC then the object can be collected during this GC cycle
// instead of waiting for the next cycle. Unfortunately the cost of
// ensure that the object holding the slot doesn't concurrently
// change to black without the mutator noticing seems prohibitive.
//
// Consider the following example where the mutator writes into
// a slot and then loads the slot's mark bit while the GC thread
// writes to the slot's mark bit and then as part of scanning reads
// the slot.
//
// Initially both [slot] and [slotmark] are 0 (nil)
// Mutator thread          GC thread
// st [slot], ptr          st [slotmark], 1
//
// ld r1, [slotmark]       ld r2, [slot]
//
// This is a classic example of independent reads of independent writes,
// aka IRIW. The question is if r1==r2==0 is allowed and for most HW the
// answer is yes without inserting a memory barriers between the st and the ld.
// These barriers are expensive so we have decided that we will
// always grey the ptr object regardless of the slot's color.
//go:nowritebarrier
func gcmarkwb_m(slot *uintptr, ptr uintptr) {
	switch Gcphase {
	default:
		_lock.Throw("gcphasework in bad gcphase")

	case GCoff, GCquiesce, GCstw, GCsweep, GCscan:
		// ok

	case GCmark, GCmarktermination:
		if ptr != 0 && inheap(ptr) {
			shade(ptr)
		}
	}
}

// needwb reports whether a write barrier is needed now
// (otherwise the write can be made directly).
//go:nosplit
func Needwb() bool {
	return Gcphase == GCmark || Gcphase == GCmarktermination || _lock.Mheap_.Shadow_enabled
}

//go:nosplit
func Writebarrierptr_nostore1(dst *uintptr, src uintptr) {
	mp := Acquirem()
	if mp.Inwb || mp.Dying > 0 {
		Releasem(mp)
		return
	}
	mp.Inwb = true
	_lock.Systemstack(func() {
		gcmarkwb_m(dst, src)
	})
	mp.Inwb = false
	Releasem(mp)
}

// Like writebarrierptr, but the store has already been applied.
// Do not reapply.
//go:nosplit
func Writebarrierptr_nostore(dst *uintptr, src uintptr) {
	if !Needwb() {
		return
	}

	if src != 0 && (src < _lock.PhysPageSize || src == _lock.PoisonStack) {
		_lock.Systemstack(func() { _lock.Throw("bad pointer in write barrier") })
	}

	// Apply changes to shadow.
	// Since *dst has been overwritten already, we cannot check
	// whether there were any missed updates, but writebarrierptr_nostore
	// is only rarely used.
	if _lock.Mheap_.Shadow_enabled {
		_lock.Systemstack(func() {
			addr := uintptr(unsafe.Pointer(dst))
			shadow := Shadowptr(addr)
			if shadow == nil {
				return
			}
			*shadow = src
		})
	}

	Writebarrierptr_nostore1(dst, src)
}

// writebarrierptr_noshadow records that the value in *dst
// has been written to using an atomic operation and the shadow
// has not been updated. (In general if dst must be manipulated
// atomically we cannot get the right bits for use in the shadow.)
//go:nosplit
func Writebarrierptr_noshadow(dst *uintptr) {
	addr := uintptr(unsafe.Pointer(dst))
	shadow := Shadowptr(addr)
	if shadow == nil {
		return
	}

	*shadow = NoShadow
}

// Shadow heap for detecting missed write barriers.

// noShadow is stored in as the shadow pointer to mark that there is no
// shadow word recorded. It matches any actual pointer word.
// noShadow is used when it is impossible to know the right word
// to store in the shadow heap, such as when the real heap word
// is being manipulated atomically.
const NoShadow uintptr = 1

// shadowptr returns a pointer to the shadow value for addr.
//go:nosplit
func Shadowptr(addr uintptr) *uintptr {
	var shadow *uintptr
	if _lock.Mheap_.Data_start <= addr && addr < _lock.Mheap_.Data_end {
		shadow = (*uintptr)(unsafe.Pointer(addr + _lock.Mheap_.Shadow_data))
	} else if inheap(addr) {
		shadow = (*uintptr)(unsafe.Pointer(addr + _lock.Mheap_.Shadow_heap))
	}
	return shadow
}

// istrackedptr reports whether the pointer value p requires a write barrier
// when stored into the heap.
func Istrackedptr(p uintptr) bool {
	return inheap(p)
}

// checkwbshadow checks that p matches its shadow word.
// The garbage collector calls checkwbshadow for each pointer during the checkmark phase.
// It is only called when mheap_.shadow_enabled is true.
func checkwbshadow(p *uintptr) {
	addr := uintptr(unsafe.Pointer(p))
	shadow := Shadowptr(addr)
	if shadow == nil {
		return
	}
	// There is no race on the accesses here, because the world is stopped,
	// but there may be racy writes that lead to the shadow and the
	// heap being inconsistent. If so, we will detect that here as a
	// missed write barrier and crash. We don't mind.
	// Code should use sync/atomic instead of racy pointer writes.
	if *shadow != *p && *shadow != NoShadow && Istrackedptr(*p) {
		_lock.Mheap_.Shadow_enabled = false
		print("runtime: checkwritebarrier p=", p, " *p=", _core.Hex(*p), " shadow=", shadow, " *shadow=", _core.Hex(*shadow), "\n")
		_lock.Throw("missed write barrier")
	}
}
