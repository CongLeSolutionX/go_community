// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sched

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	"unsafe"
)

func Dumpgstatus(gp *_core.G) {
	_g_ := _core.Getg()
	print("runtime: gp: gp=", gp, ", goid=", gp.Goid, ", gp->atomicstatus=", _lock.Readgstatus(gp), "\n")
	print("runtime:  g:  g=", _g_, ", goid=", _g_.Goid, ",  g->atomicstatus=", _lock.Readgstatus(_g_), "\n")
}

func Checkmcount() {
	// sched lock is held
	if _core.Sched.Mcount > _core.Sched.Maxmcount {
		print("runtime: program exceeds ", _core.Sched.Maxmcount, "-thread limit\n")
		_lock.Gothrow("thread exhaustion")
	}
}

func Mcommoninit(mp *_core.M) {
	_g_ := _core.Getg()

	// g0 stack won't make sense for user (and is not necessary unwindable).
	if _g_ != _g_.M.G0 {
		Callers(1, &mp.Createstack[0], len(mp.Createstack))
	}

	mp.Fastrand = 0x49f6428a + uint32(mp.Id) + uint32(Cputicks())
	if mp.Fastrand == 0 {
		mp.Fastrand = 0x49f6428a
	}

	_lock.Lock(&_core.Sched.Lock)
	mp.Id = _core.Sched.Mcount
	_core.Sched.Mcount++
	Checkmcount()
	mpreinit(mp)
	if mp.Gsignal != nil {
		mp.Gsignal.Stackguard1 = mp.Gsignal.Stack.Lo + _core.StackGuard
	}

	// Add to allm so garbage collector doesn't free g->m
	// when it is just in a register or thread-local storage.
	mp.Alllink = _lock.Allm

	// NumCgoCall() iterates over allm w/o schedlock,
	// so we need to publish it safely.
	Atomicstorep(unsafe.Pointer(&_lock.Allm), unsafe.Pointer(mp))
	_lock.Unlock(&_core.Sched.Lock)
}

// Mark gp ready to run.
func Ready(gp *_core.G) {
	status := _lock.Readgstatus(gp)

	// Mark runnable.
	_g_ := _core.Getg()
	_g_.M.Locks++ // disable preemption because it can be holding p in a local var
	if status&^_lock.Gscan != _lock.Gwaiting {
		Dumpgstatus(gp)
		_lock.Gothrow("bad g->status in ready")
	}

	// status is Gwaiting or Gscanwaiting, make Grunnable and put on runq
	Casgstatus(gp, _lock.Gwaiting, _lock.Grunnable)
	Runqput(_g_.M.P, gp)
	if _lock.Atomicload(&_core.Sched.Npidle) != 0 && _lock.Atomicload(&_core.Sched.Nmspinning) == 0 { // TODO: fast atomic
		Wakep()
	}
	_g_.M.Locks--
	if _g_.M.Locks == 0 && _g_.Preempt { // restore the preemption request in case we've cleared it in newstack
		_g_.Stackguard0 = _lock.StackPreempt
	}
}

// If asked to move to or from a Gscanstatus this will throw. Use the castogscanstatus
// and casfrom_Gscanstatus instead.
// casgstatus will loop if the g->atomicstatus is in a Gscan status until the routine that
// put it in the Gscan state is finished.
//go:nosplit
func Casgstatus(gp *_core.G, oldval, newval uint32) {
	if (oldval&_lock.Gscan != 0) || (newval&_lock.Gscan != 0) || oldval == newval {
		_lock.Systemstack(func() {
			print("casgstatus: oldval=", _core.Hex(oldval), " newval=", _core.Hex(newval), "\n")
			_lock.Gothrow("casgstatus: bad incoming values")
		})
	}

	// loop if gp->atomicstatus is in a scan state giving
	// GC time to finish and change the state to oldval.
	for !Cas(&gp.Atomicstatus, oldval, newval) {
		if oldval == _lock.Gwaiting && gp.Atomicstatus == _lock.Grunnable {
			_lock.Systemstack(func() {
				_lock.Gothrow("casgstatus: waiting for Gwaiting but is Grunnable")
			})
		}
		// Help GC if needed.
		// if gp.preemptscan && !gp.gcworkdone && (oldval == _Grunning || oldval == _Gsyscall) {
		// 	gp.preemptscan = false
		// 	systemstack(func() {
		// 		gcphasework(gp)
		// 	})
		// }
	}
}

// Called to start an M.
//go:nosplit
func Mstart() {
	_g_ := _core.Getg()

	if _g_.Stack.Lo == 0 {
		// Initialize stack bounds from system stack.
		// Cgo may have left stack size in stack.hi.
		size := _g_.Stack.Hi
		if size == 0 {
			size = 8192
		}
		_g_.Stack.Hi = uintptr(_core.Noescape(unsafe.Pointer(&size)))
		_g_.Stack.Lo = _g_.Stack.Hi - size + 1024
	}
	// Initialize stack guards so that we can start calling
	// both Go and C functions with stack growth prologues.
	_g_.Stackguard0 = _g_.Stack.Lo + _core.StackGuard
	_g_.Stackguard1 = _g_.Stackguard0
	mstart1()
}

func mstart1() {
	_g_ := _core.Getg()

	if _g_ != _g_.M.G0 {
		_lock.Gothrow("bad runtime·mstart")
	}

	// Record top of stack for use by mcall.
	// Once we call schedule we're never coming back,
	// so other calls can reuse this stack space.
	gosave(&_g_.M.G0.Sched)
	_g_.M.G0.Sched.Pc = ^uintptr(0) // make sure it is never used
	_core.Asminit()
	_core.Minit()

	// Install signal handlers; after minit so that minit can
	// prepare the thread to be able to handle the signals.
	if _g_.M == &_core.M0 {
		initsig()
	}

	if _g_.M.Mstartfn != nil {
		fn := *(*func())(unsafe.Pointer(&_g_.M.Mstartfn))
		fn()
	}

	if _g_.M.Helpgc != 0 {
		_g_.M.Helpgc = 0
		stopm()
	} else if _g_.M != &_core.M0 {
		Acquirep(_g_.M.Nextp)
		_g_.M.Nextp = nil
	}
	Schedule()

	// TODO(brainman): This point is never reached, because scheduler
	// does not release os threads at the moment. But once this path
	// is enabled, we must remove our seh here.
}

type cgothreadstart struct {
	g   *_core.G
	tls *uint64
	fn  unsafe.Pointer
}

// Allocate a new m unassociated with any thread.
// Can use p for allocation context if needed.
func Allocm(_p_ *_core.P) *_core.M {
	_g_ := _core.Getg()
	_g_.M.Locks++ // disable GC because it can be called from sysmon
	if _g_.M.P == nil {
		Acquirep(_p_) // temporarily borrow p for mallocs in this function
	}
	mp := newM()
	Mcommoninit(mp)

	// In case of cgo or Solaris, pthread_create will make us a stack.
	// Windows and Plan 9 will layout sched stack on OS stack.
	if Iscgo || _lock.GOOS == "solaris" || _lock.GOOS == "windows" || _lock.GOOS == "plan9" {
		mp.G0 = Malg(-1)
	} else {
		mp.G0 = Malg(8192)
	}
	mp.G0.M = mp

	if _p_ == _g_.M.P {
		releasep()
	}
	_g_.M.Locks--
	if _g_.M.Locks == 0 && _g_.Preempt { // restore the preemption request in case we've cleared it in newstack
		_g_.Stackguard0 = _lock.StackPreempt
	}

	return mp
}

func allocg() *_core.G {
	return newG()
}

// Create a new m.  It will start off with a call to fn, or else the scheduler.
func Newm(fn func(), _p_ *_core.P) {
	mp := Allocm(_p_)
	mp.Nextp = _p_
	mp.Mstartfn = *(*unsafe.Pointer)(unsafe.Pointer(&fn))

	if Iscgo {
		var ts cgothreadstart
		if Cgo_thread_start == nil {
			_lock.Gothrow("_cgo_thread_start missing")
		}
		ts.g = mp.G0
		ts.tls = (*uint64)(unsafe.Pointer(&mp.Tls[0]))
		ts.fn = unsafe.Pointer(_lock.FuncPC(Mstart))
		Asmcgocall(Cgo_thread_start, unsafe.Pointer(&ts))
		return
	}
	newosproc(mp, unsafe.Pointer(mp.G0.Stack.Hi))
}

// Stops execution of the current m until new work is available.
// Returns with acquired P.
func stopm() {
	_g_ := _core.Getg()

	if _g_.M.Locks != 0 {
		_lock.Gothrow("stopm holding locks")
	}
	if _g_.M.P != nil {
		_lock.Gothrow("stopm holding p")
	}
	if _g_.M.Spinning {
		_g_.M.Spinning = false
		_lock.Xadd(&_core.Sched.Nmspinning, -1)
	}

retry:
	_lock.Lock(&_core.Sched.Lock)
	mput(_g_.M)
	_lock.Unlock(&_core.Sched.Lock)
	Notesleep(&_g_.M.Park)
	Noteclear(&_g_.M.Park)
	if _g_.M.Helpgc != 0 {
		gchelper()
		_g_.M.Helpgc = 0
		_g_.M.Mcache = nil
		goto retry
	}
	Acquirep(_g_.M.Nextp)
	_g_.M.Nextp = nil
}

func mspinning() {
	_core.Getg().M.Spinning = true
}

// Schedules some M to run the p (creates an M if necessary).
// If p==nil, tries to get an idle P, if no idle P's does nothing.
func startm(_p_ *_core.P, spinning bool) {
	_lock.Lock(&_core.Sched.Lock)
	if _p_ == nil {
		_p_ = Pidleget()
		if _p_ == nil {
			_lock.Unlock(&_core.Sched.Lock)
			if spinning {
				_lock.Xadd(&_core.Sched.Nmspinning, -1)
			}
			return
		}
	}
	mp := Mget()
	_lock.Unlock(&_core.Sched.Lock)
	if mp == nil {
		var fn func()
		if spinning {
			fn = mspinning
		}
		Newm(fn, _p_)
		return
	}
	if mp.Spinning {
		_lock.Gothrow("startm: m is spinning")
	}
	if mp.Nextp != nil {
		_lock.Gothrow("startm: m has p")
	}
	mp.Spinning = spinning
	mp.Nextp = _p_
	Notewakeup(&mp.Park)
}

// Hands off P from syscall or locked M.
func Handoffp(_p_ *_core.P) {
	// if it has local work, start it straight away
	if _p_.Runqhead != _p_.Runqtail || _core.Sched.Runqsize != 0 {
		startm(_p_, false)
		return
	}
	// no local work, check that there are no spinning/idle M's,
	// otherwise our help is not required
	if _lock.Atomicload(&_core.Sched.Nmspinning)+_lock.Atomicload(&_core.Sched.Npidle) == 0 && Cas(&_core.Sched.Nmspinning, 0, 1) { // TODO: fast atomic
		startm(_p_, true)
		return
	}
	_lock.Lock(&_core.Sched.Lock)
	if _core.Sched.Gcwaiting != 0 {
		_p_.Status = _lock.Pgcstop
		_core.Sched.Stopwait--
		if _core.Sched.Stopwait == 0 {
			Notewakeup(&_core.Sched.Stopnote)
		}
		_lock.Unlock(&_core.Sched.Lock)
		return
	}
	if _core.Sched.Runqsize != 0 {
		_lock.Unlock(&_core.Sched.Lock)
		startm(_p_, false)
		return
	}
	// If this is the last running P and nobody is polling network,
	// need to wakeup another M to poll network.
	if _core.Sched.Npidle == uint32(_lock.Gomaxprocs-1) && Atomicload64(&_core.Sched.Lastpoll) != 0 {
		_lock.Unlock(&_core.Sched.Lock)
		startm(_p_, false)
		return
	}
	Pidleput(_p_)
	_lock.Unlock(&_core.Sched.Lock)
}

// Tries to add one more P to execute G's.
// Called when a G is made runnable (newproc, ready).
func Wakep() {
	// be conservative about spinning threads
	if !Cas(&_core.Sched.Nmspinning, 0, 1) {
		return
	}
	startm(nil, true)
}

// Stops execution of the current m that is locked to a g until the g is runnable again.
// Returns with acquired P.
func stoplockedm() {
	_g_ := _core.Getg()

	if _g_.M.Lockedg == nil || _g_.M.Lockedg.Lockedm != _g_.M {
		_lock.Gothrow("stoplockedm: inconsistent locking")
	}
	if _g_.M.P != nil {
		// Schedule another M to run this p.
		_p_ := releasep()
		Handoffp(_p_)
	}
	Incidlelocked(1)
	// Wait until another thread schedules lockedg again.
	Notesleep(&_g_.M.Park)
	Noteclear(&_g_.M.Park)
	status := _lock.Readgstatus(_g_.M.Lockedg)
	if status&^_lock.Gscan != _lock.Grunnable {
		print("runtime:stoplockedm: g is not Grunnable or Gscanrunnable\n")
		Dumpgstatus(_g_)
		_lock.Gothrow("stoplockedm: not runnable")
	}
	Acquirep(_g_.M.Nextp)
	_g_.M.Nextp = nil
}

// Schedules the locked m to run the locked gp.
func startlockedm(gp *_core.G) {
	_g_ := _core.Getg()

	mp := gp.Lockedm
	if mp == _g_.M {
		_lock.Gothrow("startlockedm: locked to me")
	}
	if mp.Nextp != nil {
		_lock.Gothrow("startlockedm: m has p")
	}
	// directly handoff current P to the locked m
	Incidlelocked(-1)
	_p_ := releasep()
	mp.Nextp = _p_
	Notewakeup(&mp.Park)
	stopm()
}

// Stops the current m for stoptheworld.
// Returns when the world is restarted.
func gcstopm() {
	_g_ := _core.Getg()

	if _core.Sched.Gcwaiting == 0 {
		_lock.Gothrow("gcstopm: not waiting for gc")
	}
	if _g_.M.Spinning {
		_g_.M.Spinning = false
		_lock.Xadd(&_core.Sched.Nmspinning, -1)
	}
	_p_ := releasep()
	_lock.Lock(&_core.Sched.Lock)
	_p_.Status = _lock.Pgcstop
	_core.Sched.Stopwait--
	if _core.Sched.Stopwait == 0 {
		Notewakeup(&_core.Sched.Stopnote)
	}
	_lock.Unlock(&_core.Sched.Lock)
	stopm()
}

// Schedules gp to run on the current M.
// Never returns.
func execute(gp *_core.G) {
	_g_ := _core.Getg()

	Casgstatus(gp, _lock.Grunnable, _lock.Grunning)
	gp.Waitsince = 0
	gp.Preempt = false
	gp.Stackguard0 = gp.Stack.Lo + _core.StackGuard
	_g_.M.P.Schedtick++
	_g_.M.Curg = gp
	gp.M = _g_.M

	// Check whether the profiler needs to be turned on or off.
	hz := _core.Sched.Profilehz
	if _g_.M.Profilehz != hz {
		Resetcpuprofiler(hz)
	}

	Gogo(&gp.Sched)
}

// Finds a runnable goroutine to execute.
// Tries to steal from other P's, get g from global queue, poll network.
func findrunnable() *_core.G {
	_g_ := _core.Getg()

top:
	if _core.Sched.Gcwaiting != 0 {
		gcstopm()
		goto top
	}
	if Fingwait && Fingwake {
		if gp := wakefing(); gp != nil {
			Ready(gp)
		}
	}

	// local runq
	if gp := Runqget(_g_.M.P); gp != nil {
		return gp
	}

	// global runq
	if _core.Sched.Runqsize != 0 {
		_lock.Lock(&_core.Sched.Lock)
		gp := globrunqget(_g_.M.P, 0)
		_lock.Unlock(&_core.Sched.Lock)
		if gp != nil {
			return gp
		}
	}

	// poll network - returns list of goroutines
	if gp := Netpoll(false); gp != nil { // non-blocking
		Injectglist(gp.Schedlink)
		Casgstatus(gp, _lock.Gwaiting, _lock.Grunnable)
		return gp
	}

	// If number of spinning M's >= number of busy P's, block.
	// This is necessary to prevent excessive CPU consumption
	// when GOMAXPROCS>>1 but the program parallelism is low.
	if !_g_.M.Spinning && 2*_lock.Atomicload(&_core.Sched.Nmspinning) >= uint32(_lock.Gomaxprocs)-_lock.Atomicload(&_core.Sched.Npidle) { // TODO: fast atomic
		goto stop
	}
	if !_g_.M.Spinning {
		_g_.M.Spinning = true
		_lock.Xadd(&_core.Sched.Nmspinning, 1)
	}
	// random steal from other P's
	for i := 0; i < int(2*_lock.Gomaxprocs); i++ {
		if _core.Sched.Gcwaiting != 0 {
			goto top
		}
		_p_ := _lock.Allp[_lock.Fastrand1()%uint32(_lock.Gomaxprocs)]
		var gp *_core.G
		if _p_ == _g_.M.P {
			gp = Runqget(_p_)
		} else {
			gp = Runqsteal(_g_.M.P, _p_)
		}
		if gp != nil {
			return gp
		}
	}
stop:

	// return P and block
	_lock.Lock(&_core.Sched.Lock)
	if _core.Sched.Gcwaiting != 0 {
		_lock.Unlock(&_core.Sched.Lock)
		goto top
	}
	if _core.Sched.Runqsize != 0 {
		gp := globrunqget(_g_.M.P, 0)
		_lock.Unlock(&_core.Sched.Lock)
		return gp
	}
	_p_ := releasep()
	Pidleput(_p_)
	_lock.Unlock(&_core.Sched.Lock)
	if _g_.M.Spinning {
		_g_.M.Spinning = false
		_lock.Xadd(&_core.Sched.Nmspinning, -1)
	}

	// check all runqueues once again
	for i := 0; i < int(_lock.Gomaxprocs); i++ {
		_p_ := _lock.Allp[i]
		if _p_ != nil && _p_.Runqhead != _p_.Runqtail {
			_lock.Lock(&_core.Sched.Lock)
			_p_ = Pidleget()
			_lock.Unlock(&_core.Sched.Lock)
			if _p_ != nil {
				Acquirep(_p_)
				goto top
			}
			break
		}
	}

	// poll network
	if Xchg64(&_core.Sched.Lastpoll, 0) != 0 {
		if _g_.M.P != nil {
			_lock.Gothrow("findrunnable: netpoll with p")
		}
		if _g_.M.Spinning {
			_lock.Gothrow("findrunnable: netpoll with spinning")
		}
		gp := Netpoll(true) // block until new work is available
		Atomicstore64(&_core.Sched.Lastpoll, uint64(_lock.Nanotime()))
		if gp != nil {
			_lock.Lock(&_core.Sched.Lock)
			_p_ = Pidleget()
			_lock.Unlock(&_core.Sched.Lock)
			if _p_ != nil {
				Acquirep(_p_)
				Injectglist(gp.Schedlink)
				Casgstatus(gp, _lock.Gwaiting, _lock.Grunnable)
				return gp
			}
			Injectglist(gp)
		}
	}
	stopm()
	goto top
}

func resetspinning() {
	_g_ := _core.Getg()

	var nmspinning uint32
	if _g_.M.Spinning {
		_g_.M.Spinning = false
		nmspinning = _lock.Xadd(&_core.Sched.Nmspinning, -1)
		if nmspinning < 0 {
			_lock.Gothrow("findrunnable: negative nmspinning")
		}
	} else {
		nmspinning = _lock.Atomicload(&_core.Sched.Nmspinning)
	}

	// M wakeup policy is deliberately somewhat conservative (see nmspinning handling),
	// so see if we need to wakeup another P here.
	if nmspinning == 0 && _lock.Atomicload(&_core.Sched.Npidle) > 0 {
		Wakep()
	}
}

// Injects the list of runnable G's into the scheduler.
// Can run concurrently with GC.
func Injectglist(glist *_core.G) {
	if glist == nil {
		return
	}
	_lock.Lock(&_core.Sched.Lock)
	var n int
	for n = 0; glist != nil; n++ {
		gp := glist
		glist = gp.Schedlink
		Casgstatus(gp, _lock.Gwaiting, _lock.Grunnable)
		globrunqput(gp)
	}
	_lock.Unlock(&_core.Sched.Lock)
	for ; n != 0 && _core.Sched.Npidle != 0; n-- {
		startm(nil, false)
	}
}

// One round of scheduler: find a runnable goroutine and execute it.
// Never returns.
func Schedule() {
	_g_ := _core.Getg()

	if _g_.M.Locks != 0 {
		_lock.Gothrow("schedule: holding locks")
	}

	if _g_.M.Lockedg != nil {
		stoplockedm()
		execute(_g_.M.Lockedg) // Never returns.
	}

top:
	if _core.Sched.Gcwaiting != 0 {
		gcstopm()
		goto top
	}

	var gp *_core.G
	// Check the global runnable queue once in a while to ensure fairness.
	// Otherwise two goroutines can completely occupy the local runqueue
	// by constantly respawning each other.
	tick := _g_.M.P.Schedtick
	// This is a fancy way to say tick%61==0,
	// it uses 2 MUL instructions instead of a single DIV and so is faster on modern processors.
	if uint64(tick)-((uint64(tick)*0x4325c53f)>>36)*61 == 0 && _core.Sched.Runqsize > 0 {
		_lock.Lock(&_core.Sched.Lock)
		gp = globrunqget(_g_.M.P, 1)
		_lock.Unlock(&_core.Sched.Lock)
		if gp != nil {
			resetspinning()
		}
	}
	if gp == nil {
		gp = Runqget(_g_.M.P)
		if gp != nil && _g_.M.Spinning {
			_lock.Gothrow("schedule: spinning with local work")
		}
	}
	if gp == nil {
		gp = findrunnable() // blocks until work is available
		resetspinning()
	}

	if gp.Lockedm != nil {
		// Hands off own p to the locked m,
		// then blocks waiting for a new p.
		startlockedm(gp)
		goto top
	}

	execute(gp)
}

// dropg removes the association between m and the current goroutine m->curg (gp for short).
// Typically a caller sets gp's status away from Grunning and then
// immediately calls dropg to finish the job. The caller is also responsible
// for arranging that gp will be restarted using ready at an
// appropriate time. After calling dropg and arranging for gp to be
// readied later, the caller can do other work but eventually should
// call schedule to restart the scheduling of goroutines on this m.
func Dropg() {
	_g_ := _core.Getg()

	if _g_.M.Lockedg == nil {
		_g_.M.Curg.M = nil
		_g_.M.Curg = nil
	}
}

func Parkunlock_c(gp *_core.G, lock unsafe.Pointer) bool {
	_lock.Unlock((*_core.Mutex)(lock))
	return true
}

// park continuation on g0.
func Park_m(gp *_core.G) {
	_g_ := _core.Getg()

	Casgstatus(gp, _lock.Grunning, _lock.Gwaiting)
	Dropg()

	if _g_.M.Waitunlockf != nil {
		fn := *(*func(*_core.G, unsafe.Pointer) bool)(unsafe.Pointer(&_g_.M.Waitunlockf))
		ok := fn(gp, _g_.M.Waitlock)
		_g_.M.Waitunlockf = nil
		_g_.M.Waitlock = nil
		if !ok {
			Casgstatus(gp, _lock.Gwaiting, _lock.Grunnable)
			execute(gp) // Schedule it back, never returns.
		}
	}
	Schedule()
}

// Gosched continuation on g0.
func Gosched_m(gp *_core.G) {
	status := _lock.Readgstatus(gp)
	if status&^_lock.Gscan != _lock.Grunning {
		Dumpgstatus(gp)
		_lock.Gothrow("bad g status")
	}
	Casgstatus(gp, _lock.Grunning, _lock.Grunnable)
	Dropg()
	_lock.Lock(&_core.Sched.Lock)
	globrunqput(gp)
	_lock.Unlock(&_core.Sched.Lock)

	Schedule()
}

//go:nosplit
//go:nowritebarrier
func Save(pc, sp uintptr) {
	_g_ := _core.Getg()

	_g_.Sched.Pc = pc
	_g_.Sched.Sp = sp
	_g_.Sched.Lr = 0
	_g_.Sched.Ret = 0
	_g_.Sched.Ctxt = nil
	// _g_.sched.g = _g_, but avoid write barrier, which smashes _g_.sched
	*(*uintptr)(unsafe.Pointer(&_g_.Sched.G)) = uintptr(unsafe.Pointer(_g_))
}

// The same as entersyscall(), but with a hint that the syscall is blocking.
//go:nosplit
func entersyscallblock(dummy int32) {
	_g_ := _core.Getg()

	_g_.M.Locks++ // see comment in entersyscall
	_g_.Throwsplit = true
	_g_.Stackguard0 = _lock.StackPreempt // see comment in entersyscall

	// Leave SP around for GC and traceback.
	pc := _lock.Getcallerpc(unsafe.Pointer(&dummy))
	sp := _lock.Getcallersp(unsafe.Pointer(&dummy))
	Save(pc, sp)
	_g_.Syscallsp = _g_.Sched.Sp
	_g_.Syscallpc = _g_.Sched.Pc
	if _g_.Syscallsp < _g_.Stack.Lo || _g_.Stack.Hi < _g_.Syscallsp {
		sp1 := sp
		sp2 := _g_.Sched.Sp
		sp3 := _g_.Syscallsp
		_lock.Systemstack(func() {
			print("entersyscallblock inconsistent ", _core.Hex(sp1), " ", _core.Hex(sp2), " ", _core.Hex(sp3), " [", _core.Hex(_g_.Stack.Lo), ",", _core.Hex(_g_.Stack.Hi), "]\n")
			_lock.Gothrow("entersyscallblock")
		})
	}
	Casgstatus(_g_, _lock.Grunning, _lock.Gsyscall)
	if _g_.Syscallsp < _g_.Stack.Lo || _g_.Stack.Hi < _g_.Syscallsp {
		_lock.Systemstack(func() {
			print("entersyscallblock inconsistent ", _core.Hex(sp), " ", _core.Hex(_g_.Sched.Sp), " ", _core.Hex(_g_.Syscallsp), " [", _core.Hex(_g_.Stack.Lo), ",", _core.Hex(_g_.Stack.Hi), "]\n")
			_lock.Gothrow("entersyscallblock")
		})
	}

	_lock.Systemstack(entersyscallblock_handoff)

	// Resave for traceback during blocked call.
	Save(_lock.Getcallerpc(unsafe.Pointer(&dummy)), _lock.Getcallersp(unsafe.Pointer(&dummy)))

	_g_.M.Locks--
}

func entersyscallblock_handoff() {
	Handoffp(releasep())
}

// The goroutine g exited its system call.
// Arrange for it to run on a cpu again.
// This is called only from the go syscall library, not
// from the low-level system calls used by the
//go:nosplit
func Exitsyscall(dummy int32) {
	_g_ := _core.Getg()

	_g_.M.Locks++ // see comment in entersyscall
	if _lock.Getcallersp(unsafe.Pointer(&dummy)) > _g_.Syscallsp {
		_lock.Gothrow("exitsyscall: syscall frame is no longer valid")
	}

	_g_.Waitsince = 0
	if exitsyscallfast() {
		if _g_.M.Mcache == nil {
			_lock.Gothrow("lost mcache")
		}
		// There's a cpu for us, so we can run.
		_g_.M.P.Syscalltick++
		// We need to cas the status and scan before resuming...
		Casgstatus(_g_, _lock.Gsyscall, _lock.Grunning)

		// Garbage collector isn't running (since we are),
		// so okay to clear syscallsp.
		_g_.Syscallsp = 0
		_g_.M.Locks--
		if _g_.Preempt {
			// restore the preemption request in case we've cleared it in newstack
			_g_.Stackguard0 = _lock.StackPreempt
		} else {
			// otherwise restore the real _StackGuard, we've spoiled it in entersyscall/entersyscallblock
			_g_.Stackguard0 = _g_.Stack.Lo + _core.StackGuard
		}
		_g_.Throwsplit = false
		return
	}

	_g_.M.Locks--

	// Call the scheduler.
	Mcall(exitsyscall0)

	if _g_.M.Mcache == nil {
		_lock.Gothrow("lost mcache")
	}

	// Scheduler returned, so we're allowed to run now.
	// Delete the syscallsp information that we left for
	// the garbage collector during the system call.
	// Must wait until now because until gosched returns
	// we don't know for sure that the garbage collector
	// is not running.
	_g_.Syscallsp = 0
	_g_.M.P.Syscalltick++
	_g_.Throwsplit = false
}

//go:nosplit
func exitsyscallfast() bool {
	_g_ := _core.Getg()

	// Freezetheworld sets stopwait but does not retake P's.
	if _core.Sched.Stopwait != 0 {
		_g_.M.Mcache = nil
		_g_.M.P = nil
		return false
	}

	// Try to re-acquire the last P.
	if _g_.M.P != nil && _g_.M.P.Status == _lock.Psyscall && Cas(&_g_.M.P.Status, _lock.Psyscall, _lock.Prunning) {
		// There's a cpu for us, so we can run.
		_g_.M.Mcache = _g_.M.P.Mcache
		_g_.M.P.M = _g_.M
		return true
	}

	// Try to get any other idle P.
	_g_.M.Mcache = nil
	_g_.M.P = nil
	if _core.Sched.Pidle != nil {
		var ok bool
		_lock.Systemstack(func() {
			ok = exitsyscallfast_pidle()
		})
		if ok {
			return true
		}
	}
	return false
}

func exitsyscallfast_pidle() bool {
	_lock.Lock(&_core.Sched.Lock)
	_p_ := Pidleget()
	if _p_ != nil && _lock.Atomicload(&_core.Sched.Sysmonwait) != 0 {
		_lock.Atomicstore(&_core.Sched.Sysmonwait, 0)
		Notewakeup(&_core.Sched.Sysmonnote)
	}
	_lock.Unlock(&_core.Sched.Lock)
	if _p_ != nil {
		Acquirep(_p_)
		return true
	}
	return false
}

// exitsyscall slow path on g0.
// Failed to acquire P, enqueue gp as runnable.
func exitsyscall0(gp *_core.G) {
	_g_ := _core.Getg()

	Casgstatus(gp, _lock.Gsyscall, _lock.Grunnable)
	Dropg()
	_lock.Lock(&_core.Sched.Lock)
	_p_ := Pidleget()
	if _p_ == nil {
		globrunqput(gp)
	} else if _lock.Atomicload(&_core.Sched.Sysmonwait) != 0 {
		_lock.Atomicstore(&_core.Sched.Sysmonwait, 0)
		Notewakeup(&_core.Sched.Sysmonnote)
	}
	_lock.Unlock(&_core.Sched.Lock)
	if _p_ != nil {
		Acquirep(_p_)
		execute(gp) // Never returns.
	}
	if _g_.M.Lockedg != nil {
		// Wait until another thread schedules gp and so m again.
		stoplockedm()
		execute(gp) // Never returns.
	}
	stopm()
	Schedule() // Never returns.
}

// Allocate a new g, with a stack big enough for stacksize bytes.
func Malg(stacksize int32) *_core.G {
	newg := allocg()
	if stacksize >= 0 {
		stacksize = Round2(_core.StackSystem + stacksize)
		_lock.Systemstack(func() {
			newg.Stack = Stackalloc(uint32(stacksize))
		})
		newg.Stackguard0 = newg.Stack.Lo + _core.StackGuard
		newg.Stackguard1 = ^uintptr(0)
	}
	return newg
}

func mcount() int32 {
	return _core.Sched.Mcount
}

var Prof struct {
	Lock uint32
	Hz   int32
}

func _System()       { _System() }
func _ExternalCode() { _ExternalCode() }
func _GC()           { _GC() }

var etext struct{}

// Called if we receive a SIGPROF signal.
func sigprof(pc *uint8, sp *uint8, lr *uint8, gp *_core.G, mp *_core.M) {
	var n int32
	var traceback bool
	var stk [100]uintptr

	if Prof.Hz == 0 {
		return
	}

	// Profiling runs concurrently with GC, so it must not allocate.
	mp.Mallocing++

	// Define that a "user g" is a user-created goroutine, and a "system g"
	// is one that is m->g0 or m->gsignal. We've only made sure that we
	// can unwind user g's, so exclude the system g's.
	//
	// It is not quite as easy as testing gp == m->curg (the current user g)
	// because we might be interrupted for profiling halfway through a
	// goroutine switch. The switch involves updating three (or four) values:
	// g, PC, SP, and (on arm) LR. The PC must be the last to be updated,
	// because once it gets updated the new g is running.
	//
	// When switching from a user g to a system g, LR is not considered live,
	// so the update only affects g, SP, and PC. Since PC must be last, there
	// the possible partial transitions in ordinary execution are (1) g alone is updated,
	// (2) both g and SP are updated, and (3) SP alone is updated.
	// If g is updated, we'll see a system g and not look closer.
	// If SP alone is updated, we can detect the partial transition by checking
	// whether the SP is within g's stack bounds. (We could also require that SP
	// be changed only after g, but the stack bounds check is needed by other
	// cases, so there is no need to impose an additional requirement.)
	//
	// There is one exceptional transition to a system g, not in ordinary execution.
	// When a signal arrives, the operating system starts the signal handler running
	// with an updated PC and SP. The g is updated last, at the beginning of the
	// handler. There are two reasons this is okay. First, until g is updated the
	// g and SP do not match, so the stack bounds check detects the partial transition.
	// Second, signal handlers currently run with signals disabled, so a profiling
	// signal cannot arrive during the handler.
	//
	// When switching from a system g to a user g, there are three possibilities.
	//
	// First, it may be that the g switch has no PC update, because the SP
	// either corresponds to a user g throughout (as in asmcgocall)
	// or because it has been arranged to look like a user g frame
	// (as in cgocallback_gofunc). In this case, since the entire
	// transition is a g+SP update, a partial transition updating just one of
	// those will be detected by the stack bounds check.
	//
	// Second, when returning from a signal handler, the PC and SP updates
	// are performed by the operating system in an atomic update, so the g
	// update must be done before them. The stack bounds check detects
	// the partial transition here, and (again) signal handlers run with signals
	// disabled, so a profiling signal cannot arrive then anyway.
	//
	// Third, the common case: it may be that the switch updates g, SP, and PC
	// separately, as in gogo.
	//
	// Because gogo is the only instance, we check whether the PC lies
	// within that function, and if so, not ask for a traceback. This approach
	// requires knowing the size of the gogo function, which we
	// record in arch_*.h and check in runtime_test.go.
	//
	// There is another apparently viable approach, recorded here in case
	// the "PC within gogo" check turns out not to be usable.
	// It would be possible to delay the update of either g or SP until immediately
	// before the PC update instruction. Then, because of the stack bounds check,
	// the only problematic interrupt point is just before that PC update instruction,
	// and the sigprof handler can detect that instruction and simulate stepping past
	// it in order to reach a consistent state. On ARM, the update of g must be made
	// in two places (in R10 and also in a TLS slot), so the delayed update would
	// need to be the SP update. The sigprof handler must read the instruction at
	// the current PC and if it was the known instruction (for example, JMP BX or
	// MOV R2, PC), use that other register in place of the PC value.
	// The biggest drawback to this solution is that it requires that we can tell
	// whether it's safe to read from the memory pointed at by PC.
	// In a correct program, we can test PC == nil and otherwise read,
	// but if a profiling signal happens at the instant that a program executes
	// a bad jump (before the program manages to handle the resulting fault)
	// the profiling handler could fault trying to read nonexistent memory.
	//
	// To recap, there are no constraints on the assembly being used for the
	// transition. We simply require that g and SP match and that the PC is not
	// in gogo.
	traceback = true
	usp := uintptr(unsafe.Pointer(sp))
	gogo := _lock.FuncPC(Gogo)
	if gp == nil || gp != mp.Curg ||
		usp < gp.Stack.Lo || gp.Stack.Hi < usp ||
		(gogo <= uintptr(unsafe.Pointer(pc)) && uintptr(unsafe.Pointer(pc)) < gogo+_lock.RuntimeGogoBytes) {
		traceback = false
	}

	n = 0
	if traceback {
		n = int32(_lock.Gentraceback(uintptr(unsafe.Pointer(pc)), uintptr(unsafe.Pointer(sp)), uintptr(unsafe.Pointer(lr)), gp, 0, &stk[0], len(stk), nil, nil, _lock.TraceTrap))
	}
	if !traceback || n <= 0 {
		// Normal traceback is impossible or has failed.
		// See if it falls into several common cases.
		n = 0
		if mp.Ncgo > 0 && mp.Curg != nil && mp.Curg.Syscallpc != 0 && mp.Curg.Syscallsp != 0 {
			// Cgo, we can't unwind and symbolize arbitrary C code,
			// so instead collect Go stack that leads to the cgo call.
			// This is especially important on windows, since all syscalls are cgo calls.
			n = int32(_lock.Gentraceback(mp.Curg.Syscallpc, mp.Curg.Syscallsp, 0, mp.Curg, 0, &stk[0], len(stk), nil, nil, 0))
		}
		if _lock.GOOS == "windows" && n == 0 && mp.Libcallg != nil && mp.Libcallpc != 0 && mp.Libcallsp != 0 {
			// Libcall, i.e. runtime syscall on windows.
			// Collect Go stack that leads to the call.
			n = int32(_lock.Gentraceback(mp.Libcallpc, mp.Libcallsp, 0, mp.Libcallg, 0, &stk[0], len(stk), nil, nil, 0))
		}
		if n == 0 {
			// If all of the above has failed, account it against abstract "System" or "GC".
			n = 2
			// "ExternalCode" is better than "etext".
			if uintptr(unsafe.Pointer(pc)) > uintptr(unsafe.Pointer(&etext)) {
				pc = (*uint8)(unsafe.Pointer(uintptr(_lock.FuncPC(_ExternalCode) + _lock.PCQuantum)))
			}
			stk[0] = uintptr(unsafe.Pointer(pc))
			if mp.Gcing != 0 || mp.Helpgc != 0 {
				stk[1] = _lock.FuncPC(_GC) + _lock.PCQuantum
			} else {
				stk[1] = _lock.FuncPC(_System) + _lock.PCQuantum
			}
		}
	}

	if Prof.Hz != 0 {
		// Simple cas-lock to coordinate with setcpuprofilerate.
		for !Cas(&Prof.Lock, 0, 1) {
			_core.Osyield()
		}
		if Prof.Hz != 0 {
			cpuproftick(&stk[0], n)
		}
		_lock.Atomicstore(&Prof.Lock, 0)
	}
	mp.Mallocing--
}

// Associate p and the current m.
func Acquirep(_p_ *_core.P) {
	_g_ := _core.Getg()

	if _g_.M.P != nil || _g_.M.Mcache != nil {
		_lock.Gothrow("acquirep: already in go")
	}
	if _p_.M != nil || _p_.Status != _lock.Pidle {
		id := int32(0)
		if _p_.M != nil {
			id = _p_.M.Id
		}
		print("acquirep: p->m=", _p_.M, "(", id, ") p->status=", _p_.Status, "\n")
		_lock.Gothrow("acquirep: invalid p state")
	}
	_g_.M.Mcache = _p_.Mcache
	_g_.M.P = _p_
	_p_.M = _g_.M
	_p_.Status = _lock.Prunning
}

// Disassociate p and the current m.
func releasep() *_core.P {
	_g_ := _core.Getg()

	if _g_.M.P == nil || _g_.M.Mcache == nil {
		_lock.Gothrow("releasep: invalid arg")
	}
	_p_ := _g_.M.P
	if _p_.M != _g_.M || _p_.Mcache != _g_.M.Mcache || _p_.Status != _lock.Prunning {
		print("releasep: m=", _g_.M, " m->p=", _g_.M.P, " p->m=", _p_.M, " m->mcache=", _g_.M.Mcache, " p->mcache=", _p_.Mcache, " p->status=", _p_.Status, "\n")
		_lock.Gothrow("releasep: invalid p state")
	}
	_g_.M.P = nil
	_g_.M.Mcache = nil
	_p_.M = nil
	_p_.Status = _lock.Pidle
	return _p_
}

func Incidlelocked(v int32) {
	_lock.Lock(&_core.Sched.Lock)
	_core.Sched.Nmidlelocked += v
	if v > 0 {
		checkdead()
	}
	_lock.Unlock(&_core.Sched.Lock)
}

// Check for deadlock situation.
// The check is based on number of running M's, if 0 -> deadlock.
func checkdead() {
	// If we are dying because of a signal caught on an already idle thread,
	// freezetheworld will cause all running threads to block.
	// And runtime will essentially enter into deadlock state,
	// except that there is a thread that will call exit soon.
	if _lock.Panicking > 0 {
		return
	}

	// -1 for sysmon
	run := _core.Sched.Mcount - _core.Sched.Nmidle - _core.Sched.Nmidlelocked - 1
	if run > 0 {
		return
	}
	if run < 0 {
		print("runtime: checkdead: nmidle=", _core.Sched.Nmidle, " nmidlelocked=", _core.Sched.Nmidlelocked, " mcount=", _core.Sched.Mcount, "\n")
		_lock.Gothrow("checkdead: inconsistent counts")
	}

	grunning := 0
	_lock.Lock(&_lock.Allglock)
	for i := 0; i < len(_lock.Allgs); i++ {
		gp := _lock.Allgs[i]
		if gp.Issystem {
			continue
		}
		s := _lock.Readgstatus(gp)
		switch s &^ _lock.Gscan {
		case _lock.Gwaiting:
			grunning++
		case _lock.Grunnable,
			_lock.Grunning,
			_lock.Gsyscall:
			_lock.Unlock(&_lock.Allglock)
			print("runtime: checkdead: find g ", gp.Goid, " in status ", s, "\n")
			_lock.Gothrow("checkdead: runnable g")
		}
	}
	_lock.Unlock(&_lock.Allglock)
	if grunning == 0 { // possible if main goroutine calls runtime·Goexit()
		_lock.Gothrow("no goroutines (main called runtime.Goexit) - deadlock!")
	}

	// Maybe jump time forward for playground.
	gp := timejump()
	if gp != nil {
		Casgstatus(gp, _lock.Gwaiting, _lock.Grunnable)
		globrunqput(gp)
		_p_ := Pidleget()
		if _p_ == nil {
			_lock.Gothrow("checkdead: no p for timer")
		}
		mp := Mget()
		if mp == nil {
			Newm(nil, _p_)
		} else {
			mp.Nextp = _p_
			Notewakeup(&mp.Park)
		}
		return
	}

	_core.Getg().M.Throwing = -1 // do not dump full stacks
	_lock.Gothrow("all goroutines are asleep - deadlock!")
}

// Put mp on midle list.
// Sched must be locked.
func mput(mp *_core.M) {
	mp.Schedlink = _core.Sched.Midle
	_core.Sched.Midle = mp
	_core.Sched.Nmidle++
	checkdead()
}

// Try to get an m from midle list.
// Sched must be locked.
func Mget() *_core.M {
	mp := _core.Sched.Midle
	if mp != nil {
		_core.Sched.Midle = mp.Schedlink
		_core.Sched.Nmidle--
	}
	return mp
}

// Put gp on the global runnable queue.
// Sched must be locked.
func globrunqput(gp *_core.G) {
	gp.Schedlink = nil
	if _core.Sched.Runqtail != nil {
		_core.Sched.Runqtail.Schedlink = gp
	} else {
		_core.Sched.Runqhead = gp
	}
	_core.Sched.Runqtail = gp
	_core.Sched.Runqsize++
}

// Put a batch of runnable goroutines on the global runnable queue.
// Sched must be locked.
func globrunqputbatch(ghead *_core.G, gtail *_core.G, n int32) {
	gtail.Schedlink = nil
	if _core.Sched.Runqtail != nil {
		_core.Sched.Runqtail.Schedlink = ghead
	} else {
		_core.Sched.Runqhead = ghead
	}
	_core.Sched.Runqtail = gtail
	_core.Sched.Runqsize += n
}

// Try get a batch of G's from the global runnable queue.
// Sched must be locked.
func globrunqget(_p_ *_core.P, max int32) *_core.G {
	if _core.Sched.Runqsize == 0 {
		return nil
	}

	n := _core.Sched.Runqsize/_lock.Gomaxprocs + 1
	if n > _core.Sched.Runqsize {
		n = _core.Sched.Runqsize
	}
	if max > 0 && n > max {
		n = max
	}
	if n > int32(len(_p_.Runq))/2 {
		n = int32(len(_p_.Runq)) / 2
	}

	_core.Sched.Runqsize -= n
	if _core.Sched.Runqsize == 0 {
		_core.Sched.Runqtail = nil
	}

	gp := _core.Sched.Runqhead
	_core.Sched.Runqhead = gp.Schedlink
	n--
	for ; n > 0; n-- {
		gp1 := _core.Sched.Runqhead
		_core.Sched.Runqhead = gp1.Schedlink
		Runqput(_p_, gp1)
	}
	return gp
}

// Put p to on _Pidle list.
// Sched must be locked.
func Pidleput(_p_ *_core.P) {
	_p_.Link = _core.Sched.Pidle
	_core.Sched.Pidle = _p_
	_lock.Xadd(&_core.Sched.Npidle, 1) // TODO: fast atomic
}

// Try get a p from _Pidle list.
// Sched must be locked.
func Pidleget() *_core.P {
	_p_ := _core.Sched.Pidle
	if _p_ != nil {
		_core.Sched.Pidle = _p_.Link
		_lock.Xadd(&_core.Sched.Npidle, -1) // TODO: fast atomic
	}
	return _p_
}

// Try to put g on local runnable queue.
// If it's full, put onto global queue.
// Executed only by the owner P.
func Runqput(_p_ *_core.P, gp *_core.G) {
retry:
	h := _lock.Atomicload(&_p_.Runqhead) // load-acquire, synchronize with consumers
	t := _p_.Runqtail
	if t-h < uint32(len(_p_.Runq)) {
		_p_.Runq[t%uint32(len(_p_.Runq))] = gp
		_lock.Atomicstore(&_p_.Runqtail, t+1) // store-release, makes the item available for consumption
		return
	}
	if runqputslow(_p_, gp, h, t) {
		return
	}
	// the queue is not full, now the put above must suceed
	goto retry
}

// Put g and a batch of work from local runnable queue on global queue.
// Executed only by the owner P.
func runqputslow(_p_ *_core.P, gp *_core.G, h, t uint32) bool {
	var batch [len(_p_.Runq)/2 + 1]*_core.G

	// First, grab a batch from local queue.
	n := t - h
	n = n / 2
	if n != uint32(len(_p_.Runq)/2) {
		_lock.Gothrow("runqputslow: queue is not full")
	}
	for i := uint32(0); i < n; i++ {
		batch[i] = _p_.Runq[(h+i)%uint32(len(_p_.Runq))]
	}
	if !Cas(&_p_.Runqhead, h, h+n) { // cas-release, commits consume
		return false
	}
	batch[n] = gp

	// Link the goroutines.
	for i := uint32(0); i < n; i++ {
		batch[i].Schedlink = batch[i+1]
	}

	// Now put the batch on global queue.
	_lock.Lock(&_core.Sched.Lock)
	globrunqputbatch(batch[0], batch[n], int32(n+1))
	_lock.Unlock(&_core.Sched.Lock)
	return true
}

// Get g from local runnable queue.
// Executed only by the owner P.
func Runqget(_p_ *_core.P) *_core.G {
	for {
		h := _lock.Atomicload(&_p_.Runqhead) // load-acquire, synchronize with other consumers
		t := _p_.Runqtail
		if t == h {
			return nil
		}
		gp := _p_.Runq[h%uint32(len(_p_.Runq))]
		if Cas(&_p_.Runqhead, h, h+1) { // cas-release, commits consume
			return gp
		}
	}
}

// Grabs a batch of goroutines from local runnable queue.
// batch array must be of size len(p->runq)/2. Returns number of grabbed goroutines.
// Can be executed by any P.
func runqgrab(_p_ *_core.P, batch []*_core.G) uint32 {
	for {
		h := _lock.Atomicload(&_p_.Runqhead) // load-acquire, synchronize with other consumers
		t := _lock.Atomicload(&_p_.Runqtail) // load-acquire, synchronize with the producer
		n := t - h
		n = n - n/2
		if n == 0 {
			return 0
		}
		if n > uint32(len(_p_.Runq)/2) { // read inconsistent h and t
			continue
		}
		for i := uint32(0); i < n; i++ {
			batch[i] = _p_.Runq[(h+i)%uint32(len(_p_.Runq))]
		}
		if Cas(&_p_.Runqhead, h, h+n) { // cas-release, commits consume
			return n
		}
	}
}

// Steal half of elements from local runnable queue of p2
// and put onto local runnable queue of p.
// Returns one of the stolen elements (or nil if failed).
func Runqsteal(_p_, p2 *_core.P) *_core.G {
	var batch [len(_p_.Runq) / 2]*_core.G

	n := runqgrab(p2, batch[:])
	if n == 0 {
		return nil
	}
	n--
	gp := batch[n]
	if n == 0 {
		return gp
	}
	h := _lock.Atomicload(&_p_.Runqhead) // load-acquire, synchronize with consumers
	t := _p_.Runqtail
	if t-h+n >= uint32(len(_p_.Runq)) {
		_lock.Gothrow("runqsteal: runq overflow")
	}
	for i := uint32(0); i < n; i++ {
		_p_.Runq[(t+i)%uint32(len(_p_.Runq))] = batch[i]
	}
	_lock.Atomicstore(&_p_.Runqtail, t+n) // store-release, makes the item available for consumption
	return gp
}
