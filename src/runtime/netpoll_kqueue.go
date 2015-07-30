// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd netbsd openbsd

package runtime

import (
	_base "runtime/internal/base"
	"unsafe"
)

func kqueue() int32
func closeonexec(fd int32)

func netpollinit() {
	_base.Kq = kqueue()
	if _base.Kq < 0 {
		println("netpollinit: kqueue failed with", -_base.Kq)
		_base.Throw("netpollinit: kqueue failed")
	}
	closeonexec(_base.Kq)
}

func netpollopen(fd uintptr, pd *_base.PollDesc) int32 {
	// Arm both EVFILT_READ and EVFILT_WRITE in edge-triggered mode (EV_CLEAR)
	// for the whole fd lifetime.  The notifications are automatically unregistered
	// when fd is closed.
	var ev [2]_base.Keventt
	*(*uintptr)(unsafe.Pointer(&ev[0].Ident)) = fd
	ev[0].Filter = _base.EVFILT_READ
	ev[0].Flags = _base.EV_ADD | _base.EV_CLEAR
	ev[0].Fflags = 0
	ev[0].Data = 0
	ev[0].Udata = (*byte)(unsafe.Pointer(pd))
	ev[1] = ev[0]
	ev[1].Filter = _base.EVFILT_WRITE
	n := _base.Kevent(_base.Kq, &ev[0], 2, nil, 0, nil)
	if n < 0 {
		return -n
	}
	return 0
}

func netpollclose(fd uintptr) int32 {
	// Don't need to unregister because calling close()
	// on fd will remove any kevents that reference the descriptor.
	return 0
}

func netpollarm(pd *_base.PollDesc, mode int) {
	_base.Throw("unused")
}
