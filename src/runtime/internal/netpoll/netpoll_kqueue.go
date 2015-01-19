// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd netbsd openbsd

package netpoll

import (
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
	"unsafe"
)

func kqueue() int32
func closeonexec(fd int32)

func Netpollinit() {
	_sched.Kq = kqueue()
	if _sched.Kq < 0 {
		println("netpollinit: kqueue failed with", -_sched.Kq)
		_lock.Gothrow("netpollinit: kqueue failed")
	}
	closeonexec(_sched.Kq)
}

func Netpollopen(fd uintptr, pd *_sched.PollDesc) int32 {
	// Arm both EVFILT_READ and EVFILT_WRITE in edge-triggered mode (EV_CLEAR)
	// for the whole fd lifetime.  The notifications are automatically unregistered
	// when fd is closed.
	var ev [2]_sched.Keventt
	*(*uintptr)(unsafe.Pointer(&ev[0].Ident)) = fd
	ev[0].Filter = _lock.EVFILT_READ
	ev[0].Flags = _lock.EV_ADD | _lock.EV_CLEAR
	ev[0].Fflags = 0
	ev[0].Data = 0
	ev[0].Udata = (*byte)(unsafe.Pointer(pd))
	ev[1] = ev[0]
	ev[1].Filter = _lock.EVFILT_WRITE
	n := _sched.Kevent(_sched.Kq, &ev[0], 2, nil, 0, nil)
	if n < 0 {
		return -n
	}
	return 0
}

func Netpollclose(fd uintptr) int32 {
	// Don't need to unregister because calling close()
	// on fd will remove any kevents that reference the descriptor.
	return 0
}

func Netpollarm(pd *_sched.PollDesc, mode int) {
	_lock.Gothrow("unused")
}
