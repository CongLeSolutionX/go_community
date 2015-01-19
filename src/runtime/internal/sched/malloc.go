// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sched

import (
	_core "runtime/internal/core"
)

const (
	DebugMalloc = false

	XFlagNoScan = FlagNoScan
	XFlagNoZero = FlagNoZero

	MaxTinySize   = _core.TinySize
	TinySizeClass = _core.TinySizeClass
	MaxSmallSize  = _core.MaxSmallSize

	PageShift = _core.PageShift
	PageSize  = _core.PageSize
	PageMask  = _core.PageMask

	XBitsPerPointer  = BitsPerPointer
	XBitsMask        = BitsMask
	XPointersPerByte = PointersPerByte
	XMaxGCMask       = MaxGCMask
	XBitsDead        = BitsDead
	XBitsPointer     = BitsPointer
	XBitsScalar      = BitsScalar

	XMSpanInUse = MSpanInUse

	XConcurrentSweep = ConcurrentSweep
)

// round n up to a multiple of a.  a must be a power of 2.
func Round(n, a uintptr) uintptr {
	return (n + a - 1) &^ (a - 1)
}
