// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package base

import (
	"unsafe"
)

// Puts the current goroutine into a waiting state and calls unlockf.
// If unlockf returns false, the goroutine is resumed.
func Gopark(unlockf func(*G, unsafe.Pointer) bool, lock unsafe.Pointer, reason string, traceEv byte, traceskip int) {
	mp := Acquirem()
	gp := mp.Curg
	status := Readgstatus(gp)
	if status != Grunning && status != Gscanrunning {
		Throw("gopark: bad g status")
	}
	mp.waitlock = lock
	mp.waitunlockf = *(*unsafe.Pointer)(unsafe.Pointer(&unlockf))
	gp.Waitreason = reason
	mp.waittraceev = traceEv
	mp.waittraceskip = traceskip
	Releasem(mp)
	// can't do anything that might move the G between Ms here.
	Mcall(park_m)
}

// Puts the current goroutine into a waiting state and unlocks the lock.
// The goroutine can be made runnable again by calling goready(gp).
func Goparkunlock(lock *Mutex, reason string, traceEv byte, traceskip int) {
	Gopark(parkunlock_c, unsafe.Pointer(lock), reason, traceEv, traceskip)
}

func Goready(gp *G, traceskip int) {
	Systemstack(func() {
		Ready(gp, traceskip)
	})
}

// funcPC returns the entry PC of the function f.
// It assumes that f is a func value. Otherwise the behavior is undefined.
//go:nosplit
func FuncPC(f interface{}) uintptr {
	return **(**uintptr)(Add(unsafe.Pointer(&f), PtrSize))
}

var (
	Allgs    []*G
	Allglock Mutex
)

func Allgadd(gp *G) {
	if Readgstatus(gp) == Gidle {
		Throw("allgadd: bad status Gidle")
	}

	Lock(&Allglock)
	Allgs = append(Allgs, gp)
	allg = &Allgs[0]
	Allglen = uintptr(len(Allgs))
	Unlock(&Allglock)
}
