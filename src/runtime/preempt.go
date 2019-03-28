// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Goroutine preemption
//
// A goroutine can be preempted at any safe-point. Currently, there
// are a few categories of safe-points:
//
// 1. A blocked safe-point occurs for the duration that a goroutine is
//    descheduled, blocked on synchronization, or in a system call.
//
// 2. Synchronous safe-points occur when a running goroutine checks
//    for a preemption request.
//
// At both blocked and synchronous safe-points, a goroutine's CPU
// state is minimal and the garbage collector has complete information
// about its entire stack. This makes it possible to deschedule a
// goroutine with minimal space, and to precisely scan a goroutine's
// stack.
//
// Synchronous safe-points are implemented by overloading the stack
// bound check in function prologues. To preempt a goroutine at the
// next synchronous safe-point, the runtime poisons the goroutine's
// stack bound to a value that will cause the next stack bound check
// to fail and enter the stack growth implementation, which will
// detect that it was actually a preemption and redirect to preemption
// handling.
//
// A preemption task is injected into a goroutine by setting the
// appropriate bit of the g.preempt field. These tasks may be drained
// by the goroutine itself or by another goroutine on its behalf
// (e.g., if the goroutine is at a blocked safe-point). In all cases,
// draining is protected by locking the goroutine's _Gscan bit.
//
// Because transitions between kinds of safe-points can happen
// concurrently and asynchronously, multiple mechanisms compete to
// perform preemption tasks promptly. The injecting goroutine will
// drain the tasks if the target goroutine is at a blocked safe-point,
// or else poison the stack if the target is running so it drains its
// own tasks at the next synchronous safe-point. But since a goroutine
// may transition from running to blocked without a synchronous
// safe-point, a goroutine will also drain its own preemption tasks
// after any transition to a blocked safe-point. This ensures that at
// least one of the injecting goroutine or the target goroutine will
// attempt to drain preemption tasks.

package runtime

import "runtime/internal/atomic"

// preemptFlags is a bit set of tasks to perform at a goroutine preemption.
type preemptFlags uint8

const (
	// preemptScan indicates that a goroutine's stack should be
	// scanned at the next preemption point.
	preemptScan preemptFlags = 1 << iota

	// preemptShrink indicates that a goroutine's stack should be
	// shrunk at the next preemption point.
	preemptShrink

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
// gp must be in status _Grunnable, _Gwaiting, or _Gsyscall.
func drainPreempt(gp *g) {
	if gp.preempt&(preemptScan|preemptShrink) == 0 {
		return
	}

	// Make fast-path inlinable.
	drainPreempt1(gp)
}

func drainPreempt1(gp *g) {
	// Take ownership of the G's stack and serialize preemption
	// request handling by acquiring the _Gscan lock (which is
	// effectively a spin-lock).
	s := readgstatus(gp) &^ _Gscan // Mask out the lock bit
	switch s {
	case _Grunnable, _Gwaiting, _Gsyscall:
		// Spin to set _Gscan.
		for !castogscanstatus(gp, s, s|_Gscan) {
		}

	default:
		dumpgstatus(gp)
		throw("invalid G status")
	}

	drainPreemptLocked(gp)

	casfrom_Gscanstatus(gp, s|_Gscan, s)
}

//go:systemstack
func drainPreemptLocked(gp *g) {
	if gp.preempt&preemptScan != 0 {
		gp.clearPreempt(preemptScan)
		if !gp.gcscandone {
			// gcw is safe because we're on the
			// system stack.
			gcw := &getg().m.p.ptr().gcw
			scanstack(gp, gcw)
			gp.gcscandone = true
		}
	}
	if gp.preempt&preemptShrink != 0 {
		gp.clearPreempt(preemptShrink)
		shrinkstack(gp)
	}
}
