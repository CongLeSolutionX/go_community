// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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
