// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris windows

package base

import (
	"unsafe"
)

// Integrated network poller (platform-independent part).
// A particular implementation (epoll/kqueue) must define the following functions:
// func netpollinit()			// to initialize the poller
// func netpollopen(fd uintptr, pd *pollDesc) int32	// to arm edge-triggered notifications
// and associate fd with pd.
// An implementation must call the following function to denote that the pd is ready.
// func netpollready(gpp **g, pd *pollDesc, mode int32)

// pollDesc contains 2 binary semaphores, rg and wg, to park reader and writer
// goroutines respectively. The semaphore can be in the following states:
// pdReady - io readiness notification is pending;
//           a goroutine consumes the notification by changing the state to nil.
// pdWait - a goroutine prepares to park on the semaphore, but not yet parked;
//          the goroutine commits to park by changing the state to G pointer,
//          or, alternatively, concurrent io notification changes the state to READY,
//          or, alternatively, concurrent timeout/close changes the state to nil.
// G pointer - the goroutine is blocked on the semaphore;
//             io notification or timeout/close changes the state to READY or nil respectively
//             and unparks the goroutine.
// nil - nothing of the above.
const (
	PdReady uintptr = 1
	PdWait  uintptr = 2
)

// Network poller descriptor.
type PollDesc struct {
	Link *PollDesc // in pollcache, protected by pollcache.lock

	// The lock protects pollOpen, pollSetDeadline, pollUnblock and deadlineimpl operations.
	// This fully covers seq, rt and wt variables. fd is constant throughout the PollDesc lifetime.
	// pollReset, pollWait, pollWaitCanceled and runtime·netpollready (IO readiness notification)
	// proceed w/o taking the lock. So closing, rg, rd, wg and wd are manipulated
	// in a lock-free way by all operations.
	// NOTE(dvyukov): the following code uses uintptr to store *g (rg/wg),
	// that will blow up when GC starts moving objects.
	Lock    Mutex // protects the following fields
	Fd      uintptr
	Closing bool
	Seq     uintptr // protects from stale timers and ready notifications
	Rg      uintptr // pdReady, pdWait, G waiting for read or nil
	Rt      Timer   // read deadline timer (set if rt.f != nil)
	Rd      int64   // read deadline
	Wg      uintptr // pdReady, pdWait, G waiting for write or nil
	Wt      Timer   // write deadline timer
	Wd      int64   // write deadline
	user    uint32  // user settable cookie
}

var (
	NetpollInited uint32
)

func netpollinited() bool {
	return Atomicload(&NetpollInited) != 0
}

// make pd ready, newly runnable goroutines (if any) are returned in rg/wg
// May run during STW, so write barriers are not allowed.
//go:nowritebarrier
func netpollready(gpp *Guintptr, pd *PollDesc, mode int32) {
	var rg, wg Guintptr
	if mode == 'r' || mode == 'r'+'w' {
		rg.Set(Netpollunblock(pd, 'r', true))
	}
	if mode == 'w' || mode == 'r'+'w' {
		wg.Set(Netpollunblock(pd, 'w', true))
	}
	if rg != 0 {
		rg.Ptr().Schedlink = *gpp
		*gpp = rg
	}
	if wg != 0 {
		wg.Ptr().Schedlink = *gpp
		*gpp = wg
	}
}

func Netpollunblock(pd *PollDesc, mode int32, ioready bool) *G {
	gpp := &pd.Rg
	if mode == 'w' {
		gpp = &pd.Wg
	}

	for {
		old := *gpp
		if old == PdReady {
			return nil
		}
		if old == 0 && !ioready {
			// Only set READY for ioready. runtime_pollWait
			// will check for timeout/cancel before waiting.
			return nil
		}
		var new uintptr
		if ioready {
			new = PdReady
		}
		if Casuintptr(gpp, old, new) {
			if old == PdReady || old == PdWait {
				old = 0
			}
			return (*G)(unsafe.Pointer(old))
		}
	}
}
