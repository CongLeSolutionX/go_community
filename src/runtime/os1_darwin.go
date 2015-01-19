// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
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
		_lock.Ncpu = int32(out)
	}
}

// Called from dropm to undo the effect of an minit.
func unminit() {
	_core.Signalstack(nil, 0)
}

type tmach_semdestroymsg struct {
	h         _lock.Machheader
	body      _lock.Machbody
	semaphore _lock.Machport
}

func mach_semdestroy(sem uint32) {
	var m [256]uint8
	tx := (*tmach_semdestroymsg)(unsafe.Pointer(&m))

	tx.h.Msgh_bits = _lock.MACH_MSGH_BITS_COMPLEX
	tx.h.Msgh_size = uint32(unsafe.Sizeof(*tx))
	tx.h.Msgh_remote_port = _lock.Mach_task_self()
	tx.h.Msgh_id = _lock.Tmach_semdestroy
	tx.body.Msgh_descriptor_count = 1
	tx.semaphore.Name = sem
	tx.semaphore.Disposition = _lock.MACH_MSG_TYPE_MOVE_SEND
	tx.semaphore.Type = 0

	for {
		r := _lock.Machcall(&tx.h, int32(unsafe.Sizeof(m)), 0)
		if r == 0 {
			break
		}
		if r == _lock.KERN_ABORTED { // interrupted
			continue
		}
		_lock.Macherror(r, "semaphore_destroy")
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
