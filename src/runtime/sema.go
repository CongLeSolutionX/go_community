// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Semaphore implementation exposed to Go.
// Intended use is provide a sleep and wakeup
// primitive that can be used in the contended case
// of other synchronization primitives.
// Thus it targets the same goal as Linux's futex,
// but it has much simpler semantics.
//
// That is, don't think of these as semaphores.
// Think of them as a way to implement sleep and wakeup
// such that every sleep is paired with a single wakeup,
// even if, due to races, the wakeup happens before the sleep.
//
// See Mullender and Cox, ``Semaphores in Plan 9,''
// http://swtch.com/semaphore.pdf

package runtime

import (
	_base "runtime/internal/base"
	_lock "runtime/internal/lock"
	"unsafe"
)

//go:linkname sync_runtime_Semacquire sync.runtime_Semacquire
func sync_runtime_Semacquire(addr *uint32) {
	_lock.Semacquire(addr, true)
}

//go:linkname net_runtime_Semacquire net.runtime_Semacquire
func net_runtime_Semacquire(addr *uint32) {
	_lock.Semacquire(addr, true)
}

//go:linkname sync_runtime_Semrelease sync.runtime_Semrelease
func sync_runtime_Semrelease(addr *uint32) {
	_lock.Semrelease(addr)
}

//go:linkname net_runtime_Semrelease net.runtime_Semrelease
func net_runtime_Semrelease(addr *uint32) {
	_lock.Semrelease(addr)
}

//go:linkname syncsemcheck sync.runtime_Syncsemcheck
func syncsemcheck(sz uintptr) {
	if sz != unsafe.Sizeof(_lock.SyncSema{}) {
		print("runtime: bad syncSema size - sync=", sz, " runtime=", unsafe.Sizeof(_lock.SyncSema{}), "\n")
		_base.Throw("bad syncSema size")
	}
}
