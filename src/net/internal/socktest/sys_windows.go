// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package socktest

import (
	"internal/syscall/windows"
	"sync/atomic"
	"syscall"
)

// Socket wraps syscall.Socket.
func (sw *Switch) Socket(family, sotype, proto int) (s syscall.Handle, err error) {
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
		return syscall.InvalidHandle, st.Err
	}
	sw.addBinding(uintptr(s), st.Cookie)
	sw.notify(uintptr(s), st.Cookie)
	sw.stats.getLocked(st.Cookie).Opened++
	return s, nil
}

// WSASocket wraps syscall.WSASocket.
func (sw *Switch) WSASocket(family, sotype, proto int32, protinfo *syscall.WSAProtocolInfo, group uint32, flags uint32) (s syscall.Handle, err error) {
	sw.once.Do(sw.init)
	if atomic.LoadUint32(&sw.state) == switchDisabled {
		return windows.WSASocket(family, sotype, proto, protinfo, group, flags)
	}

	st := &State{Cookie: cookie(int(family), int(sotype), int(proto))}
	s, st.Err = windows.WSASocket(family, sotype, proto, protinfo, group, flags)

	sw.statMu.Lock()
	defer sw.statMu.Unlock()
	if st.Err != nil {
		sw.stats.getLocked(st.Cookie).OpenFailed++
		return syscall.InvalidHandle, st.Err
	}
	sw.addBinding(uintptr(s), st.Cookie)
	sw.notify(uintptr(s), st.Cookie)
	sw.stats.getLocked(st.Cookie).Opened++
	return s, nil
}

// Closesocket wraps syscall.Closesocket.
func (sw *Switch) Closesocket(s syscall.Handle) (err error) {
	if atomic.LoadUint32(&sw.state) == switchDisabled {
		return syscall.Closesocket(s)
	}

	b := sw.priorBinding(uintptr(s))
	if b == nil {
		return syscall.Closesocket(s)
	}
	st := b.newState()
	f := b.filter(FilterClose)

	af, err := f.apply(st)
	if err != nil {
		return err
	}
	st.Err = syscall.Closesocket(s)
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
func (sw *Switch) Connect(s syscall.Handle, sa syscall.Sockaddr) (err error) {
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

// ConnectEx wraps syscall.ConnectEx.
func (sw *Switch) ConnectEx(s syscall.Handle, sa syscall.Sockaddr, bb *byte, n uint32, nwr *uint32, o *syscall.Overlapped) (err error) {
	if atomic.LoadUint32(&sw.state) == switchDisabled {
		return syscall.ConnectEx(s, sa, bb, n, nwr, o)
	}

	b := sw.posteriorBinding(uintptr(s))
	if b == nil {
		return syscall.ConnectEx(s, sa, bb, n, nwr, o)
	}
	st := b.newState()
	f := b.filter(FilterConnect)

	af, err := f.apply(st)
	if err != nil {
		return err
	}
	st.Err = syscall.ConnectEx(s, sa, bb, n, nwr, o)
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
func (sw *Switch) Listen(s syscall.Handle, backlog int) (err error) {
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

// AcceptEx wraps syscall.AcceptEx.
func (sw *Switch) AcceptEx(ls syscall.Handle, as syscall.Handle, bb *byte, rxdatalen uint32, laddrlen uint32, raddrlen uint32, rcvd *uint32, overlapped *syscall.Overlapped) error {
	if atomic.LoadUint32(&sw.state) == switchDisabled {
		return syscall.AcceptEx(ls, as, bb, rxdatalen, laddrlen, raddrlen, rcvd, overlapped)
	}

	b := sw.posteriorBinding(uintptr(ls))
	if b == nil {
		return syscall.AcceptEx(ls, as, bb, rxdatalen, laddrlen, raddrlen, rcvd, overlapped)
	}
	st := b.newState()
	f := b.filter(FilterAccept)

	af, err := f.apply(st)
	if err != nil {
		return err
	}
	st.Err = syscall.AcceptEx(ls, as, bb, rxdatalen, laddrlen, raddrlen, rcvd, overlapped)
	if err = af.apply(st); err != nil {
		return err
	}

	sw.statMu.Lock()
	defer sw.statMu.Unlock()
	if st.Err != nil {
		sw.stats.getLocked(st.Cookie).AcceptFailed++
		return st.Err
	}
	sw.addBinding(uintptr(as), st.Cookie)
	sw.notify(uintptr(as), st.Cookie)
	sw.stats.getLocked(st.Cookie).Accepted++
	return nil
}
