// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package runtime

import (
	_base "runtime/internal/base"
)

// setsigsegv is used on darwin/arm{,64} to fake a segmentation fault.
//go:nosplit
func setsigsegv(pc uintptr) {
	g := _base.Getg()
	g.Sig = _base.SIGSEGV
	g.Sigpc = pc
	g.Sigcode0 = _base.SEGV_MAPERR
	g.Sigcode1 = 0 // TODO: emulate si_addr
}
