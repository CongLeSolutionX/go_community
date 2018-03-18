// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

package socktest

import (
	"sync/atomic"
	"syscall"
)

// Socket wraps syscall.Socket.
func (sw *Switch) Socket(family, sotype, proto int) (s int, err error) {
	sw.once.Do(sw.init)
	if atomic.LoadUint32(&sw.state) == switchDisabled {
		return syscall.Socket(family, sotype, proto)
	}

	st := &State{Cookie: cookie(family, sotype, proto)}
	s, st.Err = syscall.Socket(family, sotype, proto)

	sw.statMu.Lock()
	defer sw.statMu.Unlock()
	if st.Err != nil {
		sw.stats.getLocked(st.Cookie).OpenFailed++
		return -1, st.Err
	}
	sw.addBinding(uintptr(s), st.Cookie)
	sw.notify(uintptr(s), st.Cookie)
	sw.stats.getLocked(st.Cookie).Opened++
	return s, nil
}

// Close wraps syscall.Close.
func (sw *Switch) Close(s int) (err error) {
	if atomic.LoadUint32(&sw.state) == switchDisabled {
		return syscall.Close(s)
	}

	b := sw.priorBinding(uintptr(s))
	if b == nil {
		return syscall.Close(s)
	}
	st := b.newState()
	f := b.filter(FilterClose)

	af, err := f.apply(st)
	if err != nil {
		return err
	}
	st.Err = syscall.Close(s)
	if err = af.apply(st); err != nil {
		return err
	}

	sw.statMu.Lock()
	defer sw.statMu.Unlock()
	if st.Err != nil {
		sw.stats.getLocked(st.Cookie).CloseFailed++
		return st.Err
	}
	sw.delBinding(uintptr(s))
	sw.stats.getLocked(st.Cookie).Closed++
	return nil
}

// Connect wraps syscall.Connect.
func (sw *Switch) Connect(s int, sa syscall.Sockaddr) (err error) {
	if atomic.LoadUint32(&sw.state) == switchDisabled {
		return syscall.Connect(s, sa)
	}

	b := sw.posteriorBinding(uintptr(s))
	if b == nil {
		return syscall.Connect(s, sa)
	}
	st := b.newState()
	f := b.filter(FilterConnect)

	af, err := f.apply(st)
	if err != nil {
		return err
	}
	st.Err = syscall.Connect(s, sa)
	if err = af.apply(st); err != nil {
		return err
	}

	sw.statMu.Lock()
	defer sw.statMu.Unlock()
	if st.Err != nil {
		sw.stats.getLocked(st.Cookie).ConnectFailed++
		return st.Err
	}
	sw.stats.getLocked(st.Cookie).Connected++
	return nil
}

// Listen wraps syscall.Listen.
func (sw *Switch) Listen(s, backlog int) (err error) {
	if atomic.LoadUint32(&sw.state) == switchDisabled {
		return syscall.Listen(s, backlog)
	}

	b := sw.posteriorBinding(uintptr(s))
	if b == nil {
		return syscall.Listen(s, backlog)
	}
	st := b.newState()
	f := b.filter(FilterListen)

	af, err := f.apply(st)
	if err != nil {
		return err
	}
	st.Err = syscall.Listen(s, backlog)
	if err = af.apply(st); err != nil {
		return err
	}

	sw.statMu.Lock()
	defer sw.statMu.Unlock()
	if st.Err != nil {
		sw.stats.getLocked(st.Cookie).ListenFailed++
		return st.Err
	}
	sw.stats.getLocked(st.Cookie).Listened++
	return nil
}

// Accept wraps syscall.Accept.
func (sw *Switch) Accept(s int) (ns int, sa syscall.Sockaddr, err error) {
	if atomic.LoadUint32(&sw.state) == switchDisabled {
		return syscall.Accept(s)
	}

	b := sw.posteriorBinding(uintptr(s))
	if b == nil {
		return syscall.Accept(s)
	}
	st := b.newState()
	f := b.filter(FilterAccept)

	af, err := f.apply(st)
	if err != nil {
		return -1, nil, err
	}
	ns, sa, st.Err = syscall.Accept(s)
	if err = af.apply(st); err != nil {
		if st.Err == nil {
			syscall.Close(ns)
		}
		return -1, nil, err
	}

	sw.statMu.Lock()
	defer sw.statMu.Unlock()
	if st.Err != nil {
		sw.stats.getLocked(st.Cookie).AcceptFailed++
		return -1, nil, st.Err
	}
	sw.addBinding(uintptr(ns), st.Cookie)
	sw.notify(uintptr(ns), st.Cookie)
	sw.stats.getLocked(st.Cookie).Accepted++
	return ns, sa, nil
}

// GetsockoptInt wraps syscall.GetsockoptInt.
func (sw *Switch) GetsockoptInt(s, lvl, opt int) (soerr int, err error) {
	if atomic.LoadUint32(&sw.state) == switchDisabled {
		return syscall.GetsockoptInt(s, lvl, opt)
	}

	b := sw.posteriorBinding(uintptr(s))
	if b == nil {
		return syscall.GetsockoptInt(s, lvl, opt)
	}
	st := b.newState()
	f := b.filter(FilterGetsockoptInt)

	af, err := f.apply(st)
	if err != nil {
		return -1, err
	}
	soerr, st.Err = syscall.GetsockoptInt(s, lvl, opt)
	st.SocketErr = syscall.Errno(soerr)
	if err = af.apply(st); err != nil {
		return -1, err
	}

	sw.statMu.Lock()
	defer sw.statMu.Unlock()
	if st.Err != nil {
		sw.stats.getLocked(st.Cookie).StatusFetchFailed++
		return -1, st.Err
	}
	if opt == syscall.SO_ERROR && (st.SocketErr == syscall.Errno(0) || st.SocketErr == syscall.EISCONN) {
		sw.stats.getLocked(st.Cookie).Connected++
	}
	sw.stats.getLocked(st.Cookie).StatusFetched++
	return soerr, nil
}
