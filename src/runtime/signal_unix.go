// Copyright 2012 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package runtime

import (
	_base "runtime/internal/base"
	"unsafe"
)

//go:linkname os_sigpipe os.sigpipe
func os_sigpipe() {
	_base.Systemstack(sigpipe)
}

// Determines if the signal should be handled by Go and if not, forwards the
// signal to the handler that was installed before Go's.  Returns whether the
// signal was forwarded.
//go:nosplit
func sigfwdgo(sig uint32, info *_base.Siginfo, ctx unsafe.Pointer) bool {
	g := _base.Getg()
	c := &_base.Sigctxt{info, ctx}
	if sig >= uint32(len(_base.Sigtable)) {
		return false
	}
	fwdFn := _base.FwdSig[sig]
	flags := _base.Sigtable[sig].Flags

	// If there is no handler to forward to, no need to forward.
	if fwdFn == _base.SIG_DFL {
		return false
	}
	// Only forward synchronous signals.
	if c.Sigcode() == _base.SI_USER || flags&_base.SigPanic == 0 {
		return false
	}
	// Determine if the signal occurred inside Go code.  We test that:
	//   (1) we were in a goroutine (i.e., m.curg != nil), and
	//   (2) we weren't in CGO (i.e., m.curg.syscallsp == 0).
	if g != nil && g.M != nil && g.M.Curg != nil && g.M.Curg.Syscallsp == 0 {
		return false
	}
	// Signal not handled by Go, forward it.
	if fwdFn != _base.SIG_IGN {
		sigfwd(fwdFn, sig, info, ctx)
	}
	return true
}
