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

func (f *preemptFlags) set(p preemptFlags) {
	atomic.Or8((*uint8)(f), uint8(p))
}

func (f *preemptFlags) clear(p preemptFlags) {
	atomic.And8((*uint8)(f), ^uint8(p))
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
		gp.preempt.clear(preemptScan)
		for !castogscanstatus(gp, _Gwaiting, _Gscanwaiting) {
			// Likely to be racing with the GC as
			// it sees a _Gwaiting and does the
			// stack scan. If so, gcworkdone will
			// be set and gcphasework will simply
			// return.
		}
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
