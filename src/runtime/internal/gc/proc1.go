// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gc

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
	"unsafe"
)

func gcprocs() int32 {
	// Figure out how many CPUs to use during GC.
	// Limited by gomaxprocs, number of actual CPUs, and MaxGcproc.
	_lock.Lock(&_core.Sched.Lock)
	n := _lock.Gomaxprocs
	if n > _lock.Ncpu {
		n = _lock.Ncpu
	}
	if n > _core.MaxGcproc {
		n = _core.MaxGcproc
	}
	if n > _core.Sched.Nmidle+1 { // one M is currently running
		n = _core.Sched.Nmidle + 1
	}
	_lock.Unlock(&_core.Sched.Lock)
	return n
}

func needaddgcproc() bool {
	_lock.Lock(&_core.Sched.Lock)
	n := _lock.Gomaxprocs
	if n > _lock.Ncpu {
		n = _lock.Ncpu
	}
	if n > _core.MaxGcproc {
		n = _core.MaxGcproc
	}
	n -= _core.Sched.Nmidle + 1 // one M is currently running
	_lock.Unlock(&_core.Sched.Lock)
	return n > 0
}

func helpgc(nproc int32) {
	_g_ := _core.Getg()
	_lock.Lock(&_core.Sched.Lock)
	pos := 0
	for n := int32(1); n < nproc; n++ { // one M is currently running
		if _lock.Allp[pos].Mcache == _g_.M.Mcache {
			pos++
		}
		mp := _sched.Mget()
		if mp == nil {
			_lock.Gothrow("gcprocs inconsistency")
		}
		mp.Helpgc = n
		mp.Mcache = _lock.Allp[pos].Mcache
		pos++
		_sched.Notewakeup(&mp.Park)
	}
	_lock.Unlock(&_core.Sched.Lock)
}

// The Gscanstatuses are acting like locks and this releases them.
// If it proves to be a performance hit we should be able to make these
// simple atomic stores but for now we are going to throw if
// we see an inconsistent state.
func Casfrom_Gscanstatus(gp *_core.G, oldval, newval uint32) {
	success := false

	// Check that transition is valid.
	switch oldval {
	default:
		print("runtime: casfrom_Gscanstatus bad oldval gp=", gp, ", oldval=", _core.Hex(oldval), ", newval=", _core.Hex(newval), "\n")
		_sched.Dumpgstatus(gp)
		_lock.Gothrow("casfrom_Gscanstatus:top gp->status is not in scan state")
	case _lock.Gscanrunnable,
		_lock.Gscanwaiting,
		_lock.Gscanrunning,
		_lock.Gscansyscall:
		if newval == oldval&^_lock.Gscan {
			success = _sched.Cas(&gp.Atomicstatus, oldval, newval)
		}
	case _lock.Gscanenqueue:
		if newval == _lock.Gwaiting {
			success = _sched.Cas(&gp.Atomicstatus, oldval, newval)
		}
	}
	if !success {
		print("runtime: casfrom_Gscanstatus failed gp=", gp, ", oldval=", _core.Hex(oldval), ", newval=", _core.Hex(newval), "\n")
		_sched.Dumpgstatus(gp)
		_lock.Gothrow("casfrom_Gscanstatus: gp->status is not in scan state")
	}
}

// This will return false if the gp is not in the expected status and the cas fails.
// This acts like a lock acquire while the casfromgstatus acts like a lock release.
func Castogscanstatus(gp *_core.G, oldval, newval uint32) bool {
	switch oldval {
	case _lock.Grunnable,
		_lock.Gwaiting,
		_lock.Gsyscall:
		if newval == oldval|_lock.Gscan {
			return _sched.Cas(&gp.Atomicstatus, oldval, newval)
		}
	case _lock.Grunning:
		if newval == _lock.Gscanrunning || newval == _lock.Gscanenqueue {
			return _sched.Cas(&gp.Atomicstatus, oldval, newval)
		}
	}
	print("runtime: castogscanstatus oldval=", _core.Hex(oldval), " newval=", _core.Hex(newval), "\n")
	_lock.Gothrow("castogscanstatus")
	panic("not reached")
}

// casgstatus(gp, oldstatus, Gcopystack), assuming oldstatus is Gwaiting or Grunnable.
// Returns old status. Cannot call casgstatus directly, because we are racing with an
// async wakeup that might come in from netpoll. If we see Gwaiting from the readgstatus,
// it might have become Grunnable by the time we get to the cas. If we called casgstatus,
// it would loop waiting for the status to go back to Gwaiting, which it never will.
//go:nosplit
func casgcopystack(gp *_core.G) uint32 {
	for {
		oldstatus := _lock.Readgstatus(gp) &^ _lock.Gscan
		if oldstatus != _lock.Gwaiting && oldstatus != _lock.Grunnable {
			_lock.Gothrow("copystack: bad status, not Gwaiting or Grunnable")
		}
		if _sched.Cas(&gp.Atomicstatus, oldstatus, _lock.Gcopystack) {
			return oldstatus
		}
	}
}

// stopg ensures that gp is stopped at a GC safe point where its stack can be scanned
// or in the context of a moving collector the pointers can be flipped from pointing
// to old object to pointing to new objects.
// If stopg returns true, the caller knows gp is at a GC safe point and will remain there until
// the caller calls restartg.
// If stopg returns false, the caller is not responsible for calling restartg. This can happen
// if another thread, either the gp itself or another GC thread is taking the responsibility
// to do the GC work related to this thread.
func Stopg(gp *_core.G) bool {
	for {
		if gp.Gcworkdone {
			return false
		}

		switch s := _lock.Readgstatus(gp); s {
		default:
			_sched.Dumpgstatus(gp)
			_lock.Gothrow("stopg: gp->atomicstatus is not valid")

		case _lock.Gdead:
			return false

		case _lock.Gcopystack:
			// Loop until a new stack is in place.

		case _lock.Grunnable,
			_lock.Gsyscall,
			_lock.Gwaiting:
			// Claim goroutine by setting scan bit.
			if !Castogscanstatus(gp, s, s|_lock.Gscan) {
				break
			}
			// In scan state, do work.
			Gcphasework(gp)
			return true

		case _lock.Gscanrunnable,
			_lock.Gscanwaiting,
			_lock.Gscansyscall:
			// Goroutine already claimed by another GC helper.
			return false

		case _lock.Grunning:
			// Claim goroutine, so we aren't racing with a status
			// transition away from Grunning.
			if !Castogscanstatus(gp, _lock.Grunning, _lock.Gscanrunning) {
				break
			}

			// Mark gp for preemption.
			if !gp.Gcworkdone {
				gp.Preemptscan = true
				gp.Preempt = true
				gp.Stackguard0 = _lock.StackPreempt
			}

			// Unclaim.
			Casfrom_Gscanstatus(gp, _lock.Gscanrunning, _lock.Grunning)
			return false
		}
	}
}

// The GC requests that this routine be moved from a scanmumble state to a mumble state.
func Restartg(gp *_core.G) {
	s := _lock.Readgstatus(gp)
	switch s {
	default:
		_sched.Dumpgstatus(gp)
		_lock.Gothrow("restartg: unexpected status")

	case _lock.Gdead:
		// ok

	case _lock.Gscanrunnable,
		_lock.Gscanwaiting,
		_lock.Gscansyscall:
		Casfrom_Gscanstatus(gp, s, s&^_lock.Gscan)

	// Scan is now completed.
	// Goroutine now needs to be made runnable.
	// We put it on the global run queue; ready blocks on the global scheduler lock.
	case _lock.Gscanenqueue:
		Casfrom_Gscanstatus(gp, _lock.Gscanenqueue, _lock.Gwaiting)
		if gp != _core.Getg().M.Curg {
			_lock.Gothrow("processing Gscanenqueue on wrong m")
		}
		_sched.Dropg()
		_sched.Ready(gp)
	}
}

// This is used by the GC as well as the routines that do stack dumps. In the case
// of GC all the routines can be reliably stopped. This is not always the case
// when the system is in panic or being exited.
func Stoptheworld() {
	_g_ := _core.Getg()

	// If we hold a lock, then we won't be able to stop another M
	// that is blocked trying to acquire the lock.
	if _g_.M.Locks > 0 {
		_lock.Gothrow("stoptheworld: holding locks")
	}

	_lock.Lock(&_core.Sched.Lock)
	_core.Sched.Stopwait = _lock.Gomaxprocs
	_lock.Atomicstore(&_core.Sched.Gcwaiting, 1)
	_lock.Preemptall()
	// stop current P
	_g_.M.P.Status = _lock.Pgcstop // Pgcstop is only diagnostic.
	_core.Sched.Stopwait--
	// try to retake all P's in Psyscall status
	for i := 0; i < int(_lock.Gomaxprocs); i++ {
		p := _lock.Allp[i]
		s := p.Status
		if s == _lock.Psyscall && _sched.Cas(&p.Status, s, _lock.Pgcstop) {
			_core.Sched.Stopwait--
		}
	}
	// stop idle P's
	for {
		p := _sched.Pidleget()
		if p == nil {
			break
		}
		p.Status = _lock.Pgcstop
		_core.Sched.Stopwait--
	}
	wait := _core.Sched.Stopwait > 0
	_lock.Unlock(&_core.Sched.Lock)

	// wait for remaining P's to stop voluntarily
	if wait {
		for {
			// wait for 100us, then try to re-preempt in case of any races
			if Notetsleep(&_core.Sched.Stopnote, 100*1000) {
				_sched.Noteclear(&_core.Sched.Stopnote)
				break
			}
			_lock.Preemptall()
		}
	}
	if _core.Sched.Stopwait != 0 {
		_lock.Gothrow("stoptheworld: not stopped")
	}
	for i := 0; i < int(_lock.Gomaxprocs); i++ {
		p := _lock.Allp[i]
		if p.Status != _lock.Pgcstop {
			_lock.Gothrow("stoptheworld: not stopped")
		}
	}
}

func mhelpgc() {
	_g_ := _core.Getg()
	_g_.M.Helpgc = -1
}

func Starttheworld() {
	_g_ := _core.Getg()

	_g_.M.Locks++               // disable preemption because it can be holding p in a local var
	gp := _sched.Netpoll(false) // non-blocking
	_sched.Injectglist(gp)
	add := needaddgcproc()
	_lock.Lock(&_core.Sched.Lock)
	if Newprocs != 0 {
		Procresize(Newprocs)
		Newprocs = 0
	} else {
		Procresize(_lock.Gomaxprocs)
	}
	_core.Sched.Gcwaiting = 0

	var p1 *_core.P
	for {
		p := _sched.Pidleget()
		if p == nil {
			break
		}
		// procresize() puts p's with work at the beginning of the list.
		// Once we reach a p without a run queue, the rest don't have one either.
		if p.Runqhead == p.Runqtail {
			_sched.Pidleput(p)
			break
		}
		p.M = _sched.Mget()
		p.Link = p1
		p1 = p
	}
	if _core.Sched.Sysmonwait != 0 {
		_core.Sched.Sysmonwait = 0
		_sched.Notewakeup(&_core.Sched.Sysmonnote)
	}
	_lock.Unlock(&_core.Sched.Lock)

	for p1 != nil {
		p := p1
		p1 = p1.Link
		if p.M != nil {
			mp := p.M
			p.M = nil
			if mp.Nextp != nil {
				_lock.Gothrow("starttheworld: inconsistent mp->nextp")
			}
			mp.Nextp = p
			_sched.Notewakeup(&mp.Park)
		} else {
			// Start M to run P.  Do not start another M below.
			_sched.Newm(nil, p)
			add = false
		}
	}

	if add {
		// If GC could have used another helper proc, start one now,
		// in the hope that it will be available next time.
		// It would have been even better to start it before the collection,
		// but doing so requires allocating memory, so it's tricky to
		// coordinate.  This lazy approach works out in practice:
		// we don't mind if the first couple gc rounds don't have quite
		// the maximum number of procs.
		_sched.Newm(mhelpgc, nil)
	}
	_g_.M.Locks--
	if _g_.M.Locks == 0 && _g_.Preempt { // restore the preemption request in case we've cleared it in newstack
		_g_.Stackguard0 = _lock.StackPreempt
	}
}

// Purge all cached G's from gfree list to the global list.
func gfpurge(_p_ *_core.P) {
	_lock.Lock(&_core.Sched.Gflock)
	for _p_.Gfreecnt != 0 {
		_p_.Gfreecnt--
		gp := _p_.Gfree
		_p_.Gfree = gp.Schedlink
		gp.Schedlink = _core.Sched.Gfree
		_core.Sched.Gfree = gp
		_core.Sched.Ngfree++
	}
	_lock.Unlock(&_core.Sched.Gflock)
}

func Gcount() int32 {
	n := int32(Allglen) - _core.Sched.Ngfree
	for i := 0; ; i++ {
		_p_ := _lock.Allp[i]
		if _p_ == nil {
			break
		}
		n -= _p_.Gfreecnt
	}

	// All these variables can be changed concurrently, so the result can be inconsistent.
	// But at least the current goroutine is running.
	if n < 1 {
		n = 1
	}
	return n
}

// Change number of processors.  The world is stopped, sched is locked.
// gcworkbufs are not being modified by either the GC or
// the write barrier code.
func Procresize(new int32) {
	old := _lock.Gomaxprocs
	if old < 0 || old > _lock.MaxGomaxprocs || new <= 0 || new > _lock.MaxGomaxprocs {
		_lock.Gothrow("procresize: invalid arg")
	}

	// initialize new P's
	for i := int32(0); i < new; i++ {
		p := _lock.Allp[i]
		if p == nil {
			p = newP()
			p.Id = i
			p.Status = _lock.Pgcstop
			_sched.Atomicstorep(unsafe.Pointer(&_lock.Allp[i]), unsafe.Pointer(p))
		}
		if p.Mcache == nil {
			if old == 0 && i == 0 {
				if _core.Getg().M.Mcache == nil {
					_lock.Gothrow("missing mcache?")
				}
				p.Mcache = _core.Getg().M.Mcache // bootstrap
			} else {
				p.Mcache = _lock.Allocmcache()
			}
		}
	}

	// redistribute runnable G's evenly
	// collect all runnable goroutines in global queue preserving FIFO order
	// FIFO order is required to ensure fairness even during frequent GCs
	// see http://golang.org/issue/7126
	empty := false
	for !empty {
		empty = true
		for i := int32(0); i < old; i++ {
			p := _lock.Allp[i]
			if p.Runqhead == p.Runqtail {
				continue
			}
			empty = false
			// pop from tail of local queue
			p.Runqtail--
			gp := p.Runq[p.Runqtail%uint32(len(p.Runq))]
			// push onto head of global queue
			gp.Schedlink = _core.Sched.Runqhead
			_core.Sched.Runqhead = gp
			if _core.Sched.Runqtail == nil {
				_core.Sched.Runqtail = gp
			}
			_core.Sched.Runqsize++
		}
	}

	// fill local queues with at most len(p.runq)/2 goroutines
	// start at 1 because current M already executes some G and will acquire allp[0] below,
	// so if we have a spare G we want to put it into allp[1].
	var _p_ _core.P
	for i := int32(1); i < new*int32(len(_p_.Runq))/2 && _core.Sched.Runqsize > 0; i++ {
		gp := _core.Sched.Runqhead
		_core.Sched.Runqhead = gp.Schedlink
		if _core.Sched.Runqhead == nil {
			_core.Sched.Runqtail = nil
		}
		_core.Sched.Runqsize--
		_sched.Runqput(_lock.Allp[i%new], gp)
	}

	// free unused P's
	for i := new; i < old; i++ {
		p := _lock.Allp[i]
		freemcache(p.Mcache)
		p.Mcache = nil
		gfpurge(p)
		p.Status = _lock.Pdead
		// can't free P itself because it can be referenced by an M in syscall
	}

	_g_ := _core.Getg()
	if _g_.M.P != nil {
		_g_.M.P.M = nil
	}
	_g_.M.P = nil
	_g_.M.Mcache = nil
	p := _lock.Allp[0]
	p.M = nil
	p.Status = _lock.Pidle
	_sched.Acquirep(p)
	for i := new - 1; i > 0; i-- {
		p := _lock.Allp[i]
		p.Status = _lock.Pidle
		_sched.Pidleput(p)
	}
	var int32p *int32 = &_lock.Gomaxprocs // make compiler check that gomaxprocs is an int32
	_lock.Atomicstore((*uint32)(unsafe.Pointer(int32p)), uint32(new))
}
