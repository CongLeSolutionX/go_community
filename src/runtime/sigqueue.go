// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file implements runtime support for signal handling.
//
// Most synchronization primitives are not available from
// the signal handler (it cannot block, allocate memory, or use locks)
// so the handler communicates with a processing goroutine
// via struct sig, below.
//
// sigsend is called by the signal handler to queue a new signal.
// signal_recv is called by the Go program to receive a newly queued signal.
// Synchronization between sigsend and signal_recv is based on the sig.state
// variable.  It can be in 3 states: sigIdle, sigReceiving and sigSending.
// sigReceiving means that signal_recv is blocked on sig.Note and there are no
// new pending signals.
// sigSending means that sig.mask *may* contain new pending signals,
// signal_recv can't be blocked in this state.
// sigIdle means that there are no new pending signals and signal_recv is not blocked.
// Transitions between states are done atomically with CAS.
// When signal_recv is unblocked, it resets sig.Note and rechecks sig.mask.
// If several sigsends and signal_recv execute concurrently, it can lead to
// unnecessary rechecks of sig.mask, but it cannot lead to missed signals
// nor deadlocks.

// +build !plan9

package runtime

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
	"unsafe"
)

// Called to receive the next queued signal.
// Must only be called from a single goroutine at a time.
func signal_recv() uint32 {
	for {
		// Serve any signals from local copy.
		for i := uint32(0); i < _core.NSIG; i++ {
			if _sched.Sig.Recv[i/32]&(1<<(i&31)) != 0 {
				_sched.Sig.Recv[i/32] &^= 1 << (i & 31)
				return i
			}
		}

		// Wait for updates to be available from signal sender.
	Receive:
		for {
			switch _lock.Atomicload(&_sched.Sig.State) {
			default:
				_lock.Throw("signal_recv: inconsistent state")
			case _sched.SigIdle:
				if _sched.Cas(&_sched.Sig.State, _sched.SigIdle, _sched.SigReceiving) {
					_sched.Notetsleepg(&_sched.Sig.Note, -1)
					_sched.Noteclear(&_sched.Sig.Note)
					break Receive
				}
			case _sched.SigSending:
				if _sched.Cas(&_sched.Sig.State, _sched.SigSending, _sched.SigIdle) {
					break Receive
				}
			}
		}

		// Incorporate updates from sender into local copy.
		for i := range _sched.Sig.Mask {
			_sched.Sig.Recv[i] = xchg(&_sched.Sig.Mask[i], 0)
		}
	}
}

// Must only be called from a single goroutine at a time.
func signal_enable(s uint32) {
	if !_sched.Sig.Inuse {
		// The first call to signal_enable is for us
		// to use for initialization.  It does not pass
		// signal information in m.
		_sched.Sig.Inuse = true // enable reception of signals; cannot disable
		_sched.Noteclear(&_sched.Sig.Note)
		return
	}

	if int(s) >= len(_sched.Sig.Wanted)*32 {
		return
	}
	_sched.Sig.Wanted[s/32] |= 1 << (s & 31)
	sigenable(s)
}

// Must only be called from a single goroutine at a time.
func signal_disable(s uint32) {
	if int(s) >= len(_sched.Sig.Wanted)*32 {
		return
	}
	_sched.Sig.Wanted[s/32] &^= 1 << (s & 31)
	sigdisable(s)
}

// This runs on a foreign stack, without an m or a g.  No stack split.
//go:nosplit
func badsignal(sig uintptr) {
	// Some external libraries, for example, OpenBLAS, create worker threads in
	// a global constructor. If we're doing cpu profiling, and the SIGPROF signal
	// comes to one of the foreign threads before we make our first cgo call, the
	// call to cgocallback below will bring down the whole process.
	// It's better to miss a few SIGPROF signals than to abort in this case.
	// See http://golang.org/issue/9456.
	if _lock.SIGPROF != 0 && sig == _lock.SIGPROF && _core.Needextram != 0 {
		return
	}
	cgocallback(unsafe.Pointer(_lock.FuncPC(_sched.Sigsend)), _core.Noescape(unsafe.Pointer(&sig)), unsafe.Sizeof(sig))
}
