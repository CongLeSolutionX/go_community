// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	"unsafe"
)

var (
	M0 M
	g0 G
)

// needm is called when a cgo callback happens on a
// thread without an m (a thread not created by Go).
// In this case, needm is expected to find an m to use
// and return with m, g initialized correctly.
// Since m and g are not set now (likely nil, but see below)
// needm is limited in what routines it can call. In particular
// it can only call nosplit functions (textflag 7) and cannot
// do any scheduling that requires an m.
//
// In order to avoid needing heavy lifting here, we adopt
// the following strategy: there is a stack of available m's
// that can be stolen. Using compare-and-swap
// to pop from the stack has ABA races, so we simulate
// a lock by doing an exchange (via casp) to steal the stack
// head and replace the top pointer with MLOCKED (1).
// This serves as a simple spin lock that we can use even
// without an m. The thread that locks the stack in this way
// unlocks the stack by storing a valid stack head pointer.
//
// In order to make sure that there is always an m structure
// available to be stolen, we maintain the invariant that there
// is always one more than needed. At the beginning of the
// program (if cgo is in use) the list is seeded with a single m.
// If needm finds that it has taken the last m off the list, its job
// is - once it has installed its own m so that it can do things like
// allocate memory - to create a spare m and put it on the list.
//
// Each of these extra m's also has a g0 and a curg that are
// pressed into service as the scheduling stack and current
// goroutine for the duration of the cgo callback.
//
// When the callback is done with the m, it calls dropm to
// put the m back on the list.
//go:nosplit
func needm(x byte) {
	if Needextram != 0 {
		// Can happen if C/C++ code calls Go from a global ctor.
		// Can not throw, because scheduler is not initialized yet.
		Write(2, unsafe.Pointer(&earlycgocallback[0]), int32(len(earlycgocallback)))
		Exit(1)
	}

	// Lock extra list, take head, unlock popped list.
	// nilokay=false is safe here because of the invariant above,
	// that the extra list always contains or will soon contain
	// at least one m.
	mp := Lockextra(false)

	// Set needextram when we've just emptied the list,
	// so that the eventual call into cgocallbackg will
	// allocate a new m for the extra list. We delay the
	// allocation until then so that it can be done
	// after exitsyscall makes sure it is okay to be
	// running at all (that is, there's no garbage collection
	// running right now).
	mp.Needextram = mp.Schedlink == nil
	Unlockextra(mp.Schedlink)

	// Install g (= m->g0) and set the stack bounds
	// to match the current stack. We don't actually know
	// how big the stack is, like we don't know how big any
	// scheduling stack is, but we assume there's at least 32 kB,
	// which is more than enough for us.
	Setg(mp.G0)
	_g_ := Getg()
	_g_.Stack.Hi = uintptr(Noescape(unsafe.Pointer(&x))) + 1024
	_g_.Stack.Lo = uintptr(Noescape(unsafe.Pointer(&x))) - 32*1024
	_g_.Stackguard0 = _g_.Stack.Lo + StackGuard

	// Initialize this thread to use the m.
	Asminit()
	Minit()
}

var earlycgocallback = []byte("fatal error: cgo callback before cgo call\n")

var extram uintptr

// lockextra locks the extra list and returns the list head.
// The caller must unlock the list by storing a new list head
// to extram. If nilokay is true, then lockextra will
// return a nil list head if that's what it finds. If nilokay is false,
// lockextra will keep waiting until the list head is no longer nil.
//go:nosplit
func Lockextra(nilokay bool) *M {
	const locked = 1

	for {
		old := Atomicloaduintptr(&extram)
		if old == locked {
			yield := Osyield
			yield()
			continue
		}
		if old == 0 && !nilokay {
			Usleep(1)
			continue
		}
		if Casuintptr(&extram, old, locked) {
			return (*M)(unsafe.Pointer(old))
		}
		yield := Osyield
		yield()
		continue
	}
}

//go:nosplit
func Unlockextra(mp *M) {
	atomicstoreuintptr(&extram, uintptr(unsafe.Pointer(mp)))
}
