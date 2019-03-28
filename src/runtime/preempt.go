// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import "runtime/internal/atomic"

// preemptFlags is a bit set of tasks to perform at a goroutine preemption.
type preemptFlags uint8

const (
	// preemptScan indicates that a goroutine's stack should be
	// scanned at the next preemption point.
	preemptScan preemptFlags = 1 << iota

	// preemptSched indicates that a goroutine switch should
	// happen at the next preemption point.
	preemptSched
)

// setPreempt sets a preempt flag for gp and poisons the stack so it
// will enter the preemption path. setPreempt can be called from any
// goroutine at any time.
func (gp *g) setPreempt(p preemptFlags) {
	atomic.Or8((*uint8)(&gp.preempt), uint8(p))
	// Poison the stack guard so the next stack check fails and
	// enters newstack, which will detect that this was a
	// preemption and enter the scheduler.
	gp.stackguard0 = stackPreempt
}

// clearPreempt clears a preemption flag for gp and, if there are no
// more preemption tasks, un-poisons gp's stack. clearPreempt can be
// called concurrently with setPreempt. The caller must own gp's
// stack.
func (gp *g) clearPreempt(p preemptFlags) {
	// Speculatively clear the stack poison. We do this before
	// clearing the preempt bit so that a race with setPreempt
	// will leave the stack poisoned.
	gp.stackguard0 = gp.stack.lo + _StackGuard
	atomic.And8((*uint8)(&gp.preempt), ^uint8(p))

	// If preempt flags are still set, re-poison the stack.
	if gp.preempt != 0 {
		gp.stackguard0 = stackPreempt
	}
}

// resetStackGuard resets gp's stack guard based on gp's preemption
// flags. The caller must own gp's stack.
//
// This is nosplit because it is called from exitsyscall on gp itself.
// The stack guard is known to be poisoned, so this had better not
// check it.
//
//go:nosplit
func (gp *g) resetStackGuard() {
	// Like in clearPreempt, speculatively clear the poison.
	gp.stackguard0 = gp.stack.lo + _StackGuard

	// Re-poison the stack if flags are set.
	if atomic.Load8((*uint8)(&gp.preempt)) != 0 {
		gp.stackguard0 = stackPreempt
	}
}

// drainPreempt processes and clears preemption tasks for gp.
//
// gp must be in status _Grunnable.
//
//go:systemstack
func drainPreempt(gp *g) {
	// Synchronize with scang.
	casgstatus(gp, _Grunnable, _Gwaiting)
	if gp.preempt&preemptScan != 0 {
		for !castogscanstatus(gp, _Gwaiting, _Gscanwaiting) {
			// Likely to be racing with the GC as
			// it sees a _Gwaiting and does the
			// stack scan. If so, gcworkdone will
			// be set and gcphasework will simply
			// return.
		}
		gp.clearPreempt(preemptScan)
		if !gp.gcscandone {
			// gcw is safe because we're on the
			// system stack.
			gcw := &getg().m.p.ptr().gcw
			scanstack(gp, gcw)
			gp.gcscandone = true
		}
		casfrom_Gscanstatus(gp, _Gscanwaiting, _Gwaiting)
	}
	casgstatus(gp, _Gwaiting, _Grunnable)
}
