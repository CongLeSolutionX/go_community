// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sched

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	"unsafe"
)

// Puts the current goroutine into a waiting state and calls unlockf.
// If unlockf returns false, the goroutine is resumed.
func Gopark(unlockf func(*_core.G, unsafe.Pointer) bool, lock unsafe.Pointer, reason string, traceEv byte) {
	mp := Acquirem()
	gp := mp.Curg
	status := _lock.Readgstatus(gp)
	if status != _lock.Grunning && status != _lock.Gscanrunning {
		_lock.Throw("gopark: bad g status")
	}
	mp.Waitlock = lock
	mp.Waitunlockf = *(*unsafe.Pointer)(unsafe.Pointer(&unlockf))
	gp.Waitreason = reason
	mp.Waittraceev = traceEv
	Releasem(mp)
	// can't do anything that might move the G between Ms here.
	Mcall(Park_m)
}

// Puts the current goroutine into a waiting state and unlocks the lock.
// The goroutine can be made runnable again by calling goready(gp).
func Goparkunlock(lock *_core.Mutex, reason string, traceEv byte) {
	Gopark(Parkunlock_c, unsafe.Pointer(lock), reason, traceEv)
}

func Goready(gp *_core.G) {
	_lock.Systemstack(func() {
		Ready(gp)
	})
}
