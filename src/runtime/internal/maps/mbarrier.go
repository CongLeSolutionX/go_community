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

package maps

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
	"unsafe"
)

// clearshadow clears the shadow copy associated with the n bytes of memory at addr.
func Clearshadow(addr, n uintptr) {
	if !_lock.Mheap_.Shadow_enabled {
		return
	}
	p := _sched.Shadowptr(addr)
	if p == nil || n <= _core.PtrSize {
		return
	}
	_core.Memclr(unsafe.Pointer(p), n)
}
