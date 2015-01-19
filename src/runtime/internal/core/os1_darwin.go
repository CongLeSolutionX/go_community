// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	"unsafe"
)

//extern SigTabTT runtime·sigtab[];

var Sigset_none = uint32(0)

// Called to initialize a new m (including the bootstrap m).
// Called on the new thread, can not allocate memory.
func Minit() {
	// Initialize signal handling.
	_g_ := Getg()
	Signalstack((*byte)(unsafe.Pointer(_g_.M.Gsignal.Stack.Lo)), 32*1024)
	Sigprocmask(SIG_SETMASK, &Sigset_none, nil)
}

//go:nosplit
func Osyield() {
	Usleep(1)
}

func Signalstack(p *byte, n int32) {
	var st Stackt
	st.ss_sp = p
	st.ss_size = uintptr(n)
	st.ss_flags = 0
	if p == nil {
		st.ss_flags = SS_DISABLE
	}
	sigaltstack(&st, nil)
}
