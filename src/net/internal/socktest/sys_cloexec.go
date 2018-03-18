// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build dragonfly freebsd linux netbsd openbsd

package socktest

import (
	"sync/atomic"
	"syscall"
)

// Accept4 wraps syscall.Accept4.
func (sw *Switch) Accept4(s, flags int) (ns int, sa syscall.Sockaddr, err error) {
	if atomic.LoadUint32(&sw.state) == switchDisabled {
		return syscall.Accept4(s, flags)
	}

	b := sw.posteriorBinding(uintptr(s))
	if b == nil {
		return syscall.Accept4(s, flags)
	}
	st := b.newState()
	f := b.filter(FilterAccept)

	af, err := f.apply(st)
	if err != nil {
		return -1, nil, err
	}
	ns, sa, st.Err = syscall.Accept4(s, flags)
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
