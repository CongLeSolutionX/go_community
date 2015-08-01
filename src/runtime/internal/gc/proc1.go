// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gc

import (
	_base "runtime/internal/base"
	"unsafe"
)

func gcprocs() int32 {
	// Figure out how many CPUs to use during GC.
	// Limited by gomaxprocs, number of actual CPUs, and MaxGcproc.
	_base.Lock(&_base.Sched.Lock)
	n := _base.Gomaxprocs
	if n > _base.Ncpu {
		n = _base.Ncpu
	}
	if n > _base.MaxGcproc {
		n = _base.MaxGcproc
	}
	if n > _base.Sched.Nmidle+1 { // one M is currently running
		n = _base.Sched.Nmidle + 1
	}
	_base.Unlock(&_base.Sched.Lock)
	return n
}

func needaddgcproc() bool {
	_base.Lock(&_base.Sched.Lock)
	n := _base.Gomaxprocs
	if n > _base.Ncpu {
		n = _base.Ncpu
	}
	if n > _base.MaxGcproc {
		n = _base.MaxGcproc
	}
	n -= _base.Sched.Nmidle + 1 // one M is currently running
	_base.Unlock(&_base.Sched.Lock)
	return n > 0
}

func helpgc(nproc int32) {
	_g_ := _base.Getg()
	_base.Lock(&_base.Sched.Lock)
	pos := 0
	for n := int32(1); n < nproc; n++ { // one M is currently running
		if _base.Allp[pos].Mcache == _g_.M.Mcache {
			pos++
		}
		mp := _base.Mget()
		if mp == nil {
			_base.Throw("gcprocs inconsistency")
		}
		mp.Helpgc = n
		mp.P.Set(_base.Allp[pos])
		mp.Mcache = _base.Allp[pos].Mcache
		pos++
		_base.Notewakeup(&mp.Park)
	}
	_base.Unlock(&_base.Sched.Lock)
}

// Ownership of gscanvalid:
//
// If gp is running (meaning status == _Grunning or _Grunning|_Gscan),
// then gp owns gp.gscanvalid, and other goroutines must not modify it.
//
// Otherwise, a second goroutine can lock the scan state by setting _Gscan
// in the status bit and then modify gscanvalid, and then unlock the scan state.
//
// Note that the first condition implies an exception to the second:
// if a second goroutine changes gp's status to _Grunning|_Gscan,
// that second goroutine still does not have the right to modify gscanvalid.

// The Gscanstatuses are acting like locks and this releases them.
// If it proves to be a performance hit we should be able to make these
// simple atomic stores but for now we are going to throw if
// we see an inconsistent state.
func Casfrom_Gscanstatus(gp *_base.G, oldval, newval uint32) {
	success := false

	// Check that transition is valid.
	switch oldval {
	default:
		print("runtime: casfrom_Gscanstatus bad oldval gp=", gp, ", oldval=", _base.Hex(oldval), ", newval=", _base.Hex(newval), "\n")
		_base.Dumpgstatus(gp)
		_base.Throw("casfrom_Gscanstatus:top gp->status is not in scan state")
	case _base.Gscanrunnable,
		_base.Gscanwaiting,
		_base.Gscanrunning,
		_base.Gscansyscall:
		if newval == oldval&^_base.Gscan {
			success = _base.Cas(&gp.Atomicstatus, oldval, newval)
		}
	case _base.Gscanenqueue:
		if newval == _base.Gwaiting {
			success = _base.Cas(&gp.Atomicstatus, oldval, newval)
		}
	}
	if !success {
		print("runtime: casfrom_Gscanstatus failed gp=", gp, ", oldval=", _base.Hex(oldval), ", newval=", _base.Hex(newval), "\n")
		_base.Dumpgstatus(gp)
		_base.Throw("casfrom_Gscanstatus: gp->status is not in scan state")
	}
	if newval == _base.Grunning {
		gp.Gcscanvalid = false
	}
}

// This will return false if the gp is not in the expected status and the cas fails.
// This acts like a lock acquire while the casfromgstatus acts like a lock release.
func Castogscanstatus(gp *_base.G, oldval, newval uint32) bool {
	switch oldval {
	case _base.Grunnable,
		_base.Gwaiting,
		_base.Gsyscall:
		if newval == oldval|_base.Gscan {
			return _base.Cas(&gp.Atomicstatus, oldval, newval)
		}
	case _base.Grunning:
		if newval == _base.Gscanrunning || newval == _base.Gscanenqueue {
			return _base.Cas(&gp.Atomicstatus, oldval, newval)
		}
	}
	print("runtime: castogscanstatus oldval=", _base.Hex(oldval), " newval=", _base.Hex(newval), "\n")
	_base.Throw("castogscanstatus")
	panic("not reached")
}

// casgstatus(gp, oldstatus, Gcopystack), assuming oldstatus is Gwaiting or Grunnable.
// Returns old status. Cannot call casgstatus directly, because we are racing with an
// async wakeup that might come in from netpoll. If we see Gwaiting from the readgstatus,
// it might have become Grunnable by the time we get to the cas. If we called casgstatus,
// it would loop waiting for the status to go back to Gwaiting, which it never will.
//go:nosplit
func casgcopystack(gp *_base.G) uint32 {
	for {
		oldstatus := _base.Readgstatus(gp) &^ _base.Gscan
		if oldstatus != _base.Gwaiting && oldstatus != _base.Grunnable {
			_base.Throw("copystack: bad status, not Gwaiting or Grunnable")
		}
		if _base.Cas(&gp.Atomicstatus, oldstatus, _base.Gcopystack) {
			return oldstatus
		}
	}
}

// scang blocks until gp's stack has been scanned.
// It might be scanned by scang or it might be scanned by the goroutine itself.
// Either way, the stack scan has completed when scang returns.
func scang(gp *_base.G) {
	// Invariant; we (the caller, markroot for a specific goroutine) own gp.gcscandone.
	// Nothing is racing with us now, but gcscandone might be set to true left over
	// from an earlier round of stack scanning (we scan twice per GC).
	// We use gcscandone to record whether the scan has been done during this round.
	// It is important that the scan happens exactly once: if called twice,
	// the installation of stack barriers will detect the double scan and die.

	gp.Gcscandone = false

	// Endeavor to get gcscandone set to true,
	// either by doing the stack scan ourselves or by coercing gp to scan itself.
	// gp.gcscandone can transition from false to true when we're not looking
	// (if we asked for preemption), so any time we lock the status using
	// castogscanstatus we have to double-check that the scan is still not done.
	for !gp.Gcscandone {
		switch s := _base.Readgstatus(gp); s {
		default:
			_base.Dumpgstatus(gp)
			_base.Throw("stopg: invalid status")

		case _base.Gdead:
			// No stack.
			gp.Gcscandone = true

		case _base.Gcopystack:
			// Stack being switched. Go around again.

		case _base.Grunnable, _base.Gsyscall, _base.Gwaiting:
			// Claim goroutine by setting scan bit.
			// Racing with execution or readying of gp.
			// The scan bit keeps them from running
			// the goroutine until we're done.
			if Castogscanstatus(gp, s, s|_base.Gscan) {
				if !gp.Gcscandone {
					Scanstack(gp)
					gp.Gcscandone = true
				}
				restartg(gp)
			}

		case _base.Gscanwaiting:
			// newstack is doing a scan for us right now. Wait.

		case _base.Grunning:
			// Goroutine running. Try to preempt execution so it can scan itself.
			// The preemption handler (in newstack) does the actual scan.

			// Optimization: if there is already a pending preemption request
			// (from the previous loop iteration), don't bother with the atomics.
			if gp.Preemptscan && gp.Preempt && gp.Stackguard0 == _base.StackPreempt {
				break
			}

			// Ask for preemption and self scan.
			if Castogscanstatus(gp, _base.Grunning, _base.Gscanrunning) {
				if !gp.Gcscandone {
					gp.Preemptscan = true
					gp.Preempt = true
					gp.Stackguard0 = _base.StackPreempt
				}
				Casfrom_Gscanstatus(gp, _base.Gscanrunning, _base.Grunning)
			}
		}
	}

	gp.Preemptscan = false // cancel scan request if no longer needed
}

// The GC requests that this routine be moved from a scanmumble state to a mumble state.
func restartg(gp *_base.G) {
	s := _base.Readgstatus(gp)
	switch s {
	default:
		_base.Dumpgstatus(gp)
		_base.Throw("restartg: unexpected status")

	case _base.Gdead:
		// ok

	case _base.Gscanrunnable,
		_base.Gscanwaiting,
		_base.Gscansyscall:
		Casfrom_Gscanstatus(gp, s, s&^_base.Gscan)

	// Scan is now completed.
	// Goroutine now needs to be made runnable.
	// We put it on the global run queue; ready blocks on the global scheduler lock.
	case _base.Gscanenqueue:
		Casfrom_Gscanstatus(gp, _base.Gscanenqueue, _base.Gwaiting)
		if gp != _base.Getg().M.Curg {
			_base.Throw("processing Gscanenqueue on wrong m")
		}
		_base.Dropg()
		_base.Ready(gp, 0)
	}
}

// Holding worldsema grants an M the right to try to stop the world
// and prevents gomaxprocs from changing concurrently.
var Worldsema uint32 = 1

// stopTheWorldWithSema is the core implementation of stopTheWorld.
// The caller is responsible for acquiring worldsema and disabling
// preemption first and then should stopTheWorldWithSema on the system
// stack:
//
//	semacquire(&worldsema, false)
//	m.preemptoff = "reason"
//	systemstack(stopTheWorldWithSema)
//
// When finished, the caller must either call startTheWorld or undo
// these three operations separately:
//
//	m.preemptoff = ""
//	systemstack(startTheWorldWithSema)
//	semrelease(&worldsema)
//
// It is allowed to acquire worldsema once and then execute multiple
// startTheWorldWithSema/stopTheWorldWithSema pairs.
// Other P's are able to execute between successive calls to
// startTheWorldWithSema and stopTheWorldWithSema.
// Holding worldsema causes any other goroutines invoking
// stopTheWorld to block.
func StopTheWorldWithSema() {
	_g_ := _base.Getg()

	// If we hold a lock, then we won't be able to stop another M
	// that is blocked trying to acquire the lock.
	if _g_.M.Locks > 0 {
		_base.Throw("stopTheWorld: holding locks")
	}

	_base.Lock(&_base.Sched.Lock)
	_base.Sched.Stopwait = _base.Gomaxprocs
	_base.Atomicstore(&_base.Sched.Gcwaiting, 1)
	_base.Preemptall()
	// stop current P
	_g_.M.P.Ptr().Status = _base.Pgcstop // Pgcstop is only diagnostic.
	_base.Sched.Stopwait--
	// try to retake all P's in Psyscall status
	for i := 0; i < int(_base.Gomaxprocs); i++ {
		p := _base.Allp[i]
		s := p.Status
		if s == _base.Psyscall && _base.Cas(&p.Status, s, _base.Pgcstop) {
			if _base.Trace.Enabled {
				_base.TraceGoSysBlock(p)
				_base.TraceProcStop(p)
			}
			p.Syscalltick++
			_base.Sched.Stopwait--
		}
	}
	// stop idle P's
	for {
		p := _base.Pidleget()
		if p == nil {
			break
		}
		p.Status = _base.Pgcstop
		_base.Sched.Stopwait--
	}
	wait := _base.Sched.Stopwait > 0
	_base.Unlock(&_base.Sched.Lock)

	// wait for remaining P's to stop voluntarily
	if wait {
		for {
			// wait for 100us, then try to re-preempt in case of any races
			if Notetsleep(&_base.Sched.Stopnote, 100*1000) {
				_base.Noteclear(&_base.Sched.Stopnote)
				break
			}
			_base.Preemptall()
		}
	}
	if _base.Sched.Stopwait != 0 {
		_base.Throw("stopTheWorld: not stopped")
	}
	for i := 0; i < int(_base.Gomaxprocs); i++ {
		p := _base.Allp[i]
		if p.Status != _base.Pgcstop {
			_base.Throw("stopTheWorld: not stopped")
		}
	}
}

func mhelpgc() {
	_g_ := _base.Getg()
	_g_.M.Helpgc = -1
}

func StartTheWorldWithSema() {
	_g_ := _base.Getg()

	_g_.M.Locks++              // disable preemption because it can be holding p in a local var
	gp := _base.Netpoll(false) // non-blocking
	_base.Injectglist(gp)
	add := needaddgcproc()
	_base.Lock(&_base.Sched.Lock)

	procs := _base.Gomaxprocs
	if Newprocs != 0 {
		procs = Newprocs
		Newprocs = 0
	}
	p1 := Procresize(procs)
	_base.Sched.Gcwaiting = 0
	if _base.Sched.Sysmonwait != 0 {
		_base.Sched.Sysmonwait = 0
		_base.Notewakeup(&_base.Sched.Sysmonnote)
	}
	_base.Unlock(&_base.Sched.Lock)

	for p1 != nil {
		p := p1
		p1 = p1.Link.Ptr()
		if p.M != 0 {
			mp := p.M.Ptr()
			p.M = 0
			if mp.Nextp != 0 {
				_base.Throw("startTheWorld: inconsistent mp->nextp")
			}
			mp.Nextp.Set(p)
			_base.Notewakeup(&mp.Park)
		} else {
			// Start M to run P.  Do not start another M below.
			_base.Newm(nil, p)
			add = false
		}
	}

	// Wakeup an additional proc in case we have excessive runnable goroutines
	// in local queues or in the global queue. If we don't, the proc will park itself.
	// If we have lots of excessive work, resetspinning will unpark additional procs as necessary.
	if _base.Atomicload(&_base.Sched.Npidle) != 0 && _base.Atomicload(&_base.Sched.Nmspinning) == 0 {
		_base.Wakep()
	}

	if add {
		// If GC could have used another helper proc, start one now,
		// in the hope that it will be available next time.
		// It would have been even better to start it before the collection,
		// but doing so requires allocating memory, so it's tricky to
		// coordinate.  This lazy approach works out in practice:
		// we don't mind if the first couple gc rounds don't have quite
		// the maximum number of procs.
		_base.Newm(mhelpgc, nil)
	}
	_g_.M.Locks--
	if _g_.M.Locks == 0 && _g_.Preempt { // restore the preemption request in case we've cleared it in newstack
		_g_.Stackguard0 = _base.StackPreempt
	}
}

// forEachP calls fn(p) for every P p when p reaches a GC safe point.
// If a P is currently executing code, this will bring the P to a GC
// safe point and execute fn on that P. If the P is not executing code
// (it is idle or in a syscall), this will call fn(p) directly while
// preventing the P from exiting its state. This does not ensure that
// fn will run on every CPU executing Go code, but it acts as a global
// memory barrier. GC uses this as a "ragged barrier."
//
// The caller must hold worldsema.
func forEachP(fn func(*_base.P)) {
	mp := _base.Acquirem()
	_p_ := _base.Getg().M.P.Ptr()

	_base.Lock(&_base.Sched.Lock)
	if _base.Sched.SafePointWait != 0 {
		_base.Throw("forEachP: sched.safePointWait != 0")
	}
	_base.Sched.SafePointWait = _base.Gomaxprocs - 1
	_base.Sched.SafePointFn = fn

	// Ask all Ps to run the safe point function.
	for _, p := range _base.Allp[:_base.Gomaxprocs] {
		if p != _p_ {
			_base.Atomicstore(&p.RunSafePointFn, 1)
		}
	}
	_base.Preemptall()

	// Any P entering _Pidle or _Psyscall from now on will observe
	// p.runSafePointFn == 1 and will call runSafePointFn when
	// changing its status to _Pidle/_Psyscall.

	// Run safe point function for all idle Ps. sched.pidle will
	// not change because we hold sched.lock.
	for p := _base.Sched.Pidle.Ptr(); p != nil; p = p.Link.Ptr() {
		if _base.Cas(&p.RunSafePointFn, 1, 0) {
			fn(p)
			_base.Sched.SafePointWait--
		}
	}

	wait := _base.Sched.SafePointWait > 0
	_base.Unlock(&_base.Sched.Lock)

	// Run fn for the current P.
	fn(_p_)

	// Force Ps currently in _Psyscall into _Pidle and hand them
	// off to induce safe point function execution.
	for i := 0; i < int(_base.Gomaxprocs); i++ {
		p := _base.Allp[i]
		s := p.Status
		if s == _base.Psyscall && p.RunSafePointFn == 1 && _base.Cas(&p.Status, s, _base.Pidle) {
			if _base.Trace.Enabled {
				_base.TraceGoSysBlock(p)
				_base.TraceProcStop(p)
			}
			p.Syscalltick++
			_base.Handoffp(p)
		}
	}

	// Wait for remaining Ps to run fn.
	if wait {
		for {
			// Wait for 100us, then try to re-preempt in
			// case of any races.
			if Notetsleep(&_base.Sched.SafePointNote, 100*1000) {
				_base.Noteclear(&_base.Sched.SafePointNote)
				break
			}
			_base.Preemptall()
		}
	}
	if _base.Sched.SafePointWait != 0 {
		_base.Throw("forEachP: not done")
	}
	for i := 0; i < int(_base.Gomaxprocs); i++ {
		p := _base.Allp[i]
		if p.RunSafePointFn != 0 {
			_base.Throw("forEachP: P did not run fn")
		}
	}

	_base.Lock(&_base.Sched.Lock)
	_base.Sched.SafePointFn = nil
	_base.Unlock(&_base.Sched.Lock)
	_base.Releasem(mp)
}

func GoschedImpl(gp *_base.G) {
	status := _base.Readgstatus(gp)
	if status&^_base.Gscan != _base.Grunning {
		_base.Dumpgstatus(gp)
		_base.Throw("bad g status")
	}
	_base.Casgstatus(gp, _base.Grunning, _base.Grunnable)
	_base.Dropg()
	_base.Lock(&_base.Sched.Lock)
	_base.Globrunqput(gp)
	_base.Unlock(&_base.Sched.Lock)

	_base.Schedule()
}

// Gosched continuation on g0.
func gosched_m(gp *_base.G) {
	if _base.Trace.Enabled {
		TraceGoSched()
	}
	GoschedImpl(gp)
}

// Purge all cached G's from gfree list to the global list.
func gfpurge(_p_ *_base.P) {
	_base.Lock(&_base.Sched.Gflock)
	for _p_.Gfreecnt != 0 {
		_p_.Gfreecnt--
		gp := _p_.Gfree
		_p_.Gfree = gp.Schedlink.Ptr()
		gp.Schedlink.Set(_base.Sched.Gfree)
		_base.Sched.Gfree = gp
		_base.Sched.Ngfree++
	}
	_base.Unlock(&_base.Sched.Gflock)
}

// Change number of processors.  The world is stopped, sched is locked.
// gcworkbufs are not being modified by either the GC or
// the write barrier code.
// Returns list of Ps with local work, they need to be scheduled by the caller.
func Procresize(nprocs int32) *_base.P {
	old := _base.Gomaxprocs
	if old < 0 || old > _base.MaxGomaxprocs || nprocs <= 0 || nprocs > _base.MaxGomaxprocs {
		_base.Throw("procresize: invalid arg")
	}
	if _base.Trace.Enabled {
		traceGomaxprocs(nprocs)
	}

	// update statistics
	now := _base.Nanotime()
	if _base.Sched.Procresizetime != 0 {
		_base.Sched.Totaltime += int64(old) * (now - _base.Sched.Procresizetime)
	}
	_base.Sched.Procresizetime = now

	// initialize new P's
	for i := int32(0); i < nprocs; i++ {
		pp := _base.Allp[i]
		if pp == nil {
			pp = new(_base.P)
			pp.Id = i
			pp.Status = _base.Pgcstop
			pp.Sudogcache = pp.Sudogbuf[:0]
			for i := range pp.Deferpool {
				pp.Deferpool[i] = pp.Deferpoolbuf[i][:0]
			}
			_base.Atomicstorep(unsafe.Pointer(&_base.Allp[i]), unsafe.Pointer(pp))
		}
		if pp.Mcache == nil {
			if old == 0 && i == 0 {
				if _base.Getg().M.Mcache == nil {
					_base.Throw("missing mcache?")
				}
				pp.Mcache = _base.Getg().M.Mcache // bootstrap
			} else {
				pp.Mcache = _base.Allocmcache()
			}
		}
	}

	// free unused P's
	for i := nprocs; i < old; i++ {
		p := _base.Allp[i]
		if _base.Trace.Enabled {
			if p == _base.Getg().M.P.Ptr() {
				// moving to p[0], pretend that we were descheduled
				// and then scheduled again to keep the trace sane.
				TraceGoSched()
				_base.TraceProcStop(p)
			}
		}
		// move all runnable goroutines to the global queue
		for p.Runqhead != p.Runqtail {
			// pop from tail of local queue
			p.Runqtail--
			gp := p.Runq[p.Runqtail%uint32(len(p.Runq))]
			// push onto head of global queue
			globrunqputhead(gp)
		}
		if p.Runnext != 0 {
			globrunqputhead(p.Runnext.Ptr())
			p.Runnext = 0
		}
		// if there's a background worker, make it runnable and put
		// it on the global queue so it can clean itself up
		if p.GcBgMarkWorker != nil {
			_base.Casgstatus(p.GcBgMarkWorker, _base.Gwaiting, _base.Grunnable)
			if _base.Trace.Enabled {
				_base.TraceGoUnpark(p.GcBgMarkWorker, 0)
			}
			_base.Globrunqput(p.GcBgMarkWorker)
			p.GcBgMarkWorker = nil
		}
		for i := range p.Sudogbuf {
			p.Sudogbuf[i] = nil
		}
		p.Sudogcache = p.Sudogbuf[:0]
		for i := range p.Deferpool {
			for j := range p.Deferpoolbuf[i] {
				p.Deferpoolbuf[i][j] = nil
			}
			p.Deferpool[i] = p.Deferpoolbuf[i][:0]
		}
		freemcache(p.Mcache)
		p.Mcache = nil
		gfpurge(p)
		traceProcFree(p)
		p.Status = _base.Pdead
		// can't free P itself because it can be referenced by an M in syscall
	}

	_g_ := _base.Getg()
	if _g_.M.P != 0 && _g_.M.P.Ptr().Id < nprocs {
		// continue to use the current P
		_g_.M.P.Ptr().Status = _base.Prunning
	} else {
		// release the current P and acquire allp[0]
		if _g_.M.P != 0 {
			_g_.M.P.Ptr().M = 0
		}
		_g_.M.P = 0
		_g_.M.Mcache = nil
		p := _base.Allp[0]
		p.M = 0
		p.Status = _base.Pidle
		_base.Acquirep(p)
		if _base.Trace.Enabled {
			_base.TraceGoStart()
		}
	}
	var runnablePs *_base.P
	for i := nprocs - 1; i >= 0; i-- {
		p := _base.Allp[i]
		if _g_.M.P.Ptr() == p {
			continue
		}
		p.Status = _base.Pidle
		if _base.Runqempty(p) {
			_base.Pidleput(p)
		} else {
			p.M.Set(_base.Mget())
			p.Link.Set(runnablePs)
			runnablePs = p
		}
	}
	var int32p *int32 = &_base.Gomaxprocs // make compiler check that gomaxprocs is an int32
	_base.Atomicstore((*uint32)(unsafe.Pointer(int32p)), uint32(nprocs))
	return runnablePs
}

// Put gp at the head of the global runnable queue.
// Sched must be locked.
// May run during STW, so write barriers are not allowed.
//go:nowritebarrier
func globrunqputhead(gp *_base.G) {
	gp.Schedlink = _base.Sched.Runqhead
	_base.Sched.Runqhead.Set(gp)
	if _base.Sched.Runqtail == 0 {
		_base.Sched.Runqtail.Set(gp)
	}
	_base.Sched.Runqsize++
}
