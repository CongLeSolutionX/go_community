// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Garbage collector: marking and scanning

package runtime

import (
	_base "runtime/internal/base"
)

// nextBarrierPC returns the original return PC of the next stack barrier.
// Used by getcallerpc, so it must be nosplit.
//go:nosplit
func nextBarrierPC() uintptr {
	gp := _base.Getg()
	return gp.Stkbar[gp.StkbarPos].SavedLRVal
}

// setNextBarrierPC sets the return PC of the next stack barrier.
// Used by setcallerpc, so it must be nosplit.
//go:nosplit
func setNextBarrierPC(pc uintptr) {
	gp := _base.Getg()
	gp.Stkbar[gp.StkbarPos].SavedLRVal = pc
}
