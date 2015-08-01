// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	_base "runtime/internal/base"
	_gc "runtime/internal/gc"
	_print "runtime/internal/print"
	"unsafe"
)

// BSD interface for threading.
func osinit() {
	// bsdthread_register delayed until end of goenvs so that we
	// can look at the environment first.

	// Use sysctl to fetch hw.ncpu.
	mib := [2]uint32{6, 3}
	out := uint32(0)
	nout := unsafe.Sizeof(out)
	ret := sysctl(&mib[0], 2, (*byte)(unsafe.Pointer(&out)), &nout, nil, 0)
	if ret >= 0 {
		_base.Ncpu = int32(out)
	}
}

var urandom_dev = []byte("/dev/urandom\x00")

//go:nosplit
func getRandomData(r []byte) {
	fd := open(&urandom_dev[0], 0 /* O_RDONLY */, 0)
	n := read(fd, unsafe.Pointer(&r[0]), int32(len(r)))
	closefd(fd)
	extendRandom(r, int(n))
}

func goenvs() {
	goenvs_unix()

	// Register our thread-creation callback (see sys_darwin_{amd64,386}.s)
	// but only if we're not using cgo.  If we are using cgo we need
	// to let the C pthread library install its own thread-creation callback.
	if !_base.Iscgo {
		if bsdthread_register() != 0 {
			if _gc.Gogetenv("DYLD_INSERT_LIBRARIES") != "" {
				_base.Throw("runtime: bsdthread_register error (unset DYLD_INSERT_LIBRARIES)")
			}
			_base.Throw("runtime: bsdthread_register error")
		}
	}
}

// newosproc0 is a version of newosproc that can be called before the runtime
// is initialized.
//
// As Go uses bsdthread_register when running without cgo, this function is
// not safe to use after initialization as it does not pass an M as fnarg.
//
//go:nosplit
func newosproc0(stacksize uintptr, fn unsafe.Pointer, fnarg uintptr) {
	stack := _base.SysAlloc(stacksize, &_base.Memstats.Stacks_sys)
	if stack == nil {
		_print.Write(2, unsafe.Pointer(&failallocatestack[0]), int32(len(failallocatestack)))
		_base.Exit(1)
	}
	stk := unsafe.Pointer(uintptr(stack) + stacksize)

	var oset uint32
	_base.Sigprocmask(_base.SIG_SETMASK, &_base.Sigset_all, &oset)
	errno := _base.Bsdthread_create(stk, fn, fnarg)
	_base.Sigprocmask(_base.SIG_SETMASK, &oset, nil)

	if errno < 0 {
		_print.Write(2, unsafe.Pointer(&failthreadcreate[0]), int32(len(failthreadcreate)))
		_base.Exit(1)
	}
}

var failallocatestack = []byte("runtime: failed to allocate stack for the new OS thread\n")
var failthreadcreate = []byte("runtime: failed to create new OS thread\n")

// Called from dropm to undo the effect of an minit.
func unminit() {
	_g_ := _base.Getg()
	smask := (*uint32)(unsafe.Pointer(&_g_.M.Sigmask))
	_base.Sigprocmask(_base.SIG_SETMASK, smask, nil)
	_base.Signalstack(nil)
}

type tmach_semdestroymsg struct {
	h         _base.Machheader
	body      _base.Machbody
	semaphore _base.Machport
}

func mach_semdestroy(sem uint32) {
	var m [256]uint8
	tx := (*tmach_semdestroymsg)(unsafe.Pointer(&m))

	tx.h.Msgh_bits = _base.MACH_MSGH_BITS_COMPLEX
	tx.h.Msgh_size = uint32(unsafe.Sizeof(*tx))
	tx.h.Msgh_remote_port = _base.Mach_task_self()
	tx.h.Msgh_id = _base.Tmach_semdestroy
	tx.body.Msgh_descriptor_count = 1
	tx.semaphore.Name = sem
	tx.semaphore.Disposition = _base.MACH_MSG_TYPE_MOVE_SEND
	tx.semaphore.Type = 0

	for {
		r := _base.Machcall(&tx.h, int32(unsafe.Sizeof(m)), 0)
		if r == 0 {
			break
		}
		if r == _base.KERN_ABORTED { // interrupted
			continue
		}
		_base.Macherror(r, "semaphore_destroy")
	}
}
func mach_semaphore_signal_all(sema uint32) int32

func memlimit() uintptr {
	// NOTE(rsc): Could use getrlimit here,
	// like on FreeBSD or Linux, but Darwin doesn't enforce
	// ulimit -v, so it's unclear why we'd try to stay within
	// the limit.
	return 0
}

func unblocksig(sig int32) {
	mask := uint32(1) << (uint32(sig) - 1)
	_base.Sigprocmask(_base.SIG_UNBLOCK, &mask, nil)
}
