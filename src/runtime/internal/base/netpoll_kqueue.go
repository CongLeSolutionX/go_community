// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd netbsd openbsd

package base

import (
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
func Netpoll(block bool) *G {
	if Kq == -1 {
		return nil
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
		if n != -EINTR && n != netpolllasterr {
			netpolllasterr = n
			println("runtime: kevent on fd", Kq, "failed with", -n)
		}
		goto retry
	}
	var gp Guintptr
	for i := 0; i < int(n); i++ {
		ev := &events[i]
		var mode int32
		if ev.Filter == EVFILT_READ {
			mode += 'r'
		}
		if ev.Filter == EVFILT_WRITE {
			mode += 'w'
		}
		if mode != 0 {
			netpollready(&gp, (*PollDesc)(unsafe.Pointer(ev.Udata)), mode)
		}
	}
	if block && gp == 0 {
		goto retry
	}
	return gp.Ptr()
}
