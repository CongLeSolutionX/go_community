// Copyright 2011 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin nacl netbsd openbsd plan9 solaris windows

package sched

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	"unsafe"
)

// One-time notifications.
func Noteclear(n *_core.Note) {
	n.Key = 0
}

func Notewakeup(n *_core.Note) {
	var v uintptr
	for {
		v = _core.Atomicloaduintptr(&n.Key)
		if _core.Casuintptr(&n.Key, v, _lock.Locked) {
			break
		}
	}

	// Successfully set waitm to locked.
	// What was it before?
	switch {
	case v == 0:
		// Nothing was waiting. Done.
	case v == _lock.Locked:
		// Two notewakeups!  Not allowed.
		_lock.Throw("notewakeup - double wakeup")
	default:
		// Must be the waiting m.  Wake it up.
		_lock.Semawakeup((*_core.M)(unsafe.Pointer(v)))
	}
}

func Notesleep(n *_core.Note) {
	gp := _core.Getg()
	if gp != gp.M.G0 {
		_lock.Throw("notesleep not on g0")
	}
	if gp.M.Waitsema == 0 {
		gp.M.Waitsema = _lock.Semacreate()
	}
	if !_core.Casuintptr(&n.Key, 0, uintptr(unsafe.Pointer(gp.M))) {
		// Must be locked (got wakeup).
		if n.Key != _lock.Locked {
			_lock.Throw("notesleep - waitm out of sync")
		}
		return
	}
	// Queued.  Sleep.
	gp.M.Blocked = true
	_lock.Semasleep(-1)
	gp.M.Blocked = false
}

//go:nosplit
func Notetsleep_internal(n *_core.Note, ns int64, gp *_core.G, deadline int64) bool {
	// gp and deadline are logically local variables, but they are written
	// as parameters so that the stack space they require is charged
	// to the caller.
	// This reduces the nosplit footprint of notetsleep_internal.
	gp = _core.Getg()

	// Register for wakeup on n->waitm.
	if !_core.Casuintptr(&n.Key, 0, uintptr(unsafe.Pointer(gp.M))) {
		// Must be locked (got wakeup).
		if n.Key != _lock.Locked {
			_lock.Throw("notetsleep - waitm out of sync")
		}
		return true
	}
	if ns < 0 {
		// Queued.  Sleep.
		gp.M.Blocked = true
		_lock.Semasleep(-1)
		gp.M.Blocked = false
		return true
	}

	deadline = _lock.Nanotime() + ns
	for {
		// Registered.  Sleep.
		gp.M.Blocked = true
		if _lock.Semasleep(ns) >= 0 {
			gp.M.Blocked = false
			// Acquired semaphore, semawakeup unregistered us.
			// Done.
			return true
		}
		gp.M.Blocked = false
		// Interrupted or timed out.  Still registered.  Semaphore not acquired.
		ns = deadline - _lock.Nanotime()
		if ns <= 0 {
			break
		}
		// Deadline hasn't arrived.  Keep sleeping.
	}

	// Deadline arrived.  Still registered.  Semaphore not acquired.
	// Want to give up and return, but have to unregister first,
	// so that any notewakeup racing with the return does not
	// try to grant us the semaphore when we don't expect it.
	for {
		v := _core.Atomicloaduintptr(&n.Key)
		switch v {
		case uintptr(unsafe.Pointer(gp.M)):
			// No wakeup yet; unregister if possible.
			if _core.Casuintptr(&n.Key, v, 0) {
				return false
			}
		case _lock.Locked:
			// Wakeup happened so semaphore is available.
			// Grab it to avoid getting out of sync.
			gp.M.Blocked = true
			if _lock.Semasleep(-1) < 0 {
				_lock.Throw("runtime: unable to acquire - semaphore out of sync")
			}
			gp.M.Blocked = false
			return true
		default:
			_lock.Throw("runtime: unexpected waitm - semaphore out of sync")
		}
	}
}

// same as runtime·notetsleep, but called on user g (not g0)
// calls only nosplit functions between entersyscallblock/exitsyscall
func Notetsleepg(n *_core.Note, ns int64) bool {
	gp := _core.Getg()
	if gp == gp.M.G0 {
		_lock.Throw("notetsleepg on g0")
	}
	if gp.M.Waitsema == 0 {
		gp.M.Waitsema = _lock.Semacreate()
	}
	entersyscallblock(0)
	ok := Notetsleep_internal(n, ns, nil, 0)
	Exitsyscall(0)
	return ok
}
