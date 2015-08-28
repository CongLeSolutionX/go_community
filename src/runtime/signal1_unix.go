// Copyright 2012 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package runtime

import (
	_base "runtime/internal/base"
)

// channels for synchronizing signal mask updates with the signal mask
// thread
var (
	disableSigChan  chan uint32
	enableSigChan   chan uint32
	maskUpdatedChan chan struct{}
)

func sigenable(sig uint32) {
	if sig >= uint32(len(_base.Sigtable)) {
		return
	}

	t := &_base.Sigtable[sig]
	if t.Flags&_base.SigNotify != 0 {
		ensureSigM()
		enableSigChan <- sig
		<-maskUpdatedChan
		if t.Flags&_base.SigHandling == 0 {
			t.Flags |= _base.SigHandling
			if _base.Getsig(int32(sig)) == _base.SIG_IGN {
				t.Flags |= _base.SigIgnored
			}
			_base.Setsig(int32(sig), _base.FuncPC(_base.Sighandler), true)
		}
	}
}

func sigdisable(sig uint32) {
	if sig >= uint32(len(_base.Sigtable)) {
		return
	}

	t := &_base.Sigtable[sig]
	if t.Flags&_base.SigNotify != 0 {
		ensureSigM()
		disableSigChan <- sig
		<-maskUpdatedChan
		if t.Flags&_base.SigHandling != 0 {
			t.Flags &^= _base.SigHandling
			if t.Flags&_base.SigIgnored != 0 {
				_base.Setsig(int32(sig), _base.SIG_IGN, true)
			} else {
				_base.Setsig(int32(sig), _base.SIG_DFL, true)
			}
		}
	}
}

func sigignore(sig uint32) {
	if sig >= uint32(len(_base.Sigtable)) {
		return
	}

	t := &_base.Sigtable[sig]
	if t.Flags&_base.SigNotify != 0 {
		t.Flags &^= _base.SigHandling
		_base.Setsig(int32(sig), _base.SIG_IGN, true)
	}
}

func sigpipe() {
	_base.Setsig(_base.SIGPIPE, _base.SIG_DFL, false)
	_base.Raise(_base.SIGPIPE)
}

// raisebadsignal is called when a signal is received on a non-Go
// thread, and the Go program does not want to handle it (that is, the
// program has not called os/signal.Notify for the signal).
func raisebadsignal(sig int32) {
	if sig == _base.SIGPROF {
		// Ignore profiling signals that arrive on non-Go threads.
		return
	}

	var handler uintptr
	if sig >= _base.NSIG {
		handler = _base.SIG_DFL
	} else {
		handler = _base.FwdSig[sig]
	}

	// Reset the signal handler and raise the signal.
	// We are currently running inside a signal handler, so the
	// signal is blocked.  We need to unblock it before raising the
	// signal, or the signal we raise will be ignored until we return
	// from the signal handler.  We know that the signal was unblocked
	// before entering the handler, or else we would not have received
	// it.  That means that we don't have to worry about blocking it
	// again.
	unblocksig(sig)
	_base.Setsig(sig, handler, false)
	_base.Raise(sig)

	// If the signal didn't cause the program to exit, restore the
	// Go signal handler and carry on.
	//
	// We may receive another instance of the signal before we
	// restore the Go handler, but that is not so bad: we know
	// that the Go program has been ignoring the signal.
	_base.Setsig(sig, _base.FuncPC(_base.Sighandler), true)
}

// createSigM starts one global, sleeping thread to make sure at least one thread
// is available to catch signals enabled for os/signal.
func ensureSigM() {
	if maskUpdatedChan != nil {
		return
	}
	maskUpdatedChan = make(chan struct{})
	disableSigChan = make(chan uint32)
	enableSigChan = make(chan uint32)
	go func() {
		// Signal masks are per-thread, so make sure this goroutine stays on one
		// thread.
		LockOSThread()
		defer UnlockOSThread()
		// The sigBlocked mask contains the signals not active for os/signal,
		// initially all signals except the essential. When signal.Notify()/Stop is called,
		// sigenable/sigdisable in turn notify this thread to update its signal
		// mask accordingly.
		var sigBlocked _base.Sigmask
		for i := range sigBlocked {
			sigBlocked[i] = ^uint32(0)
		}
		for i := range _base.Sigtable {
			if _base.Sigtable[i].Flags&_base.SigUnblock != 0 {
				sigBlocked[(i-1)/32] &^= 1 << ((uint32(i) - 1) & 31)
			}
		}
		_base.Updatesigmask(sigBlocked)
		for {
			select {
			case sig := <-enableSigChan:
				if b := sig - 1; b >= 0 {
					sigBlocked[b/32] &^= (1 << (b & 31))
				}
			case sig := <-disableSigChan:
				if b := sig - 1; b >= 0 {
					sigBlocked[b/32] |= (1 << (b & 31))
				}
			}
			_base.Updatesigmask(sigBlocked)
			maskUpdatedChan <- struct{}{}
		}
	}()
}
