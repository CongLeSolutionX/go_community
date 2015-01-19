// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd netbsd openbsd

package sched

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	"unsafe"
)

//go:noescape
func Kevent(kq int32, ch *Keventt, nch int32, ev *Keventt, nev int32, ts *timespec) int32

var (
	Kq             int32 = -1
	netpolllasterr int32
)

// Polls for ready network connections.
// Returns list of goroutines that become runnable.
func Netpoll(block bool) (gp *_core.G) {
	if Kq == -1 {
		return
	}
	var tp *timespec
	var ts timespec
	if !block {
		tp = &ts
	}
	var events [64]Keventt
retry:
	n := Kevent(Kq, nil, 0, &events[0], int32(len(events)), tp)
	if n < 0 {
		if n != -_lock.EINTR && n != netpolllasterr {
			netpolllasterr = n
			println("runtime: kevent on fd", Kq, "failed with", -n)
		}
		goto retry
	}
	for i := 0; i < int(n); i++ {
		ev := &events[i]
		var mode int32
		if ev.Filter == _lock.EVFILT_READ {
			mode += 'r'
		}
		if ev.Filter == _lock.EVFILT_WRITE {
			mode += 'w'
		}
		if mode != 0 {
			netpollready((**_core.G)(_core.Noescape(unsafe.Pointer(&gp))), (*PollDesc)(unsafe.Pointer(ev.Udata)), mode)
		}
	}
	if block && gp == nil {
		goto retry
	}
	return gp
}
