// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris windows

package runtime

import (
	_base "runtime/internal/base"
	_gc "runtime/internal/gc"
	"unsafe"
)

const pollBlockSize = 4 * 1024

type pollCache struct {
	lock  _base.Mutex
	first *_base.PollDesc
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
	netpollinit()
	_base.Atomicstore(&_base.NetpollInited, 1)
}

//go:linkname net_runtime_pollOpen net.runtime_pollOpen
func net_runtime_pollOpen(fd uintptr) (*_base.PollDesc, int) {
	pd := pollcache.alloc()
	_base.Lock(&pd.Lock)
	if pd.Wg != 0 && pd.Wg != _base.PdReady {
		_base.Throw("netpollOpen: blocked write on free descriptor")
	}
	if pd.Rg != 0 && pd.Rg != _base.PdReady {
		_base.Throw("netpollOpen: blocked read on free descriptor")
	}
	pd.Fd = fd
	pd.Closing = false
	pd.Seq++
	pd.Rg = 0
	pd.Rd = 0
	pd.Wg = 0
	pd.Wd = 0
	_base.Unlock(&pd.Lock)

	var errno int32
	errno = netpollopen(fd, pd)
	return pd, int(errno)
}

//go:linkname net_runtime_pollClose net.runtime_pollClose
func net_runtime_pollClose(pd *_base.PollDesc) {
	if !pd.Closing {
		_base.Throw("netpollClose: close w/o unblock")
	}
	if pd.Wg != 0 && pd.Wg != _base.PdReady {
		_base.Throw("netpollClose: blocked write on closing descriptor")
	}
	if pd.Rg != 0 && pd.Rg != _base.PdReady {
		_base.Throw("netpollClose: blocked read on closing descriptor")
	}
	netpollclose(uintptr(pd.Fd))
	pollcache.free(pd)
}

func (c *pollCache) free(pd *_base.PollDesc) {
	_base.Lock(&c.lock)
	pd.Link = c.first
	c.first = pd
	_base.Unlock(&c.lock)
}

//go:linkname net_runtime_pollReset net.runtime_pollReset
func net_runtime_pollReset(pd *_base.PollDesc, mode int) int {
	err := netpollcheckerr(pd, int32(mode))
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
func net_runtime_pollWait(pd *_base.PollDesc, mode int) int {
	err := netpollcheckerr(pd, int32(mode))
	if err != 0 {
		return err
	}
	// As for now only Solaris uses level-triggered IO.
	if _base.GOOS == "solaris" {
		netpollarm(pd, mode)
	}
	for !netpollblock(pd, int32(mode), false) {
		err = netpollcheckerr(pd, int32(mode))
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
func net_runtime_pollWaitCanceled(pd *_base.PollDesc, mode int) {
	// This function is used only on windows after a failed attempt to cancel
	// a pending async IO operation. Wait for ioready, ignore closing or timeouts.
	for !netpollblock(pd, int32(mode), true) {
	}
}

//go:linkname net_runtime_pollSetDeadline net.runtime_pollSetDeadline
func net_runtime_pollSetDeadline(pd *_base.PollDesc, d int64, mode int) {
	_base.Lock(&pd.Lock)
	if pd.Closing {
		_base.Unlock(&pd.Lock)
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
	if d != 0 && d <= _base.Nanotime() {
		d = -1
	}
	if mode == 'r' || mode == 'r'+'w' {
		pd.Rd = d
	}
	if mode == 'w' || mode == 'r'+'w' {
		pd.Wd = d
	}
	if pd.Rd > 0 && pd.Rd == pd.Wd {
		pd.Rt.F = netpollDeadline
		pd.Rt.When = pd.Rd
		// Copy current seq into the timer arg.
		// Timer func will check the seq against current descriptor seq,
		// if they differ the descriptor was reused or timers were reset.
		pd.Rt.Arg = pd
		pd.Rt.Seq = pd.Seq
		addtimer(&pd.Rt)
	} else {
		if pd.Rd > 0 {
			pd.Rt.F = netpollReadDeadline
			pd.Rt.When = pd.Rd
			pd.Rt.Arg = pd
			pd.Rt.Seq = pd.Seq
			addtimer(&pd.Rt)
		}
		if pd.Wd > 0 {
			pd.Wt.F = netpollWriteDeadline
			pd.Wt.When = pd.Wd
			pd.Wt.Arg = pd
			pd.Wt.Seq = pd.Seq
			addtimer(&pd.Wt)
		}
	}
	// If we set the new deadline in the past, unblock currently pending IO if any.
	var rg, wg *_base.G
	_base.Atomicstorep(unsafe.Pointer(&wg), nil) // full memory barrier between stores to rd/wd and load of rg/wg in netpollunblock
	if pd.Rd < 0 {
		rg = _base.Netpollunblock(pd, 'r', false)
	}
	if pd.Wd < 0 {
		wg = _base.Netpollunblock(pd, 'w', false)
	}
	_base.Unlock(&pd.Lock)
	if rg != nil {
		_gc.Goready(rg, 3)
	}
	if wg != nil {
		_gc.Goready(wg, 3)
	}
}

//go:linkname net_runtime_pollUnblock net.runtime_pollUnblock
func net_runtime_pollUnblock(pd *_base.PollDesc) {
	_base.Lock(&pd.Lock)
	if pd.Closing {
		_base.Throw("netpollUnblock: already closing")
	}
	pd.Closing = true
	pd.Seq++
	var rg, wg *_base.G
	_base.Atomicstorep(unsafe.Pointer(&rg), nil) // full memory barrier between store to closing and read of rg/wg in netpollunblock
	rg = _base.Netpollunblock(pd, 'r', false)
	wg = _base.Netpollunblock(pd, 'w', false)
	if pd.Rt.F != nil {
		deltimer(&pd.Rt)
		pd.Rt.F = nil
	}
	if pd.Wt.F != nil {
		deltimer(&pd.Wt)
		pd.Wt.F = nil
	}
	_base.Unlock(&pd.Lock)
	if rg != nil {
		_gc.Goready(rg, 3)
	}
	if wg != nil {
		_gc.Goready(wg, 3)
	}
}

func netpollcheckerr(pd *_base.PollDesc, mode int32) int {
	if pd.Closing {
		return 1 // errClosing
	}
	if (mode == 'r' && pd.Rd < 0) || (mode == 'w' && pd.Wd < 0) {
		return 2 // errTimeout
	}
	return 0
}

func netpollblockcommit(gp *_base.G, gpp unsafe.Pointer) bool {
	return _base.Casuintptr((*uintptr)(gpp), _base.PdWait, uintptr(unsafe.Pointer(gp)))
}

// returns true if IO is ready, or false if timedout or closed
// waitio - wait only for completed IO, ignore errors
func netpollblock(pd *_base.PollDesc, mode int32, waitio bool) bool {
	gpp := &pd.Rg
	if mode == 'w' {
		gpp = &pd.Wg
	}

	// set the gpp semaphore to WAIT
	for {
		old := *gpp
		if old == _base.PdReady {
			*gpp = 0
			return true
		}
		if old != 0 {
			_base.Throw("netpollblock: double wait")
		}
		if _base.Casuintptr(gpp, 0, _base.PdWait) {
			break
		}
	}

	// need to recheck error states after setting gpp to WAIT
	// this is necessary because runtime_pollUnblock/runtime_pollSetDeadline/deadlineimpl
	// do the opposite: store to closing/rd/wd, membarrier, load of rg/wg
	if waitio || netpollcheckerr(pd, mode) == 0 {
		_base.Gopark(netpollblockcommit, unsafe.Pointer(gpp), "IO wait", _base.TraceEvGoBlockNet, 5)
	}
	// be careful to not lose concurrent READY notification
	old := xchguintptr(gpp, 0)
	if old > _base.PdWait {
		_base.Throw("netpollblock: corrupted state")
	}
	return old == _base.PdReady
}

func netpolldeadlineimpl(pd *_base.PollDesc, seq uintptr, read, write bool) {
	_base.Lock(&pd.Lock)
	// Seq arg is seq when the timer was set.
	// If it's stale, ignore the timer event.
	if seq != pd.Seq {
		// The descriptor was reused or timers were reset.
		_base.Unlock(&pd.Lock)
		return
	}
	var rg *_base.G
	if read {
		if pd.Rd <= 0 || pd.Rt.F == nil {
			_base.Throw("netpolldeadlineimpl: inconsistent read deadline")
		}
		pd.Rd = -1
		_base.Atomicstorep(unsafe.Pointer(&pd.Rt.F), nil) // full memory barrier between store to rd and load of rg in netpollunblock
		rg = _base.Netpollunblock(pd, 'r', false)
	}
	var wg *_base.G
	if write {
		if pd.Wd <= 0 || pd.Wt.F == nil && !read {
			_base.Throw("netpolldeadlineimpl: inconsistent write deadline")
		}
		pd.Wd = -1
		_base.Atomicstorep(unsafe.Pointer(&pd.Wt.F), nil) // full memory barrier between store to wd and load of wg in netpollunblock
		wg = _base.Netpollunblock(pd, 'w', false)
	}
	_base.Unlock(&pd.Lock)
	if rg != nil {
		_gc.Goready(rg, 0)
	}
	if wg != nil {
		_gc.Goready(wg, 0)
	}
}

func netpollDeadline(arg interface{}, seq uintptr) {
	netpolldeadlineimpl(arg.(*_base.PollDesc), seq, true, true)
}

func netpollReadDeadline(arg interface{}, seq uintptr) {
	netpolldeadlineimpl(arg.(*_base.PollDesc), seq, true, false)
}

func netpollWriteDeadline(arg interface{}, seq uintptr) {
	netpolldeadlineimpl(arg.(*_base.PollDesc), seq, false, true)
}

func (c *pollCache) alloc() *_base.PollDesc {
	_base.Lock(&c.lock)
	if c.first == nil {
		const pdSize = unsafe.Sizeof(_base.PollDesc{})
		n := pollBlockSize / pdSize
		if n == 0 {
			n = 1
		}
		// Must be in non-GC memory because can be referenced
		// only from epoll/kqueue internals.
		mem := _base.Persistentalloc(n*pdSize, 0, &_base.Memstats.Other_sys)
		for i := uintptr(0); i < n; i++ {
			pd := (*_base.PollDesc)(_base.Add(mem, i*pdSize))
			pd.Link = c.first
			c.first = pd
		}
	}
	pd := c.first
	c.first = pd.Link
	_base.Unlock(&c.lock)
	return pd
}
