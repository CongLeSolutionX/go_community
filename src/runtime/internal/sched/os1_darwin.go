// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sched

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	"unsafe"
)

var sigset_all = ^uint32(0)

func newosproc(mp *_core.M, stk unsafe.Pointer) {
	mp.Tls[0] = uintptr(mp.Id) // so 386 asm can find it
	if false {
		print("newosproc stk=", stk, " m=", mp, " g=", mp.G0, " id=", mp.Id, "/", int(mp.Tls[0]), " ostk=", &mp, "\n")
	}

	var oset uint32
	_core.Sigprocmask(_core.SIG_SETMASK, &sigset_all, &oset)
	errno := bsdthread_create(stk, mp, mp.G0, _lock.FuncPC(Mstart))
	_core.Sigprocmask(_core.SIG_SETMASK, &oset, nil)

	if errno < 0 {
		print("runtime: failed to create new OS thread (have ", mcount(), " already; errno=", -errno, ")\n")
		_lock.Throw("runtime.newosproc")
	}
}

// Called to initialize a new m (including the bootstrap m).
// Called on the parent thread (main thread in case of bootstrap), can allocate memory.
func mpreinit(mp *_core.M) {
	mp.Gsignal = Malg(32 * 1024) // OS X wants >= 8K
	mp.Gsignal.M = mp
}

func setsigstack(i int32) {
	_lock.Throw("setsigstack")
}

func Getsig(i int32) uintptr {
	var sa _lock.Sigactiont
	_core.Memclr(unsafe.Pointer(&sa), unsafe.Sizeof(sa))
	_lock.Sigaction(uint32(i), nil, &sa)
	return *(*uintptr)(unsafe.Pointer(&sa.Sigaction_u))
}
