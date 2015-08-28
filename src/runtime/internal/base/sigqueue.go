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

package base

var Sig struct {
	Note   Note
	Mask   [(NSIG + 31) / 32]uint32
	Wanted [(NSIG + 31) / 32]uint32
	Recv   [(NSIG + 31) / 32]uint32
	State  uint32
	Inuse  bool
}

const (
	SigIdle = iota
	SigReceiving
	SigSending
)

// Called from sighandler to send a signal back out of the signal handling thread.
// Reports whether the signal was sent. If not, the caller typically crashes the program.
func Sigsend(s uint32) bool {
	bit := uint32(1) << uint(s&31)
	if !Sig.Inuse || s < 0 || int(s) >= 32*len(Sig.Wanted) || Sig.Wanted[s/32]&bit == 0 {
		return false
	}

	// Add signal to outgoing queue.
	for {
		mask := Sig.Mask[s/32]
		if mask&bit != 0 {
			return true // signal already in queue
		}
		if Cas(&Sig.Mask[s/32], mask, mask|bit) {
			break
		}
	}

	// Notify receiver that queue has new bit.
Send:
	for {
		switch Atomicload(&Sig.State) {
		default:
			Throw("sigsend: inconsistent state")
		case SigIdle:
			if Cas(&Sig.State, SigIdle, SigSending) {
				break Send
			}
		case SigSending:
			// notification already pending
			break Send
		case SigReceiving:
			if Cas(&Sig.State, SigReceiving, SigIdle) {
				Notewakeup(&Sig.Note)
				break Send
			}
		}
	}

	return true
}
