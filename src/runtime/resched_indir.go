// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows
// +build amd64

// This implements indirect fault-based loop preemption. This is
// similar to resched_direct.go, but rather than loading directly from
// a location in the BSS, the preemption check loads the address of
// the fault location from a global and then dereferences it.
//
// On amd64 and 386, this requires two instructions and a register
// clobber. However, this indirection is necessary on Windows because
// it doesn't let us unmap pages from the BSS.

package runtime

import "unsafe"

// reschedulePage is the address of a page that will be unmapped to
// interrupt preemptible loops.
var reschedulePage unsafe.Pointer

func reschedinit() {
	var dummyStat uint64
	reschedulePage = sysAlloc(physPageSize, &dummyStat)
	if reschedulePage == nil {
		throw("failed to allocate rescheduling page")
	}
}

// reschedFaultAddr returns the address at which preemptible loops
// will fault.
func reschedFaultAddr() uintptr {
	return uintptr(reschedulePage)
}

func preemptLoops() {
	// Force loop preemption by unmapping the rescheduling page.
	sysFault(reschedulePage, physPageSize)
}

func unpreemptLoops() {
	// Remap the loop preemption page.
	var dummyStat uint64
	sysMap(reschedulePage, physPageSize, true, &dummyStat)
}
