// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sched

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	"unsafe"
)

// tracebacktrap is like traceback but expects that the PC and SP were obtained
// from a trap, not from gp->sched or gp->syscallpc/gp->syscallsp or getcallerpc/getcallersp.
// Because they are from a trap instead of from a saved pair,
// the initial PC must not be rewound to the previous instruction.
// (All the saved pairs record a PC that is a return address, so we
// rewind it into the CALL instruction.)
func tracebacktrap(pc uintptr, sp uintptr, lr uintptr, gp *_core.G) {
	_lock.Traceback1(pc, sp, lr, gp, _lock.TraceTrap)
}

func Callers(skip int, pcbuf *uintptr, m int) int {
	sp := _lock.Getcallersp(unsafe.Pointer(&skip))
	pc := uintptr(_lock.Getcallerpc(unsafe.Pointer(&skip)))
	var n int
	_lock.Systemstack(func() {
		n = _lock.Gentraceback(pc, sp, 0, _core.Getg(), skip, pcbuf, m, nil, nil, 0)
	})
	return n
}

func Gcallers(gp *_core.G, skip int, pcbuf *uintptr, m int) int {
	return _lock.Gentraceback(^uintptr(0), ^uintptr(0), 0, gp, skip, pcbuf, m, nil, nil, 0)
}
