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
	_base "runtime/internal/base"
	"unsafe"
)

// Called to receive the next queued signal.
// Must only be called from a single goroutine at a time.
func signal_recv() uint32 {
	for {
		// Serve any signals from local copy.
		for i := uint32(0); i < _base.NSIG; i++ {
			if _base.Sig.Recv[i/32]&(1<<(i&31)) != 0 {
				_base.Sig.Recv[i/32] &^= 1 << (i & 31)
				return i
			}
		}

		// Wait for updates to be available from signal sender.
	Receive:
		for {
			switch _base.Atomicload(&_base.Sig.State) {
			default:
				_base.Throw("signal_recv: inconsistent state")
			case _base.SigIdle:
				if _base.Cas(&_base.Sig.State, _base.SigIdle, _base.SigReceiving) {
					_base.Notetsleepg(&_base.Sig.Note, -1)
					_base.Noteclear(&_base.Sig.Note)
					break Receive
				}
			case _base.SigSending:
				if _base.Cas(&_base.Sig.State, _base.SigSending, _base.SigIdle) {
					break Receive
				}
			}
		}

		// Incorporate updates from sender into local copy.
		for i := range _base.Sig.Mask {
			_base.Sig.Recv[i] = xchg(&_base.Sig.Mask[i], 0)
		}
	}
}

// Must only be called from a single goroutine at a time.
func signal_enable(s uint32) {
	if !_base.Sig.Inuse {
		// The first call to signal_enable is for us
		// to use for initialization.  It does not pass
		// signal information in m.
		_base.Sig.Inuse = true // enable reception of signals; cannot disable
		_base.Noteclear(&_base.Sig.Note)
		return
	}

	if int(s) >= len(_base.Sig.Wanted)*32 {
		return
	}
	_base.Sig.Wanted[s/32] |= 1 << (s & 31)
	sigenable(s)
}

// Must only be called from a single goroutine at a time.
func signal_disable(s uint32) {
	if int(s) >= len(_base.Sig.Wanted)*32 {
		return
	}
	_base.Sig.Wanted[s/32] &^= 1 << (s & 31)
	sigdisable(s)
}

// Must only be called from a single goroutine at a time.
func signal_ignore(s uint32) {
	if int(s) >= len(_base.Sig.Wanted)*32 {
		return
	}
	_base.Sig.Wanted[s/32] &^= 1 << (s & 31)
	sigignore(s)
}

// This runs on a foreign stack, without an m or a g.  No stack split.
//go:nosplit
//go:norace
func badsignal(sig uintptr) {
	cgocallback(unsafe.Pointer(_base.FuncPC(badsignalgo)), _base.Noescape(unsafe.Pointer(&sig)), unsafe.Sizeof(sig))
}

func badsignalgo(sig uintptr) {
	if !_base.Sigsend(uint32(sig)) {
		// A foreign thread received the signal sig, and the
		// Go code does not want to handle it.
		raisebadsignal(int32(sig))
	}
}
