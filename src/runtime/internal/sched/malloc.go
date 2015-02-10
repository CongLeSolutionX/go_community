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

	XMSpanInUse = MSpanInUse

	XConcurrentSweep = ConcurrentSweep
)
