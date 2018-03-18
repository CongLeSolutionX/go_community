// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package socktest provides utilities for socket testing.
package socktest

import (
	"fmt"
	"sync"
	"sync/atomic"
)

const switchDisabled = iota + 1

// A Switch represents a callpath point switch for socket system
// calls.
type Switch struct {
	once  sync.Once
	state uint32 // requires atomic op.

	statMu sync.RWMutex
	stats  stats

	// bindings can hold stale socket descriptors, which may have
	// the same descriptor number of newly allocated socket
	// descriptors, during descriptor recycling.
	bindMu   sync.RWMutex
	bindings []*binding

	noteMu  sync.RWMutex
	noteMap map[string]func(uintptr, Cookie)
}

func (sw *Switch) init() {
	sw.stats = make(stats)
	sw.noteMap = make(map[string]func(uintptr, Cookie))
}

// Stats returns a list of per-cookie socket statistics.
func (sw *Switch) Stats() []Stat {
	sw.once.Do(sw.init)

	var sts []Stat
	sw.statMu.RLock()
	defer sw.statMu.RUnlock()
	for _, st := range sw.stats {
		nst := *st
		sts = append(sts, nst)
	}
	return sts
}

// Sockets returns a list of inflight socket states.
func (sw *Switch) Sockets() []State {
	sw.once.Do(sw.init)

	sw.bindMu.RLock()
	defer sw.bindMu.RUnlock()
	sts := make([]State, 0, len(sw.bindings))
	for _, b := range sw.bindings {
		sts = append(sts, *b.newState())
	}
	return sts
}

// Cookie returns the cookie associated with s.
func (sw *Switch) Cookie(s uintptr) Cookie {
	sw.once.Do(sw.init)

	sw.bindMu.RLock()
	defer sw.bindMu.RUnlock()
	for _, b := range sw.bindings {
		if b.sysfd == s {
			return b.cookie
		}
	}
	return Cookie(1<<64 - 1)
}

// A Cookie represents a 3-tuple of a socket; address family, socket
// type and protocol number.
type Cookie uint64

// Family returns an address family.
func (c Cookie) Family() int { return int(c >> 48) }

// Type returns a socket type.
func (c Cookie) Type() int { return int(c << 16 >> 32) }

// Protocol returns a protocol number.
func (c Cookie) Protocol() int { return int(c & 0xff) }

func cookie(family, sotype, proto int) Cookie {
	return Cookie(family)<<48 | Cookie(sotype)&0xffffffff<<16 | Cookie(proto)&0xff
}

func (st State) String() string {
	return fmt.Sprintf("sysfd=%#x family=%s type=%s proto=%s syscallerr=%v socketerr=%v", st.Sysfd, familyString(st.Cookie.Family()), typeString(st.Cookie.Type()), protocolString(st.Cookie.Protocol()), st.Err, st.SocketErr)
}

// A Stat represents a per-cookie socket statistics.
type Stat struct {
	Family   int // address family
	Type     int // socket type
	Protocol int // protocol number

	Opened        uint64 // number of sockets opened
	Connected     uint64 // number of sockets connected
	Listened      uint64 // number of sockets listened
	Accepted      uint64 // number of sockets accepted
	Closed        uint64 // number of sockets closed
	StatusFetched uint64 // number of status fetched

	OpenFailed        uint64 // number of sockets open failed
	ConnectFailed     uint64 // number of sockets connect failed
	ListenFailed      uint64 // number of sockets listen failed
	AcceptFailed      uint64 // number of sockets accept failed
	CloseFailed       uint64 // number of sockets close failed
	StatusFetchFailed uint64 // number of status fetch failed
}

func (st Stat) String() string {
	return fmt.Sprintf("(%s %s %s) opened=%d connected=%d listened=%d accepted=%d closed=%d statusfetched=%d openfailed=%d connectfailed=%d listenfailed=%d acceptfailed=%d closefailed=%d statusfetchfailed=%d", familyString(st.Family), typeString(st.Type), protocolString(st.Protocol), st.Opened, st.Connected, st.Listened, st.Accepted, st.Closed, st.StatusFetched, st.OpenFailed, st.ConnectFailed, st.ListenFailed, st.AcceptFailed, st.CloseFailed, st.StatusFetchFailed)
}

type stats map[Cookie]*Stat

func (sts stats) getLocked(c Cookie) *Stat {
	st, ok := sts[c]
	if !ok {
		st = &Stat{Family: c.Family(), Type: c.Type(), Protocol: c.Protocol()}
		sts[c] = st
	}
	return st
}

// A binding represents a binding between the system calls and net
// package API calls.
type binding struct {
	sysfd  uintptr // socket descriptor
	cookie Cookie

	fltMu   sync.RWMutex
	filters [filterMax]Filter
}

func (b *binding) filter(t FilterType) Filter {
	b.fltMu.RLock()
	defer b.fltMu.RUnlock()
	return b.filters[t]
}

func (b *binding) addFilter(t FilterType, f Filter) {
	b.fltMu.Lock()
	b.filters[t] = f
	b.fltMu.Unlock()
}

func (b *binding) delFilter(t FilterType) {
	b.fltMu.Lock()
	b.filters[t] = nil
	b.fltMu.Unlock()
}

// A FilterType represents a filter type.
type FilterType int

const (
	FilterConnect       FilterType = iota // for Connect or ConnectEx
	FilterListen                          // for Listen
	FilterAccept                          // for Accept, Accept4 or AcceptEx
	FilterGetsockoptInt                   // for GetsockoptInt
	FilterClose                           // for Close or Closesocket
	filterMax
)

// A Filter represents a socket system call filter.
//
// It will only be executed before a system call for a socket that has
// an entry in internal table.
// If the filter returns a non-nil error, the execution of system call
// will be canceled and the system call function returns the non-nil
// error.
// It can return a non-nil AfterFilter for filtering after the
// execution of the system call.
type Filter func(*State) (AfterFilter, error)

func (f Filter) apply(st *State) (AfterFilter, error) {
	if f == nil {
		return nil, nil
	}
	return f(st)
}

// An AfterFilter represents a socket system call filter after an
// execution of a system call.
//
// It will only be executed after a system call for a socket that has
// an entry in internal table.
// If the filter returns a non-nil error, the system call function
// returns the non-nil error.
type AfterFilter func(*State) error

func (f AfterFilter) apply(st *State) error {
	if f == nil {
		return nil
	}
	return f(st)
}

func (sw *Switch) priorBinding(s uintptr) *binding {
	sw.bindMu.RLock()
	defer sw.bindMu.RUnlock()
	for i := range sw.bindings {
		if sw.bindings[i].sysfd == s {
			return sw.bindings[i]
		}
	}
	return nil
}

func (sw *Switch) posteriorBinding(s uintptr) *binding {
	sw.bindMu.RLock()
	defer sw.bindMu.RUnlock()
	for i := len(sw.bindings) - 1; i >= 0; i-- {
		if sw.bindings[i].sysfd == s {
			return sw.bindings[i]
		}
	}
	return nil
}

func (sw *Switch) addBinding(s uintptr, cookie Cookie) {
	sw.bindMu.Lock()
	b := &binding{sysfd: uintptr(s), cookie: cookie}
	sw.bindings = append(sw.bindings, b)
	sw.bindMu.Unlock()
}

func (sw *Switch) delBinding(s uintptr) {
	sw.bindMu.Lock()
	for i := range sw.bindings {
		if sw.bindings[i].sysfd != s {
			continue
		}
		sw.bindings = append(sw.bindings[:i], sw.bindings[i+1:]...)
		break
	}
	sw.bindMu.Unlock()
}

// AddFilter deploys the socket system call filter f associated with
// the filter type t on the socket descriptor s.
func (sw *Switch) AddFilter(s uintptr, t FilterType, f Filter) {
	sw.once.Do(sw.init)

	b := sw.posteriorBinding(s)
	if b == nil {
		return
	}
	b.addFilter(t, f)
}

// DelFilter retreats all socket system call filters associated with
// the filter type t on the socket descriptor s.
func (sw *Switch) DelFilter(s uintptr, t FilterType) {
	sw.once.Do(sw.init)

	b := sw.posteriorBinding(s)
	if b != nil {
		return
	}
	b.delFilter(t)
}

// Register registers the callback function fn associated with the
// tag.
//
// The callback function fn is called when the socket descriptor is
// created.
func (sw *Switch) Register(tag string, fn func(s uintptr, cookie Cookie)) {
	sw.once.Do(sw.init)

	sw.noteMu.Lock()
	sw.noteMap[tag] = fn
	sw.noteMu.Unlock()
}

func (sw *Switch) notify(s uintptr, cookie Cookie) {
	sw.noteMu.RLock()
	for _, fn := range sw.noteMap {
		fn(s, cookie)
	}
	sw.noteMu.RUnlock()
}

// Deregister deregisters the callback function associated with the
// tag.
func (sw *Switch) Deregister(tag string) {
	sw.once.Do(sw.init)

	sw.noteMu.Lock()
	delete(sw.noteMap, tag)
	sw.noteMu.Unlock()
}

// Disable disables the switch.
//
// The disabled switch uses only normal callpaths.
func (sw *Switch) Disable() {
	if atomic.LoadUint32(&sw.state) == switchDisabled {
		return
	}
	atomic.StoreUint32(&sw.state, switchDisabled)
}
