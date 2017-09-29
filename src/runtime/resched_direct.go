// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris
// +build amd64

// This implements direct fault-based loop preemption. In this model,
// the compiler inserts a load from reschedulePage.check at
// preemptible points in loops. To preempt all loops, the runtime
// unmaps the page containing this field and traps the resulting
// memory faults and redirects the user code to the PC of the flush
// path recorded in _PCDATA_ReschedulePC.
//
// On amd64 and 386, this check can be done in a single instruction
// that clobbers no registers.

package runtime

// reschedulePagePad is the bytes of padding around the loop
// rescheduling byte. This must be at least the physical page size.
// Since this only uses BSS space, there's not much need to keep this
// low, so we just set it to the largest page size of any system we
// support.
//
// This must be kept in sync with the compiler:
// ../cmd/compile/internal/amd64/ssa.go:reschedulePagePad
const reschedulePagePad = 64 << 10

// reschedulePage contains a page that will be unmapped to
// cause traps at safe points in loops.
var reschedulePage struct {
	before [reschedulePagePad]uint8
	check  uint8
	after  [reschedulePagePad - 1]uint8
}
