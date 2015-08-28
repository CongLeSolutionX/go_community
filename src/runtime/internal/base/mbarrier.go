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

package base

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
//
// Dealing with memory ordering:
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
// Without an expensive memory barrier between the st and the ld, the final
// result on most HW (including 386/amd64) can be r1==r2==0. This is a classic
// example of what can happen when loads are allowed to be reordered with older
// stores (avoiding such reorderings lies at the heart of the classic
// Peterson/Dekker algorithms for mutual exclusion). Rather than require memory
// barriers, which will slow down both the mutator and the GC, we always grey
// the ptr object regardless of the slot's color.
//
// Another place where we intentionally omit memory barriers is when
// accessing mheap_.arena_used to check if a pointer points into the
// heap. On relaxed memory machines, it's possible for a mutator to
// extend the size of the heap by updating arena_used, allocate an
// object from this new region, and publish a pointer to that object,
// but for tracing running on another processor to observe the pointer
// but use the old value of arena_used. In this case, tracing will not
// mark the object, even though it's reachable. However, the mutator
// is guaranteed to execute a write barrier when it publishes the
// pointer, so it will take care of marking the object. A general
// consequence of this is that the garbage collector may cache the
// value of mheap_.arena_used. (See issue #9984.)
//
//
// Stack writes:
//
// The compiler omits write barriers for writes to the current frame,
// but if a stack pointer has been passed down the call stack, the
// compiler will generate a write barrier for writes through that
// pointer (because it doesn't know it's not a heap pointer).
//
// One might be tempted to ignore the write barrier if slot points
// into to the stack. Don't do it! Mark termination only re-scans
// frames that have potentially been active since the concurrent scan,
// so it depends on write barriers to track changes to pointers in
// stack frames that have not been active.
//go:nowritebarrier
func gcmarkwb_m(slot *uintptr, ptr uintptr) {
	if WriteBarrierEnabled {
		if ptr != 0 && Inheap(ptr) {
			shade(ptr)
		}
	}
}

// Write barrier calls must not happen during critical GC and scheduler
// related operations. In particular there are times when the GC assumes
// that the world is stopped but scheduler related code is still being
// executed, dealing with syscalls, dealing with putting gs on runnable
// queues and so forth. This code can not execute write barriers because
// the GC might drop them on the floor. Stopping the world involves removing
// the p associated with an m. We use the fact that m.p == nil to indicate
// that we are in one these critical section and throw if the write is of
// a pointer to a heap object.
//go:nosplit
func Writebarrierptr_nostore1(dst *uintptr, src uintptr) {
	mp := Acquirem()
	if mp.inwb || mp.dying > 0 {
		Releasem(mp)
		return
	}
	Systemstack(func() {
		if mp.P == 0 && Memstats.Enablegc && !mp.inwb && Inheap(src) {
			Throw("writebarrierptr_nostore1 called with mp.p == nil")
		}
		mp.inwb = true
		gcmarkwb_m(dst, src)
	})
	mp.inwb = false
	Releasem(mp)
}

// Like writebarrierptr, but the store has already been applied.
// Do not reapply.
//go:nosplit
func Writebarrierptr_nostore(dst *uintptr, src uintptr) {
	if !WriteBarrierEnabled {
		return
	}
	if src != 0 && (src < PhysPageSize || src == PoisonStack) {
		Systemstack(func() { Throw("bad pointer in write barrier") })
	}
	Writebarrierptr_nostore1(dst, src)
}
