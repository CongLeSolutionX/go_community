// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris windows

package runtime

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	_netpoll "runtime/internal/netpoll"
	_sched "runtime/internal/sched"
	"unsafe"
)

const pollBlockSize = 4 * 1024

type pollCache struct {
	lock  _core.Mutex
	first *_sched.PollDesc
	// PollDesc objects must be type-stable,
	// because we can get ready notification from epoll/kqueue
	// after the descriptor is closed/reused.
	// Stale notifications are detected using seq variable,
	// seq is incremented when deadlines are changed or descriptor is reused.
}

var (
	pollcache pollCache
)

//go:linkname net_runtime_pollServerInit net.runtime_pollServerInit
func net_runtime_pollServerInit() {
	_netpoll.Netpollinit()
	_lock.Atomicstore(&_sched.NetpollInited, 1)
}

//go:linkname net_runtime_pollOpen net.runtime_pollOpen
func net_runtime_pollOpen(fd uintptr) (*_sched.PollDesc, int) {
	pd := pollcache.alloc()
	_lock.Lock(&pd.Lock)
	if pd.Wg != 0 && pd.Wg != _sched.PdReady {
		_lock.Throw("netpollOpen: blocked write on free descriptor")
	}
	if pd.Rg != 0 && pd.Rg != _sched.PdReady {
		_lock.Throw("netpollOpen: blocked read on free descriptor")
	}
	pd.Fd = fd
	pd.Closing = false
	pd.Seq++
	pd.Rg = 0
	pd.Rd = 0
	pd.Wg = 0
	pd.Wd = 0
	_lock.Unlock(&pd.Lock)

	var errno int32
	errno = _netpoll.Netpollopen(fd, pd)
	return pd, int(errno)
}

//go:linkname net_runtime_pollClose net.runtime_pollClose
func net_runtime_pollClose(pd *_sched.PollDesc) {
	if !pd.Closing {
		_lock.Throw("netpollClose: close w/o unblock")
	}
	if pd.Wg != 0 && pd.Wg != _sched.PdReady {
		_lock.Throw("netpollClose: blocked write on closing descriptor")
	}
	if pd.Rg != 0 && pd.Rg != _sched.PdReady {
		_lock.Throw("netpollClose: blocked read on closing descriptor")
	}
	_netpoll.Netpollclose(uintptr(pd.Fd))
	pollcache.free(pd)
}

func (c *pollCache) free(pd *_sched.PollDesc) {
	_lock.Lock(&c.lock)
	pd.Link = c.first
	c.first = pd
	_lock.Unlock(&c.lock)
}

//go:linkname net_runtime_pollReset net.runtime_pollReset
func net_runtime_pollReset(pd *_sched.PollDesc, mode int) int {
	err := _netpoll.Netpollcheckerr(pd, int32(mode))
	if err != 0 {
		return err
	}
	if mode == 'r' {
		pd.Rg = 0
	} else if mode == 'w' {
		pd.Wg = 0
	}
	return 0
}

//go:linkname net_runtime_pollWait net.runtime_pollWait
func net_runtime_pollWait(pd *_sched.PollDesc, mode int) int {
	err := _netpoll.Netpollcheckerr(pd, int32(mode))
	if err != 0 {
		return err
	}
	// As for now only Solaris uses level-triggered IO.
	if _lock.GOOS == "solaris" {
		_netpoll.Netpollarm(pd, mode)
	}
	for !_netpoll.Netpollblock(pd, int32(mode), false) {
		err = _netpoll.Netpollcheckerr(pd, int32(mode))
		if err != 0 {
			return err
		}
		// Can happen if timeout has fired and unblocked us,
		// but before we had a chance to run, timeout has been reset.
		// Pretend it has not happened and retry.
	}
	return 0
}

//go:linkname net_runtime_pollWaitCanceled net.runtime_pollWaitCanceled
func net_runtime_pollWaitCanceled(pd *_sched.PollDesc, mode int) {
	// This function is used only on windows after a failed attempt to cancel
	// a pending async IO operation. Wait for ioready, ignore closing or timeouts.
	for !_netpoll.Netpollblock(pd, int32(mode), true) {
	}
}

//go:linkname net_runtime_pollSetDeadline net.runtime_pollSetDeadline
func net_runtime_pollSetDeadline(pd *_sched.PollDesc, d int64, mode int) {
	_lock.Lock(&pd.Lock)
	if pd.Closing {
		_lock.Unlock(&pd.Lock)
		return
	}
	pd.Seq++ // invalidate current timers
	// Reset current timers.
	if pd.Rt.F != nil {
		deltimer(&pd.Rt)
		pd.Rt.F = nil
	}
	if pd.Wt.F != nil {
		deltimer(&pd.Wt)
		pd.Wt.F = nil
	}
	// Setup new timers.
	if d != 0 && d <= _lock.Nanotime() {
		d = -1
	}
	if mode == 'r' || mode == 'r'+'w' {
		pd.Rd = d
	}
	if mode == 'w' || mode == 'r'+'w' {
		pd.Wd = d
	}
	if pd.Rd > 0 && pd.Rd == pd.Wd {
		pd.Rt.F = _netpoll.NetpollDeadline
		pd.Rt.When = pd.Rd
		// Copy current seq into the timer arg.
		// Timer func will check the seq against current descriptor seq,
		// if they differ the descriptor was reused or timers were reset.
		pd.Rt.Arg = pd
		pd.Rt.Seq = pd.Seq
		_sched.Addtimer(&pd.Rt)
	} else {
		if pd.Rd > 0 {
			pd.Rt.F = _netpoll.NetpollReadDeadline
			pd.Rt.When = pd.Rd
			pd.Rt.Arg = pd
			pd.Rt.Seq = pd.Seq
			_sched.Addtimer(&pd.Rt)
		}
		if pd.Wd > 0 {
			pd.Wt.F = _netpoll.NetpollWriteDeadline
			pd.Wt.When = pd.Wd
			pd.Wt.Arg = pd
			pd.Wt.Seq = pd.Seq
			_sched.Addtimer(&pd.Wt)
		}
	}
	// If we set the new deadline in the past, unblock currently pending IO if any.
	var rg, wg *_core.G
	_sched.Atomicstorep(unsafe.Pointer(&wg), nil) // full memory barrier between stores to rd/wd and load of rg/wg in netpollunblock
	if pd.Rd < 0 {
		rg = _sched.Netpollunblock(pd, 'r', false)
	}
	if pd.Wd < 0 {
		wg = _sched.Netpollunblock(pd, 'w', false)
	}
	_lock.Unlock(&pd.Lock)
	if rg != nil {
		_sched.Goready(rg)
	}
	if wg != nil {
		_sched.Goready(wg)
	}
}

//go:linkname net_runtime_pollUnblock net.runtime_pollUnblock
func net_runtime_pollUnblock(pd *_sched.PollDesc) {
	_lock.Lock(&pd.Lock)
	if pd.Closing {
		_lock.Throw("netpollUnblock: already closing")
	}
	pd.Closing = true
	pd.Seq++
	var rg, wg *_core.G
	_sched.Atomicstorep(unsafe.Pointer(&rg), nil) // full memory barrier between store to closing and read of rg/wg in netpollunblock
	rg = _sched.Netpollunblock(pd, 'r', false)
	wg = _sched.Netpollunblock(pd, 'w', false)
	if pd.Rt.F != nil {
		deltimer(&pd.Rt)
		pd.Rt.F = nil
	}
	if pd.Wt.F != nil {
		deltimer(&pd.Wt)
		pd.Wt.F = nil
	}
	_lock.Unlock(&pd.Lock)
	if rg != nil {
		_sched.Goready(rg)
	}
	if wg != nil {
		_sched.Goready(wg)
	}
}

func (c *pollCache) alloc() *_sched.PollDesc {
	_lock.Lock(&c.lock)
	if c.first == nil {
		const pdSize = unsafe.Sizeof(_sched.PollDesc{})
		n := pollBlockSize / pdSize
		if n == 0 {
			n = 1
		}
		// Must be in non-GC memory because can be referenced
		// only from epoll/kqueue internals.
		mem := _lock.Persistentalloc(n*pdSize, 0, &_lock.Memstats.Other_sys)
		for i := uintptr(0); i < n; i++ {
			pd := (*_sched.PollDesc)(_core.Add(mem, i*pdSize))
			pd.Link = c.first
			c.first = pd
		}
	}
	pd := c.first
	c.first = pd.Link
	_lock.Unlock(&c.lock)
	return pd
}
