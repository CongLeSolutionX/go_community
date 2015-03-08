// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import "sync/atomic"

// fdMutex is a specialized synchronization primitive that manages
// lifetime of an fd and serializes access to Read, Write and Close
// methods on netFD.
type fdMutex struct {
	state uint64
	rsema uint32
	wsema uint32
}

// fdMutex.state is organized as follows:
// 1 bit - whether netFD is closed, if set all subsequent lock operations will fail.
// 1 bit - lock for read operations.
// 1 bit - lock for write operations.
// 20 bits - total number of references (read+write+misc).
// 20 bits - number of outstanding read waiters.
// 20 bits - number of outstanding write waiters.
const (
	mutexClosed  = 1 << 0
	mutexRLock   = 1 << 1
	mutexWLock   = 1 << 2
	mutexRef     = 1 << 3
	mutexRefMask = (1<<20 - 1) << 3
	mutexRWait   = 1 << 23
	mutexRMask   = (1<<20 - 1) << 23
	mutexWWait   = 1 << 43
	mutexWMask   = (1<<20 - 1) << 43
)

// fdMutexType represents a request type of mutex operation.
type fdMutexType int

const (
	fdRMutex fdMutexType = mutexRLock               // for read operations
	fdWMutex fdMutexType = mutexWLock               // for write operations
	fdOMutex fdMutexType = mutexClosed | mutexWLock // for open operations
)

// Read operations must do lock(fdRMutex)/unlock(fdRMutex).
//
// Write operations must do lock(fdWMutex)/unlock(fdWMutex).
//
// Open operations must do lock(fdOMutex)/unlock(fdOMutex).
//
// Misc operations must do incref/decref.
// Misc operations include functions like setsockopt and setDeadline.
// They need to use incref/decref to ensure that they operate on the
// correct fd in presence of a concurrent close call (otherwise fd can
// be closed under their feet).
//
// Close operations must do increfAndClose/decref.

// incref adds a reference to mu.
// It reports whether mu is available for reading or writing.
func (mu *fdMutex) incref() bool {
	for {
		old := atomic.LoadUint64(&mu.state)
		if old&mutexClosed != 0 {
			return false
		}
		new := old + mutexRef
		if new&mutexRefMask == 0 {
			panic("net: inconsistent fdMutex")
		}
		if atomic.CompareAndSwapUint64(&mu.state, old, new) {
			return true
		}
	}
}

// increfAndClose sets the state of mu to closed.
// It reports whether there is no remaining reference.
func (mu *fdMutex) increfAndClose() bool {
	for {
		old := atomic.LoadUint64(&mu.state)
		if old&mutexClosed != 0 {
			return false
		}
		// Mark as closed and acquire a reference.
		new := (old | mutexClosed) + mutexRef
		if new&mutexRefMask == 0 {
			panic("net: inconsistent fdMutex")
		}
		// Remove all read and write waiters.
		new &^= mutexRMask | mutexWMask
		if atomic.CompareAndSwapUint64(&mu.state, old, new) {
			// Wake all read and write waiters,
			// they will observe closed flag after wakeup.
			for old&mutexRMask != 0 {
				old -= mutexRWait
				runtime_Semrelease(&mu.rsema)
			}
			for old&mutexWMask != 0 {
				old -= mutexWWait
				runtime_Semrelease(&mu.wsema)
			}
			return true
		}
	}
}

// decref removes a reference from mu.
// It reports whether there is no remaining reference.
func (mu *fdMutex) decref() bool {
	for {
		old := atomic.LoadUint64(&mu.state)
		if old&mutexRefMask == 0 {
			panic("net: inconsistent fdMutex")
		}
		new := old - mutexRef
		if atomic.CompareAndSwapUint64(&mu.state, old, new) {
			return new&(mutexClosed|mutexRefMask) == mutexClosed
		}
	}
}

// lock adds a reference to mu and locks mu.
// It reports whether mu is available for reading or writing.
func (mu *fdMutex) lock(typ fdMutexType) bool {
	var mutexBit, mutexWait, mutexMask uint64
	var mutexSema *uint32
	if typ == fdRMutex {
		mutexBit = mutexRLock
		mutexWait = mutexRWait
		mutexMask = mutexRMask
		mutexSema = &mu.rsema
	} else {
		mutexBit = mutexWLock
		mutexWait = mutexWWait
		mutexMask = mutexWMask
		mutexSema = &mu.wsema
	}
	for {
		old := atomic.LoadUint64(&mu.state)
		// An open operation allows to take a write lock on
		// the closed-state mu.
		if old&mutexClosed != 0 && typ != fdOMutex {
			return false
		}
		var new uint64
		if old&mutexBit == 0 {
			// Lock is free, acquire it.
			new = (old | mutexBit) + mutexRef
			if new&mutexRefMask == 0 {
				panic("net: inconsistent fdMutex")
			}
		} else {
			// Wait for lock.
			new = old + mutexWait
			if new&mutexMask == 0 {
				panic("net: inconsistent fdMutex")
			}
		}
		if atomic.CompareAndSwapUint64(&mu.state, old, new) {
			if old&mutexBit == 0 {
				return new&mutexClosed == 0
			}
			runtime_Semacquire(mutexSema)
			// The signaller has subtracted mutexWait.
		}
	}
}

// unlock removes a reference from mu and unlocks mu.
// It reports whether there is no remaining reference.
func (mu *fdMutex) unlock(typ fdMutexType) bool {
	var mutexBit, mutexWait, mutexMask uint64
	var mutexSema *uint32
	if typ == fdRMutex {
		mutexBit = mutexRLock
		mutexWait = mutexRWait
		mutexMask = mutexRMask
		mutexSema = &mu.rsema
	} else {
		mutexBit = mutexWLock
		mutexWait = mutexWWait
		mutexMask = mutexWMask
		mutexSema = &mu.wsema
	}
	for {
		old := atomic.LoadUint64(&mu.state)
		if old&mutexBit == 0 || old&mutexRefMask == 0 {
			panic("net: inconsistent fdMutex")
		}
		// Drop lock, drop reference and wake a waiter if
		// present. An open operation makes the closed-state
		// mu ready for reading and writing.
		var new uint64
		if typ == fdOMutex {
			new = (old &^ (mutexClosed | mutexBit)) - mutexRef
		} else {
			new = (old &^ mutexBit) - mutexRef
		}
		if old&mutexMask != 0 {
			new -= mutexWait
		}
		if atomic.CompareAndSwapUint64(&mu.state, old, new) {
			if old&mutexMask != 0 {
				runtime_Semrelease(mutexSema)
			}
			return new&(mutexClosed|mutexRefMask) == mutexClosed
		}
	}
}

// Implemented in runtime package.
func runtime_Semacquire(sema *uint32)
func runtime_Semrelease(sema *uint32)

// incref adds a reference to fd.
// It returns an error when fd cannot be used.
func (fd *netFD) incref() error {
	if !fd.fdmu.incref() {
		return errClosing
	}
	return nil
}

// decref removes a reference from fd.
// It also closes fd when the state of fd is set to closed and there
// is no remaining reference.
func (fd *netFD) decref() {
	if fd.fdmu.decref() {
		fd.destroy()
	}
}

// openLock adds a reference to fd and locks fd for opening.
// It returns an error when fd cannot be used for writing.
func (fd *netFD) openLock() error {
	if !fd.fdmu.lock(fdOMutex) {
		return errClosing
	}
	return nil
}

// openUnlock removes a reference from fd and unlocks fd for reading
// and writing.
func (fd *netFD) openUnlock() {
	fd.fdmu.unlock(fdOMutex)
}

// readLock adds a reference to fd and locks fd for reading.
// It returns an error when fd cannot be used for reading.
func (fd *netFD) readLock() error {
	if !fd.fdmu.lock(fdRMutex) {
		return errClosing
	}
	return nil
}

// readUnlock removes a reference from fd and unlocks fd for reading.
// It also closes fd when the state of fd is set to closed and there
// is no remaining reference.
func (fd *netFD) readUnlock() {
	if fd.fdmu.unlock(fdRMutex) {
		fd.destroy()
	}
}

// writeLock adds a reference to fd and locks fd for writing.
// It returns an error when fd cannot be used for writing.
func (fd *netFD) writeLock() error {
	if !fd.fdmu.lock(fdWMutex) {
		return errClosing
	}
	return nil
}

// writeUnlock removes a reference from fd and unlocks fd for writing.
// It also closes fd when the state of fd is set to closed and there
// is no remaining reference.
func (fd *netFD) writeUnlock() {
	if fd.fdmu.unlock(fdWMutex) {
		fd.destroy()
	}
}
