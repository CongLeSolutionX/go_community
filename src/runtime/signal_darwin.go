// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	_base "runtime/internal/base"
	"unsafe"
)

//go:noescape
func sigfwd(fn uintptr, sig uint32, info *_base.Siginfo, ctx unsafe.Pointer)

//go:noescape
func sigreturn(ctx unsafe.Pointer, infostyle uint32)

//go:nosplit
func sigtrampgo(fn uintptr, infostyle, sig uint32, info *_base.Siginfo, ctx unsafe.Pointer) {
	if sigfwdgo(sig, info, ctx) {
		sigreturn(ctx, infostyle)
		return
	}
	g := _base.Getg()
	if g == nil {
		badsignal(uintptr(sig))
		sigreturn(ctx, infostyle)
		return
	}
	setg(g.M.Gsignal)
	_base.Sighandler(sig, info, ctx, g)
	setg(g)
	sigreturn(ctx, infostyle)
}
