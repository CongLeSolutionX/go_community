// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package base

import (
	"unsafe"
)

var (
	M0 M
)

func Dumpgstatus(gp *G) {
	_g_ := Getg()
	print("runtime: gp: gp=", gp, ", goid=", gp.Goid, ", gp->atomicstatus=", Readgstatus(gp), "\n")
	print("runtime:  g:  g=", _g_, ", goid=", _g_.Goid, ",  g->atomicstatus=", Readgstatus(_g_), "\n")
}

func Checkmcount() {
	// sched lock is held
	if Sched.mcount > Sched.Maxmcount {
		print("runtime: program exceeds ", Sched.Maxmcount, "-thread limit\n")
		Throw("thread exhaustion")
	}
}

func Mcommoninit(mp *M) {
	_g_ := Getg()

	// g0 stack won't make sense for user (and is not necessary unwindable).
	if _g_ != _g_.M.G0 {
		Callers(1, mp.Createstack[:])
	}

	mp.fastrand = 0x49f6428a + uint32(mp.Id) + uint32(Cputicks())
	if mp.fastrand == 0 {
		mp.fastrand = 0x49f6428a
	}

	Lock(&Sched.Lock)
	mp.Id = Sched.mcount
	Sched.mcount++
	Checkmcount()
	mpreinit(mp)
	if mp.Gsignal != nil {
		mp.Gsignal.stackguard1 = mp.Gsignal.Stack.Lo + StackGuard
	}

	// Add to allm so garbage collector doesn't free g->m
	// when it is just in a register or thread-local storage.
	mp.Alllink = Allm

	// NumCgoCall() iterates over allm w/o schedlock,
	// so we need to publish it safely.
	Atomicstorep(unsafe.Pointer(&Allm), unsafe.Pointer(mp))
	Unlock(&Sched.Lock)
}

// Mark gp ready to run.
func Ready(gp *G, traceskip int) {
	if Trace.Enabled {
		TraceGoUnpark(gp, traceskip)
	}

	status := Readgstatus(gp)

	// Mark runnable.
	_g_ := Getg()
	_g_.M.Locks++ // disable preemption because it can be holding p in a local var
	if status&^Gscan != Gwaiting {
		Dumpgstatus(gp)
		Throw("bad g->status in ready")
	}

	// status is Gwaiting or Gscanwaiting, make Grunnable and put on runq
	Casgstatus(gp, Gwaiting, Grunnable)
	Runqput(_g_.M.P.Ptr(), gp, true)
	if Atomicload(&Sched.Npidle) != 0 && Atomicload(&Sched.Nmspinning) == 0 { // TODO: fast atomic
		Wakep()
	}
	_g_.M.Locks--
	if _g_.M.Locks == 0 && _g_.Preempt { // restore the preemption request in case we've cleared it in newstack
		_g_.Stackguard0 = StackPreempt
	}
}

// freezeStopWait is a large value that freezetheworld sets
// sched.stopwait to in order to request that all Gs permanently stop.
const freezeStopWait = 0x7fffffff

// Similar to stopTheWorld but best-effort and can be called several times.
// There is no reverse operation, used during crashing.
// This function must not lock any mutexes.
func freezetheworld() {
	// stopwait and preemption requests can be lost
	// due to races with concurrently executing threads,
	// so try several times
	for i := 0; i < 5; i++ {
		// this should tell the scheduler to not start any new goroutines
		Sched.Stopwait = freezeStopWait
		Atomicstore(&Sched.Gcwaiting, 1)
		// this should stop running goroutines
		if !Preemptall() {
			break // no running goroutines
		}
		Usleep(1000)
	}
	// to be sure
	Usleep(1000)
	Preemptall()
	Usleep(1000)
}

// All reads and writes of g's status go through readgstatus, casgstatus
// castogscanstatus, casfrom_Gscanstatus.
//go:nosplit
func Readgstatus(gp *G) uint32 {
	return Atomicload(&gp.Atomicstatus)
}

// If asked to move to or from a Gscanstatus this will throw. Use the castogscanstatus
// and casfrom_Gscanstatus instead.
// casgstatus will loop if the g->atomicstatus is in a Gscan status until the routine that
// put it in the Gscan state is finished.
//go:nosplit
func Casgstatus(gp *G, oldval, newval uint32) {
	if (oldval&Gscan != 0) || (newval&Gscan != 0) || oldval == newval {
		Systemstack(func() {
			print("runtime: casgstatus: oldval=", Hex(oldval), " newval=", Hex(newval), "\n")
			Throw("casgstatus: bad incoming values")
		})
	}

	if oldval == Grunning && gp.Gcscanvalid {
		// If oldvall == _Grunning, then the actual status must be
		// _Grunning or _Grunning|_Gscan; either way,
		// we own gp.gcscanvalid, so it's safe to read.
		// gp.gcscanvalid must not be true when we are running.
		print("runtime: casgstatus ", Hex(oldval), "->", Hex(newval), " gp.status=", Hex(gp.Atomicstatus), " gp.gcscanvalid=true\n")
		Throw("casgstatus")
	}

	// loop if gp->atomicstatus is in a scan state giving
	// GC time to finish and change the state to oldval.
	for !Cas(&gp.Atomicstatus, oldval, newval) {
		if oldval == Gwaiting && gp.Atomicstatus == Grunnable {
			Systemstack(func() {
				Throw("casgstatus: waiting for Gwaiting but is Grunnable")
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
	if newval == Grunning {
		gp.Gcscanvalid = false
	}
}

// Called to start an M.
//go:nosplit
func Mstart() {
	_g_ := Getg()

	if _g_.Stack.Lo == 0 {
		// Initialize stack bounds from system stack.
		// Cgo may have left stack size in stack.hi.
		size := _g_.Stack.Hi
		if size == 0 {
			size = 8192 * stackGuardMultiplier
		}
		_g_.Stack.Hi = uintptr(Noescape(unsafe.Pointer(&size)))
		_g_.Stack.Lo = _g_.Stack.Hi - size + 1024
	}
	// Initialize stack guards so that we can start calling
	// both Go and C functions with stack growth prologues.
	_g_.Stackguard0 = _g_.Stack.Lo + StackGuard
	_g_.stackguard1 = _g_.Stackguard0
	mstart1()
}

func mstart1() {
	_g_ := Getg()

	if _g_ != _g_.M.G0 {
		Throw("bad runtime·mstart")
	}

	// Record top of stack for use by mcall.
	// Once we call schedule we're never coming back,
	// so other calls can reuse this stack space.
	gosave(&_g_.M.G0.Sched)
	_g_.M.G0.Sched.Pc = ^uintptr(0) // make sure it is never used
	Asminit()
	Minit()

	// Install signal handlers; after minit so that minit can
	// prepare the thread to be able to handle the signals.
	if _g_.M == &M0 {
		// Create an extra M for callbacks on threads not created by Go.
		if Iscgo && !CgoHasExtraM {
			CgoHasExtraM = true
			Newextram()
		}
		initsig()
	}

	if fn := _g_.M.mstartfn; fn != nil {
		fn()
	}

	if _g_.M.Helpgc != 0 {
		_g_.M.Helpgc = 0
		stopm()
	} else if _g_.M != &M0 {
		Acquirep(_g_.M.Nextp.Ptr())
		_g_.M.Nextp = 0
	}
	Schedule()
}

// runSafePointFn runs the safe point function, if any, for this P.
// This should be called like
//
//     if getg().m.p.runSafePointFn != 0 {
//         runSafePointFn()
//     }
//
// runSafePointFn must be checked on any transition in to _Pidle or
// _Psyscall to avoid a race where forEachP sees that the P is running
// just before the P goes into _Pidle/_Psyscall and neither forEachP
// nor the P run the safe-point function.
func RunSafePointFn() {
	p := Getg().M.P.Ptr()
	// Resolve the race between forEachP running the safe-point
	// function on this P's behalf and this P running the
	// safe-point function directly.
	if !Cas(&p.RunSafePointFn, 1, 0) {
		return
	}
	Sched.SafePointFn(p)
	Lock(&Sched.Lock)
	Sched.SafePointWait--
	if Sched.SafePointWait == 0 {
		Notewakeup(&Sched.SafePointNote)
	}
	Unlock(&Sched.Lock)
}

type cgothreadstart struct {
	g   Guintptr
	tls *uint64
	fn  unsafe.Pointer
}

// Allocate a new m unassociated with any thread.
// Can use p for allocation context if needed.
// fn is recorded as the new m's m.mstartfn.
func allocm(_p_ *P, fn func()) *M {
	_g_ := Getg()
	_g_.M.Locks++ // disable GC because it can be called from sysmon
	if _g_.M.P == 0 {
		Acquirep(_p_) // temporarily borrow p for mallocs in this function
	}
	mp := new(M)
	mp.mstartfn = fn
	Mcommoninit(mp)

	// In case of cgo or Solaris, pthread_create will make us a stack.
	// Windows and Plan 9 will layout sched stack on OS stack.
	if Iscgo || GOOS == "solaris" || GOOS == "windows" || GOOS == "plan9" {
		mp.G0 = Malg(-1)
	} else {
		mp.G0 = Malg(8192 * stackGuardMultiplier)
	}
	mp.G0.M = mp

	if _p_ == _g_.M.P.Ptr() {
		releasep()
	}
	_g_.M.Locks--
	if _g_.M.Locks == 0 && _g_.Preempt { // restore the preemption request in case we've cleared it in newstack
		_g_.Stackguard0 = StackPreempt
	}

	return mp
}

// newextram allocates an m and puts it on the extra list.
// It is called with a working local m, so that it can do things
// like call schedlock and allocate.
func Newextram() {
	// Create extra goroutine locked to extra m.
	// The goroutine is the context in which the cgo callback will run.
	// The sched.pc will never be returned to, but setting it to
	// goexit makes clear to the traceback routines where
	// the goroutine stack ends.
	mp := allocm(nil, nil)
	gp := Malg(4096)
	gp.Sched.Pc = FuncPC(Goexit) + PCQuantum
	gp.Sched.Sp = gp.Stack.Hi
	gp.Sched.Sp -= 4 * RegSize // extra space in case of reads slightly beyond frame
	gp.Sched.Lr = 0
	gp.Sched.G = Guintptr(unsafe.Pointer(gp))
	gp.Syscallpc = gp.Sched.Pc
	gp.Syscallsp = gp.Sched.Sp
	// malg returns status as Gidle, change to Gsyscall before adding to allg
	// where GC will see it.
	Casgstatus(gp, Gidle, Gsyscall)
	gp.M = mp
	mp.Curg = gp
	mp.Locked = LockInternal
	mp.Lockedg = gp
	gp.Lockedm = mp
	gp.Goid = int64(Xadd64(&Sched.Goidgen, 1))
	if Raceenabled {
		gp.Racectx = Racegostart(FuncPC(Newextram))
	}
	// put on allg for garbage collector
	Allgadd(gp)

	// Add m to the extra list.
	mnext := Lockextra(true)
	mp.Schedlink.Set(mnext)
	Unlockextra(mp)
}

var extram uintptr

// lockextra locks the extra list and returns the list head.
// The caller must unlock the list by storing a new list head
// to extram. If nilokay is true, then lockextra will
// return a nil list head if that's what it finds. If nilokay is false,
// lockextra will keep waiting until the list head is no longer nil.
//go:nosplit
func Lockextra(nilokay bool) *M {
	const locked = 1

	for {
		old := Atomicloaduintptr(&extram)
		if old == locked {
			yield := Osyield
			yield()
			continue
		}
		if old == 0 && !nilokay {
			Usleep(1)
			continue
		}
		if Casuintptr(&extram, old, locked) {
			return (*M)(unsafe.Pointer(old))
		}
		yield := Osyield
		yield()
		continue
	}
}

//go:nosplit
func Unlockextra(mp *M) {
	atomicstoreuintptr(&extram, uintptr(unsafe.Pointer(mp)))
}

// Create a new m.  It will start off with a call to fn, or else the scheduler.
// fn needs to be static and not a heap allocated closure.
// May run with m.p==nil, so write barriers are not allowed.
//go:nowritebarrier
func Newm(fn func(), _p_ *P) {
	mp := allocm(_p_, fn)
	mp.Nextp.Set(_p_)
	Msigsave(mp)
	if Iscgo {
		var ts cgothreadstart
		if Cgo_thread_start == nil {
			Throw("_cgo_thread_start missing")
		}
		ts.g.Set(mp.G0)
		ts.tls = (*uint64)(unsafe.Pointer(&mp.tls[0]))
		ts.fn = unsafe.Pointer(FuncPC(Mstart))
		Asmcgocall(Cgo_thread_start, unsafe.Pointer(&ts))
		return
	}
	newosproc(mp, unsafe.Pointer(mp.G0.Stack.Hi))
}

// Stops execution of the current m until new work is available.
// Returns with acquired P.
func stopm() {
	_g_ := Getg()

	if _g_.M.Locks != 0 {
		Throw("stopm holding locks")
	}
	if _g_.M.P != 0 {
		Throw("stopm holding p")
	}
	if _g_.M.spinning {
		_g_.M.spinning = false
		Xadd(&Sched.Nmspinning, -1)
	}

retry:
	Lock(&Sched.Lock)
	mput(_g_.M)
	Unlock(&Sched.Lock)
	Notesleep(&_g_.M.Park)
	Noteclear(&_g_.M.Park)
	if _g_.M.Helpgc != 0 {
		gchelper()
		_g_.M.Helpgc = 0
		_g_.M.Mcache = nil
		_g_.M.P = 0
		goto retry
	}
	Acquirep(_g_.M.Nextp.Ptr())
	_g_.M.Nextp = 0
}

func mspinning() {
	gp := Getg()
	if !Runqempty(gp.M.Nextp.Ptr()) {
		// Something (presumably the GC) was readied while the
		// runtime was starting up this M, so the M is no
		// longer spinning.
		if int32(Xadd(&Sched.Nmspinning, -1)) < 0 {
			Throw("mspinning: nmspinning underflowed")
		}
	} else {
		gp.M.spinning = true
	}
}

// Schedules some M to run the p (creates an M if necessary).
// If p==nil, tries to get an idle P, if no idle P's does nothing.
// May run with m.p==nil, so write barriers are not allowed.
//go:nowritebarrier
func startm(_p_ *P, spinning bool) {
	Lock(&Sched.Lock)
	if _p_ == nil {
		_p_ = Pidleget()
		if _p_ == nil {
			Unlock(&Sched.Lock)
			if spinning {
				Xadd(&Sched.Nmspinning, -1)
			}
			return
		}
	}
	mp := Mget()
	Unlock(&Sched.Lock)
	if mp == nil {
		var fn func()
		if spinning {
			fn = mspinning
		}
		Newm(fn, _p_)
		return
	}
	if mp.spinning {
		Throw("startm: m is spinning")
	}
	if mp.Nextp != 0 {
		Throw("startm: m has p")
	}
	if spinning && !Runqempty(_p_) {
		Throw("startm: p has runnable gs")
	}
	mp.spinning = spinning
	mp.Nextp.Set(_p_)
	Notewakeup(&mp.Park)
}

// Hands off P from syscall or locked M.
// Always runs without a P, so write barriers are not allowed.
//go:nowritebarrier
func Handoffp(_p_ *P) {
	// if it has local work, start it straight away
	if !Runqempty(_p_) || Sched.Runqsize != 0 {
		startm(_p_, false)
		return
	}
	// no local work, check that there are no spinning/idle M's,
	// otherwise our help is not required
	if Atomicload(&Sched.Nmspinning)+Atomicload(&Sched.Npidle) == 0 && Cas(&Sched.Nmspinning, 0, 1) { // TODO: fast atomic
		startm(_p_, true)
		return
	}
	Lock(&Sched.Lock)
	if Sched.Gcwaiting != 0 {
		_p_.Status = Pgcstop
		Sched.Stopwait--
		if Sched.Stopwait == 0 {
			Notewakeup(&Sched.Stopnote)
		}
		Unlock(&Sched.Lock)
		return
	}
	if _p_.RunSafePointFn != 0 && Cas(&_p_.RunSafePointFn, 1, 0) {
		Sched.SafePointFn(_p_)
		Sched.SafePointWait--
		if Sched.SafePointWait == 0 {
			Notewakeup(&Sched.SafePointNote)
		}
	}
	if Sched.Runqsize != 0 {
		Unlock(&Sched.Lock)
		startm(_p_, false)
		return
	}
	// If this is the last running P and nobody is polling network,
	// need to wakeup another M to poll network.
	if Sched.Npidle == uint32(Gomaxprocs-1) && Atomicload64(&Sched.Lastpoll) != 0 {
		Unlock(&Sched.Lock)
		startm(_p_, false)
		return
	}
	Pidleput(_p_)
	Unlock(&Sched.Lock)
}

// Tries to add one more P to execute G's.
// Called when a G is made runnable (newproc, ready).
func Wakep() {
	// be conservative about spinning threads
	if !Cas(&Sched.Nmspinning, 0, 1) {
		return
	}
	startm(nil, true)
}

// Stops execution of the current m that is locked to a g until the g is runnable again.
// Returns with acquired P.
func stoplockedm() {
	_g_ := Getg()

	if _g_.M.Lockedg == nil || _g_.M.Lockedg.Lockedm != _g_.M {
		Throw("stoplockedm: inconsistent locking")
	}
	if _g_.M.P != 0 {
		// Schedule another M to run this p.
		_p_ := releasep()
		Handoffp(_p_)
	}
	Incidlelocked(1)
	// Wait until another thread schedules lockedg again.
	Notesleep(&_g_.M.Park)
	Noteclear(&_g_.M.Park)
	status := Readgstatus(_g_.M.Lockedg)
	if status&^Gscan != Grunnable {
		print("runtime:stoplockedm: g is not Grunnable or Gscanrunnable\n")
		Dumpgstatus(_g_)
		Throw("stoplockedm: not runnable")
	}
	Acquirep(_g_.M.Nextp.Ptr())
	_g_.M.Nextp = 0
}

// Schedules the locked m to run the locked gp.
// May run during STW, so write barriers are not allowed.
//go:nowritebarrier
func startlockedm(gp *G) {
	_g_ := Getg()

	mp := gp.Lockedm
	if mp == _g_.M {
		Throw("startlockedm: locked to me")
	}
	if mp.Nextp != 0 {
		Throw("startlockedm: m has p")
	}
	// directly handoff current P to the locked m
	Incidlelocked(-1)
	_p_ := releasep()
	mp.Nextp.Set(_p_)
	Notewakeup(&mp.Park)
	stopm()
}

// Stops the current m for stopTheWorld.
// Returns when the world is restarted.
func gcstopm() {
	_g_ := Getg()

	if Sched.Gcwaiting == 0 {
		Throw("gcstopm: not waiting for gc")
	}
	if _g_.M.spinning {
		_g_.M.spinning = false
		Xadd(&Sched.Nmspinning, -1)
	}
	_p_ := releasep()
	Lock(&Sched.Lock)
	_p_.Status = Pgcstop
	Sched.Stopwait--
	if Sched.Stopwait == 0 {
		Notewakeup(&Sched.Stopnote)
	}
	Unlock(&Sched.Lock)
	stopm()
}

// Schedules gp to run on the current M.
// If inheritTime is true, gp inherits the remaining time in the
// current time slice. Otherwise, it starts a new time slice.
// Never returns.
func execute(gp *G, inheritTime bool) {
	_g_ := Getg()

	Casgstatus(gp, Grunnable, Grunning)
	gp.Waitsince = 0
	gp.Preempt = false
	gp.Stackguard0 = gp.Stack.Lo + StackGuard
	if !inheritTime {
		_g_.M.P.Ptr().Schedtick++
	}
	_g_.M.Curg = gp
	gp.M = _g_.M

	// Check whether the profiler needs to be turned on or off.
	hz := Sched.Profilehz
	if _g_.M.Profilehz != hz {
		Resetcpuprofiler(hz)
	}

	if Trace.Enabled {
		// GoSysExit has to happen when we have a P, but before GoStart.
		// So we emit it here.
		if gp.Syscallsp != 0 && gp.Sysblocktraced {
			traceGoSysExit(gp.sysexitticks)
		}
		TraceGoStart()
	}

	Gogo(&gp.Sched)
}

// Finds a runnable goroutine to execute.
// Tries to steal from other P's, get g from global queue, poll network.
func findrunnable() (gp *G, inheritTime bool) {
	_g_ := Getg()

top:
	if Sched.Gcwaiting != 0 {
		gcstopm()
		goto top
	}
	if _g_.M.P.Ptr().RunSafePointFn != 0 {
		RunSafePointFn()
	}
	if Fingwait && Fingwake {
		if gp := wakefing(); gp != nil {
			Ready(gp, 0)
		}
	}

	// local runq
	if gp, inheritTime := Runqget(_g_.M.P.Ptr()); gp != nil {
		return gp, inheritTime
	}

	// global runq
	if Sched.Runqsize != 0 {
		Lock(&Sched.Lock)
		gp := globrunqget(_g_.M.P.Ptr(), 0)
		Unlock(&Sched.Lock)
		if gp != nil {
			return gp, false
		}
	}

	// Poll network.
	// This netpoll is only an optimization before we resort to stealing.
	// We can safely skip it if there a thread blocked in netpoll already.
	// If there is any kind of logical race with that blocked thread
	// (e.g. it has already returned from netpoll, but does not set lastpoll yet),
	// this thread will do blocking netpoll below anyway.
	if netpollinited() && Sched.Lastpoll != 0 {
		if gp := Netpoll(false); gp != nil { // non-blocking
			// netpoll returns list of goroutines linked by schedlink.
			Injectglist(gp.Schedlink.Ptr())
			Casgstatus(gp, Gwaiting, Grunnable)
			if Trace.Enabled {
				TraceGoUnpark(gp, 0)
			}
			return gp, false
		}
	}

	// If number of spinning M's >= number of busy P's, block.
	// This is necessary to prevent excessive CPU consumption
	// when GOMAXPROCS>>1 but the program parallelism is low.
	if !_g_.M.spinning && 2*Atomicload(&Sched.Nmspinning) >= uint32(Gomaxprocs)-Atomicload(&Sched.Npidle) { // TODO: fast atomic
		goto stop
	}
	if !_g_.M.spinning {
		_g_.M.spinning = true
		Xadd(&Sched.Nmspinning, 1)
	}
	// random steal from other P's
	for i := 0; i < int(4*Gomaxprocs); i++ {
		if Sched.Gcwaiting != 0 {
			goto top
		}
		_p_ := Allp[Fastrand1()%uint32(Gomaxprocs)]
		var gp *G
		if _p_ == _g_.M.P.Ptr() {
			gp, _ = Runqget(_p_)
		} else {
			stealRunNextG := i > 2*int(Gomaxprocs) // first look for ready queues with more than 1 g
			gp = Runqsteal(_g_.M.P.Ptr(), _p_, stealRunNextG)
		}
		if gp != nil {
			return gp, false
		}
	}

stop:

	// We have nothing to do. If we're in the GC mark phase and can
	// safely scan and blacken objects, run idle-time marking
	// rather than give up the P.
	if _p_ := _g_.M.P.Ptr(); GcBlackenEnabled != 0 && _p_.GcBgMarkWorker != nil && gcMarkWorkAvailable(_p_) {
		_p_.GcMarkWorkerMode = GcMarkWorkerIdleMode
		gp := _p_.GcBgMarkWorker
		Casgstatus(gp, Gwaiting, Grunnable)
		if Trace.Enabled {
			TraceGoUnpark(gp, 0)
		}
		return gp, false
	}

	// return P and block
	Lock(&Sched.Lock)
	if Sched.Gcwaiting != 0 || _g_.M.P.Ptr().RunSafePointFn != 0 {
		Unlock(&Sched.Lock)
		goto top
	}
	if Sched.Runqsize != 0 {
		gp := globrunqget(_g_.M.P.Ptr(), 0)
		Unlock(&Sched.Lock)
		return gp, false
	}
	_p_ := releasep()
	Pidleput(_p_)
	Unlock(&Sched.Lock)
	if _g_.M.spinning {
		_g_.M.spinning = false
		Xadd(&Sched.Nmspinning, -1)
	}

	// check all runqueues once again
	for i := 0; i < int(Gomaxprocs); i++ {
		_p_ := Allp[i]
		if _p_ != nil && !Runqempty(_p_) {
			Lock(&Sched.Lock)
			_p_ = Pidleget()
			Unlock(&Sched.Lock)
			if _p_ != nil {
				Acquirep(_p_)
				goto top
			}
			break
		}
	}

	// poll network
	if netpollinited() && Xchg64(&Sched.Lastpoll, 0) != 0 {
		if _g_.M.P != 0 {
			Throw("findrunnable: netpoll with p")
		}
		if _g_.M.spinning {
			Throw("findrunnable: netpoll with spinning")
		}
		gp := Netpoll(true) // block until new work is available
		Atomicstore64(&Sched.Lastpoll, uint64(Nanotime()))
		if gp != nil {
			Lock(&Sched.Lock)
			_p_ = Pidleget()
			Unlock(&Sched.Lock)
			if _p_ != nil {
				Acquirep(_p_)
				Injectglist(gp.Schedlink.Ptr())
				Casgstatus(gp, Gwaiting, Grunnable)
				if Trace.Enabled {
					TraceGoUnpark(gp, 0)
				}
				return gp, false
			}
			Injectglist(gp)
		}
	}
	stopm()
	goto top
}

func resetspinning() {
	_g_ := Getg()

	var nmspinning uint32
	if _g_.M.spinning {
		_g_.M.spinning = false
		nmspinning = Xadd(&Sched.Nmspinning, -1)
		if nmspinning < 0 {
			Throw("findrunnable: negative nmspinning")
		}
	} else {
		nmspinning = Atomicload(&Sched.Nmspinning)
	}

	// M wakeup policy is deliberately somewhat conservative (see nmspinning handling),
	// so see if we need to wakeup another P here.
	if nmspinning == 0 && Atomicload(&Sched.Npidle) > 0 {
		Wakep()
	}
}

// Injects the list of runnable G's into the scheduler.
// Can run concurrently with GC.
func Injectglist(glist *G) {
	if glist == nil {
		return
	}
	if Trace.Enabled {
		for gp := glist; gp != nil; gp = gp.Schedlink.Ptr() {
			TraceGoUnpark(gp, 0)
		}
	}
	Lock(&Sched.Lock)
	var n int
	for n = 0; glist != nil; n++ {
		gp := glist
		glist = gp.Schedlink.Ptr()
		Casgstatus(gp, Gwaiting, Grunnable)
		Globrunqput(gp)
	}
	Unlock(&Sched.Lock)
	for ; n != 0 && Sched.Npidle != 0; n-- {
		startm(nil, false)
	}
}

// One round of scheduler: find a runnable goroutine and execute it.
// Never returns.
func Schedule() {
	_g_ := Getg()

	if _g_.M.Locks != 0 {
		Throw("schedule: holding locks")
	}

	if _g_.M.Lockedg != nil {
		stoplockedm()
		execute(_g_.M.Lockedg, false) // Never returns.
	}

top:
	if Sched.Gcwaiting != 0 {
		gcstopm()
		goto top
	}
	if _g_.M.P.Ptr().RunSafePointFn != 0 {
		RunSafePointFn()
	}

	var gp *G
	var inheritTime bool
	if Trace.Enabled || Trace.Shutdown {
		gp = traceReader()
		if gp != nil {
			Casgstatus(gp, Gwaiting, Grunnable)
			TraceGoUnpark(gp, 0)
			resetspinning()
		}
	}
	if gp == nil && GcBlackenEnabled != 0 {
		gp = GcController.findRunnableGCWorker(_g_.M.P.Ptr())
		if gp != nil {
			resetspinning()
		}
	}
	if gp == nil {
		// Check the global runnable queue once in a while to ensure fairness.
		// Otherwise two goroutines can completely occupy the local runqueue
		// by constantly respawning each other.
		if _g_.M.P.Ptr().Schedtick%61 == 0 && Sched.Runqsize > 0 {
			Lock(&Sched.Lock)
			gp = globrunqget(_g_.M.P.Ptr(), 1)
			Unlock(&Sched.Lock)
			if gp != nil {
				resetspinning()
			}
		}
	}
	if gp == nil {
		gp, inheritTime = Runqget(_g_.M.P.Ptr())
		if gp != nil && _g_.M.spinning {
			Throw("schedule: spinning with local work")
		}
	}
	if gp == nil {
		gp, inheritTime = findrunnable() // blocks until work is available
		resetspinning()
	}

	if gp.Lockedm != nil {
		// Hands off own p to the locked m,
		// then blocks waiting for a new p.
		startlockedm(gp)
		goto top
	}

	execute(gp, inheritTime)
}

// dropg removes the association between m and the current goroutine m->curg (gp for short).
// Typically a caller sets gp's status away from Grunning and then
// immediately calls dropg to finish the job. The caller is also responsible
// for arranging that gp will be restarted using ready at an
// appropriate time. After calling dropg and arranging for gp to be
// readied later, the caller can do other work but eventually should
// call schedule to restart the scheduling of goroutines on this m.
func Dropg() {
	_g_ := Getg()

	if _g_.M.Lockedg == nil {
		_g_.M.Curg.M = nil
		_g_.M.Curg = nil
	}
}

func parkunlock_c(gp *G, lock unsafe.Pointer) bool {
	Unlock((*Mutex)(lock))
	return true
}

// park continuation on g0.
func park_m(gp *G) {
	_g_ := Getg()

	if Trace.Enabled {
		traceGoPark(_g_.M.waittraceev, _g_.M.waittraceskip, gp)
	}

	Casgstatus(gp, Grunning, Gwaiting)
	Dropg()

	if _g_.M.waitunlockf != nil {
		fn := *(*func(*G, unsafe.Pointer) bool)(unsafe.Pointer(&_g_.M.waitunlockf))
		ok := fn(gp, _g_.M.waitlock)
		_g_.M.waitunlockf = nil
		_g_.M.waitlock = nil
		if !ok {
			if Trace.Enabled {
				TraceGoUnpark(gp, 2)
			}
			Casgstatus(gp, Gwaiting, Grunnable)
			execute(gp, true) // Schedule it back, never returns.
		}
	}
	Schedule()
}

//go:nosplit
//go:nowritebarrier
func Save(pc, sp uintptr) {
	_g_ := Getg()

	_g_.Sched.Pc = pc
	_g_.Sched.Sp = sp
	_g_.Sched.Lr = 0
	_g_.Sched.Ret = 0
	_g_.Sched.Ctxt = nil
	_g_.Sched.G = Guintptr(unsafe.Pointer(_g_))
}

// The same as entersyscall(), but with a hint that the syscall is blocking.
//go:nosplit
func entersyscallblock(dummy int32) {
	_g_ := Getg()

	_g_.M.Locks++ // see comment in entersyscall
	_g_.Throwsplit = true
	_g_.Stackguard0 = StackPreempt // see comment in entersyscall
	_g_.M.Syscalltick = _g_.M.P.Ptr().Syscalltick
	_g_.Sysblocktraced = true
	_g_.M.P.Ptr().Syscalltick++

	// Leave SP around for GC and traceback.
	pc := Getcallerpc(unsafe.Pointer(&dummy))
	sp := Getcallersp(unsafe.Pointer(&dummy))
	Save(pc, sp)
	_g_.Syscallsp = _g_.Sched.Sp
	_g_.Syscallpc = _g_.Sched.Pc
	if _g_.Syscallsp < _g_.Stack.Lo || _g_.Stack.Hi < _g_.Syscallsp {
		sp1 := sp
		sp2 := _g_.Sched.Sp
		sp3 := _g_.Syscallsp
		Systemstack(func() {
			print("entersyscallblock inconsistent ", Hex(sp1), " ", Hex(sp2), " ", Hex(sp3), " [", Hex(_g_.Stack.Lo), ",", Hex(_g_.Stack.Hi), "]\n")
			Throw("entersyscallblock")
		})
	}
	Casgstatus(_g_, Grunning, Gsyscall)
	if _g_.Syscallsp < _g_.Stack.Lo || _g_.Stack.Hi < _g_.Syscallsp {
		Systemstack(func() {
			print("entersyscallblock inconsistent ", Hex(sp), " ", Hex(_g_.Sched.Sp), " ", Hex(_g_.Syscallsp), " [", Hex(_g_.Stack.Lo), ",", Hex(_g_.Stack.Hi), "]\n")
			Throw("entersyscallblock")
		})
	}

	Systemstack(entersyscallblock_handoff)

	// Resave for traceback during blocked call.
	Save(Getcallerpc(unsafe.Pointer(&dummy)), Getcallersp(unsafe.Pointer(&dummy)))

	_g_.M.Locks--
}

func entersyscallblock_handoff() {
	if Trace.Enabled {
		TraceGoSysCall()
		TraceGoSysBlock(Getg().M.P.Ptr())
	}
	Handoffp(releasep())
}

// The goroutine g exited its system call.
// Arrange for it to run on a cpu again.
// This is called only from the go syscall library, not
// from the low-level system calls used by the
//go:nosplit
func Exitsyscall(dummy int32) {
	_g_ := Getg()

	_g_.M.Locks++ // see comment in entersyscall
	if Getcallersp(unsafe.Pointer(&dummy)) > _g_.Syscallsp {
		Throw("exitsyscall: syscall frame is no longer valid")
	}

	_g_.Waitsince = 0
	oldp := _g_.M.P.Ptr()
	if exitsyscallfast() {
		if _g_.M.Mcache == nil {
			Throw("lost mcache")
		}
		if Trace.Enabled {
			if oldp != _g_.M.P.Ptr() || _g_.M.Syscalltick != _g_.M.P.Ptr().Syscalltick {
				Systemstack(TraceGoStart)
			}
		}
		// There's a cpu for us, so we can run.
		_g_.M.P.Ptr().Syscalltick++
		// We need to cas the status and scan before resuming...
		Casgstatus(_g_, Gsyscall, Grunning)

		// Garbage collector isn't running (since we are),
		// so okay to clear syscallsp.
		_g_.Syscallsp = 0
		_g_.M.Locks--
		if _g_.Preempt {
			// restore the preemption request in case we've cleared it in newstack
			_g_.Stackguard0 = StackPreempt
		} else {
			// otherwise restore the real _StackGuard, we've spoiled it in entersyscall/entersyscallblock
			_g_.Stackguard0 = _g_.Stack.Lo + StackGuard
		}
		_g_.Throwsplit = false
		return
	}

	_g_.sysexitticks = 0
	if Trace.Enabled {
		// Wait till traceGoSysBlock event is emitted.
		// This ensures consistency of the trace (the goroutine is started after it is blocked).
		for oldp != nil && oldp.Syscalltick == _g_.M.Syscalltick {
			Osyield()
		}
		// We can't trace syscall exit right now because we don't have a P.
		// Tracing code can invoke write barriers that cannot run without a P.
		// So instead we remember the syscall exit time and emit the event
		// in execute when we have a P.
		_g_.sysexitticks = Cputicks()
	}

	_g_.M.Locks--

	// Call the scheduler.
	Mcall(exitsyscall0)

	if _g_.M.Mcache == nil {
		Throw("lost mcache")
	}

	// Scheduler returned, so we're allowed to run now.
	// Delete the syscallsp information that we left for
	// the garbage collector during the system call.
	// Must wait until now because until gosched returns
	// we don't know for sure that the garbage collector
	// is not running.
	_g_.Syscallsp = 0
	_g_.M.P.Ptr().Syscalltick++
	_g_.Throwsplit = false
}

//go:nosplit
func exitsyscallfast() bool {
	_g_ := Getg()

	// Freezetheworld sets stopwait but does not retake P's.
	if Sched.Stopwait == freezeStopWait {
		_g_.M.Mcache = nil
		_g_.M.P = 0
		return false
	}

	// Try to re-acquire the last P.
	if _g_.M.P != 0 && _g_.M.P.Ptr().Status == Psyscall && Cas(&_g_.M.P.Ptr().Status, Psyscall, Prunning) {
		// There's a cpu for us, so we can run.
		_g_.M.Mcache = _g_.M.P.Ptr().Mcache
		_g_.M.P.Ptr().M.Set(_g_.M)
		if _g_.M.Syscalltick != _g_.M.P.Ptr().Syscalltick {
			if Trace.Enabled {
				// The p was retaken and then enter into syscall again (since _g_.m.syscalltick has changed).
				// traceGoSysBlock for this syscall was already emitted,
				// but here we effectively retake the p from the new syscall running on the same p.
				Systemstack(func() {
					// Denote blocking of the new syscall.
					TraceGoSysBlock(_g_.M.P.Ptr())
					// Denote completion of the current syscall.
					traceGoSysExit(0)
				})
			}
			_g_.M.P.Ptr().Syscalltick++
		}
		return true
	}

	// Try to get any other idle P.
	oldp := _g_.M.P.Ptr()
	_g_.M.Mcache = nil
	_g_.M.P = 0
	if Sched.Pidle != 0 {
		var ok bool
		Systemstack(func() {
			ok = exitsyscallfast_pidle()
			if ok && Trace.Enabled {
				if oldp != nil {
					// Wait till traceGoSysBlock event is emitted.
					// This ensures consistency of the trace (the goroutine is started after it is blocked).
					for oldp.Syscalltick == _g_.M.Syscalltick {
						Osyield()
					}
				}
				traceGoSysExit(0)
			}
		})
		if ok {
			return true
		}
	}
	return false
}

func exitsyscallfast_pidle() bool {
	Lock(&Sched.Lock)
	_p_ := Pidleget()
	if _p_ != nil && Atomicload(&Sched.Sysmonwait) != 0 {
		Atomicstore(&Sched.Sysmonwait, 0)
		Notewakeup(&Sched.Sysmonnote)
	}
	Unlock(&Sched.Lock)
	if _p_ != nil {
		Acquirep(_p_)
		return true
	}
	return false
}

// exitsyscall slow path on g0.
// Failed to acquire P, enqueue gp as runnable.
func exitsyscall0(gp *G) {
	_g_ := Getg()

	Casgstatus(gp, Gsyscall, Grunnable)
	Dropg()
	Lock(&Sched.Lock)
	_p_ := Pidleget()
	if _p_ == nil {
		Globrunqput(gp)
	} else if Atomicload(&Sched.Sysmonwait) != 0 {
		Atomicstore(&Sched.Sysmonwait, 0)
		Notewakeup(&Sched.Sysmonnote)
	}
	Unlock(&Sched.Lock)
	if _p_ != nil {
		Acquirep(_p_)
		execute(gp, false) // Never returns.
	}
	if _g_.M.Lockedg != nil {
		// Wait until another thread schedules gp and so m again.
		stoplockedm()
		execute(gp, false) // Never returns.
	}
	stopm()
	Schedule() // Never returns.
}

// Allocate a new g, with a stack big enough for stacksize bytes.
func Malg(stacksize int32) *G {
	newg := new(G)
	if stacksize >= 0 {
		stacksize = Round2(StackSystem + stacksize)
		Systemstack(func() {
			newg.Stack, newg.Stkbar = Stackalloc(uint32(stacksize))
		})
		newg.Stackguard0 = newg.Stack.Lo + StackGuard
		newg.stackguard1 = ^uintptr(0)
		newg.StackAlloc = uintptr(stacksize)
	}
	return newg
}

func mcount() int32 {
	return Sched.mcount
}

var Prof struct {
	Lock uint32
	Hz   int32
}

func _System()       { _System() }
func _ExternalCode() { _ExternalCode() }
func _GC()           { _GC() }

// Called if we receive a SIGPROF signal.
func sigprof(pc, sp, lr uintptr, gp *G, mp *M) {
	if Prof.Hz == 0 {
		return
	}

	// Profiling runs concurrently with GC, so it must not allocate.
	mp.Mallocing++

	// Define that a "user g" is a user-created goroutine, and a "system g"
	// is one that is m->g0 or m->gsignal.
	//
	// We might be interrupted for profiling halfway through a
	// goroutine switch. The switch involves updating three (or four) values:
	// g, PC, SP, and (on arm) LR. The PC must be the last to be updated,
	// because once it gets updated the new g is running.
	//
	// When switching from a user g to a system g, LR is not considered live,
	// so the update only affects g, SP, and PC. Since PC must be last, there
	// the possible partial transitions in ordinary execution are (1) g alone is updated,
	// (2) both g and SP are updated, and (3) SP alone is updated.
	// If SP or g alone is updated, we can detect the partial transition by checking
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
	// separately. If the PC is within any of the functions that does this,
	// we don't ask for a traceback. C.F. the function setsSP for more about this.
	//
	// There is another apparently viable approach, recorded here in case
	// the "PC within setsSP function" check turns out not to be usable.
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
	traceback := true
	if gp == nil || sp < gp.Stack.Lo || gp.Stack.Hi < sp || setsSP(pc) {
		traceback = false
	}
	var stk [maxCPUProfStack]uintptr
	n := 0
	if mp.Ncgo > 0 && mp.Curg != nil && mp.Curg.Syscallpc != 0 && mp.Curg.Syscallsp != 0 {
		// Cgo, we can't unwind and symbolize arbitrary C code,
		// so instead collect Go stack that leads to the cgo call.
		// This is especially important on windows, since all syscalls are cgo calls.
		n = Gentraceback(mp.Curg.Syscallpc, mp.Curg.Syscallsp, 0, mp.Curg, 0, &stk[0], len(stk), nil, nil, 0)
	} else if traceback {
		n = Gentraceback(pc, sp, lr, gp, 0, &stk[0], len(stk), nil, nil, _TraceTrap|_TraceJumpStack)
	}
	if !traceback || n <= 0 {
		// Normal traceback is impossible or has failed.
		// See if it falls into several common cases.
		n = 0
		if GOOS == "windows" && n == 0 && mp.libcallg != 0 && mp.libcallpc != 0 && mp.Libcallsp != 0 {
			// Libcall, i.e. runtime syscall on windows.
			// Collect Go stack that leads to the call.
			n = Gentraceback(mp.libcallpc, mp.Libcallsp, 0, mp.libcallg.Ptr(), 0, &stk[0], len(stk), nil, nil, 0)
		}
		if n == 0 {
			// If all of the above has failed, account it against abstract "System" or "GC".
			n = 2
			// "ExternalCode" is better than "etext".
			if pc > Firstmoduledata.etext {
				pc = FuncPC(_ExternalCode) + PCQuantum
			}
			stk[0] = pc
			if mp.Preemptoff != "" || mp.Helpgc != 0 {
				stk[1] = FuncPC(_GC) + PCQuantum
			} else {
				stk[1] = FuncPC(_System) + PCQuantum
			}
		}
	}

	if Prof.Hz != 0 {
		// Simple cas-lock to coordinate with setcpuprofilerate.
		for !Cas(&Prof.Lock, 0, 1) {
			Osyield()
		}
		if Prof.Hz != 0 {
			Cpuprof.add(stk[:n])
		}
		Atomicstore(&Prof.Lock, 0)
	}
	mp.Mallocing--
}

// Reports whether a function will set the SP
// to an absolute value. Important that
// we don't traceback when these are at the bottom
// of the stack since we can't be sure that we will
// find the caller.
//
// If the function is not on the bottom of the stack
// we assume that it will have set it up so that traceback will be consistent,
// either by being a traceback terminating function
// or putting one on the stack at the right offset.
func setsSP(pc uintptr) bool {
	f := Findfunc(pc)
	if f == nil {
		// couldn't find the function for this PC,
		// so assume the worst and stop traceback
		return true
	}
	switch f.Entry {
	case GogoPC, SystemstackPC, McallPC, MorestackPC:
		return true
	}
	return false
}

// Associate p and the current m.
func Acquirep(_p_ *P) {
	acquirep1(_p_)

	// have p; write barriers now allowed
	_g_ := Getg()
	_g_.M.Mcache = _p_.Mcache

	if Trace.Enabled {
		TraceProcStart()
	}
}

// May run during STW, so write barriers are not allowed.
//go:nowritebarrier
func acquirep1(_p_ *P) {
	_g_ := Getg()

	if _g_.M.P != 0 || _g_.M.Mcache != nil {
		Throw("acquirep: already in go")
	}
	if _p_.M != 0 || _p_.Status != Pidle {
		id := int32(0)
		if _p_.M != 0 {
			id = _p_.M.Ptr().Id
		}
		print("acquirep: p->m=", _p_.M, "(", id, ") p->status=", _p_.Status, "\n")
		Throw("acquirep: invalid p state")
	}
	_g_.M.P.Set(_p_)
	_p_.M.Set(_g_.M)
	_p_.Status = Prunning
}

// Disassociate p and the current m.
func releasep() *P {
	_g_ := Getg()

	if _g_.M.P == 0 || _g_.M.Mcache == nil {
		Throw("releasep: invalid arg")
	}
	_p_ := _g_.M.P.Ptr()
	if _p_.M.Ptr() != _g_.M || _p_.Mcache != _g_.M.Mcache || _p_.Status != Prunning {
		print("releasep: m=", _g_.M, " m->p=", _g_.M.P.Ptr(), " p->m=", _p_.M, " m->mcache=", _g_.M.Mcache, " p->mcache=", _p_.Mcache, " p->status=", _p_.Status, "\n")
		Throw("releasep: invalid p state")
	}
	if Trace.Enabled {
		TraceProcStop(_g_.M.P.Ptr())
	}
	_g_.M.P = 0
	_g_.M.Mcache = nil
	_p_.M = 0
	_p_.Status = Pidle
	return _p_
}

func Incidlelocked(v int32) {
	Lock(&Sched.Lock)
	Sched.nmidlelocked += v
	if v > 0 {
		checkdead()
	}
	Unlock(&Sched.Lock)
}

// Check for deadlock situation.
// The check is based on number of running M's, if 0 -> deadlock.
func checkdead() {
	// For -buildmode=c-shared or -buildmode=c-archive it's OK if
	// there are no running goroutines.  The calling program is
	// assumed to be running.
	if Islibrary || Isarchive {
		return
	}

	// If we are dying because of a signal caught on an already idle thread,
	// freezetheworld will cause all running threads to block.
	// And runtime will essentially enter into deadlock state,
	// except that there is a thread that will call exit soon.
	if Panicking > 0 {
		return
	}

	// -1 for sysmon
	run := Sched.mcount - Sched.Nmidle - Sched.nmidlelocked - 1
	if run > 0 {
		return
	}
	if run < 0 {
		print("runtime: checkdead: nmidle=", Sched.Nmidle, " nmidlelocked=", Sched.nmidlelocked, " mcount=", Sched.mcount, "\n")
		Throw("checkdead: inconsistent counts")
	}

	grunning := 0
	Lock(&Allglock)
	for i := 0; i < len(Allgs); i++ {
		gp := Allgs[i]
		if IsSystemGoroutine(gp) {
			continue
		}
		s := Readgstatus(gp)
		switch s &^ Gscan {
		case Gwaiting:
			grunning++
		case Grunnable,
			Grunning,
			Gsyscall:
			Unlock(&Allglock)
			print("runtime: checkdead: find g ", gp.Goid, " in status ", s, "\n")
			Throw("checkdead: runnable g")
		}
	}
	Unlock(&Allglock)
	if grunning == 0 { // possible if main goroutine calls runtime·Goexit()
		Throw("no goroutines (main called runtime.Goexit) - deadlock!")
	}

	// Maybe jump time forward for playground.
	gp := timejump()
	if gp != nil {
		Casgstatus(gp, Gwaiting, Grunnable)
		Globrunqput(gp)
		_p_ := Pidleget()
		if _p_ == nil {
			Throw("checkdead: no p for timer")
		}
		mp := Mget()
		if mp == nil {
			Newm(nil, _p_)
		} else {
			mp.Nextp.Set(_p_)
			Notewakeup(&mp.Park)
		}
		return
	}

	Getg().M.Throwing = -1 // do not dump full stacks
	Throw("all goroutines are asleep - deadlock!")
}

// Tell all goroutines that they have been preempted and they should stop.
// This function is purely best-effort.  It can fail to inform a goroutine if a
// processor just started running it.
// No locks need to be held.
// Returns true if preemption request was issued to at least one goroutine.
func Preemptall() bool {
	res := false
	for i := int32(0); i < Gomaxprocs; i++ {
		_p_ := Allp[i]
		if _p_ == nil || _p_.Status != Prunning {
			continue
		}
		if Preemptone(_p_) {
			res = true
		}
	}
	return res
}

// Tell the goroutine running on processor P to stop.
// This function is purely best-effort.  It can incorrectly fail to inform the
// goroutine.  It can send inform the wrong goroutine.  Even if it informs the
// correct goroutine, that goroutine might ignore the request if it is
// simultaneously executing newstack.
// No lock needs to be held.
// Returns true if preemption request was issued.
// The actual preemption will happen at some point in the future
// and will be indicated by the gp->status no longer being
// Grunning
func Preemptone(_p_ *P) bool {
	mp := _p_.M.Ptr()
	if mp == nil || mp == Getg().M {
		return false
	}
	gp := mp.Curg
	if gp == nil || gp == mp.G0 {
		return false
	}

	gp.Preempt = true

	// Every call in a go routine checks for stack overflow by
	// comparing the current stack pointer to gp->stackguard0.
	// Setting gp->stackguard0 to StackPreempt folds
	// preemption into the normal stack overflow check.
	gp.Stackguard0 = StackPreempt
	return true
}

var starttime int64

func Schedtrace(detailed bool) {
	now := Nanotime()
	if starttime == 0 {
		starttime = now
	}

	Lock(&Sched.Lock)
	print("SCHED ", (now-starttime)/1e6, "ms: gomaxprocs=", Gomaxprocs, " idleprocs=", Sched.Npidle, " threads=", Sched.mcount, " spinningthreads=", Sched.Nmspinning, " idlethreads=", Sched.Nmidle, " runqueue=", Sched.Runqsize)
	if detailed {
		print(" gcwaiting=", Sched.Gcwaiting, " nmidlelocked=", Sched.nmidlelocked, " stopwait=", Sched.Stopwait, " sysmonwait=", Sched.Sysmonwait, "\n")
	}
	// We must be careful while reading data from P's, M's and G's.
	// Even if we hold schedlock, most data can be changed concurrently.
	// E.g. (p->m ? p->m->id : -1) can crash if p->m changes from non-nil to nil.
	for i := int32(0); i < Gomaxprocs; i++ {
		_p_ := Allp[i]
		if _p_ == nil {
			continue
		}
		mp := _p_.M.Ptr()
		h := Atomicload(&_p_.Runqhead)
		t := Atomicload(&_p_.Runqtail)
		if detailed {
			id := int32(-1)
			if mp != nil {
				id = mp.Id
			}
			print("  P", i, ": status=", _p_.Status, " schedtick=", _p_.Schedtick, " syscalltick=", _p_.Syscalltick, " m=", id, " runqsize=", t-h, " gfreecnt=", _p_.Gfreecnt, "\n")
		} else {
			// In non-detailed mode format lengths of per-P run queues as:
			// [len1 len2 len3 len4]
			print(" ")
			if i == 0 {
				print("[")
			}
			print(t - h)
			if i == Gomaxprocs-1 {
				print("]\n")
			}
		}
	}

	if !detailed {
		Unlock(&Sched.Lock)
		return
	}

	for mp := Allm; mp != nil; mp = mp.Alllink {
		_p_ := mp.P.Ptr()
		gp := mp.Curg
		lockedg := mp.Lockedg
		id1 := int32(-1)
		if _p_ != nil {
			id1 = _p_.Id
		}
		id2 := int64(-1)
		if gp != nil {
			id2 = gp.Goid
		}
		id3 := int64(-1)
		if lockedg != nil {
			id3 = lockedg.Goid
		}
		print("  M", mp.Id, ": p=", id1, " curg=", id2, " mallocing=", mp.Mallocing, " throwing=", mp.Throwing, " preemptoff=", mp.Preemptoff, ""+" locks=", mp.Locks, " dying=", mp.dying, " helpgc=", mp.Helpgc, " spinning=", mp.spinning, " blocked=", Getg().M.blocked, " lockedg=", id3, "\n")
	}

	Lock(&Allglock)
	for gi := 0; gi < len(Allgs); gi++ {
		gp := Allgs[gi]
		mp := gp.M
		lockedm := gp.Lockedm
		id1 := int32(-1)
		if mp != nil {
			id1 = mp.Id
		}
		id2 := int32(-1)
		if lockedm != nil {
			id2 = lockedm.Id
		}
		print("  G", gp.Goid, ": status=", Readgstatus(gp), "(", gp.Waitreason, ") m=", id1, " lockedm=", id2, "\n")
	}
	Unlock(&Allglock)
	Unlock(&Sched.Lock)
}

// Put mp on midle list.
// Sched must be locked.
// May run during STW, so write barriers are not allowed.
//go:nowritebarrier
func mput(mp *M) {
	mp.Schedlink = Sched.midle
	Sched.midle.Set(mp)
	Sched.Nmidle++
	checkdead()
}

// Try to get an m from midle list.
// Sched must be locked.
// May run during STW, so write barriers are not allowed.
//go:nowritebarrier
func Mget() *M {
	mp := Sched.midle.Ptr()
	if mp != nil {
		Sched.midle = mp.Schedlink
		Sched.Nmidle--
	}
	return mp
}

// Put gp on the global runnable queue.
// Sched must be locked.
// May run during STW, so write barriers are not allowed.
//go:nowritebarrier
func Globrunqput(gp *G) {
	gp.Schedlink = 0
	if Sched.Runqtail != 0 {
		Sched.Runqtail.Ptr().Schedlink.Set(gp)
	} else {
		Sched.Runqhead.Set(gp)
	}
	Sched.Runqtail.Set(gp)
	Sched.Runqsize++
}

// Put a batch of runnable goroutines on the global runnable queue.
// Sched must be locked.
func globrunqputbatch(ghead *G, gtail *G, n int32) {
	gtail.Schedlink = 0
	if Sched.Runqtail != 0 {
		Sched.Runqtail.Ptr().Schedlink.Set(ghead)
	} else {
		Sched.Runqhead.Set(ghead)
	}
	Sched.Runqtail.Set(gtail)
	Sched.Runqsize += n
}

// Try get a batch of G's from the global runnable queue.
// Sched must be locked.
func globrunqget(_p_ *P, max int32) *G {
	if Sched.Runqsize == 0 {
		return nil
	}

	n := Sched.Runqsize/Gomaxprocs + 1
	if n > Sched.Runqsize {
		n = Sched.Runqsize
	}
	if max > 0 && n > max {
		n = max
	}
	if n > int32(len(_p_.Runq))/2 {
		n = int32(len(_p_.Runq)) / 2
	}

	Sched.Runqsize -= n
	if Sched.Runqsize == 0 {
		Sched.Runqtail = 0
	}

	gp := Sched.Runqhead.Ptr()
	Sched.Runqhead = gp.Schedlink
	n--
	for ; n > 0; n-- {
		gp1 := Sched.Runqhead.Ptr()
		Sched.Runqhead = gp1.Schedlink
		Runqput(_p_, gp1, false)
	}
	return gp
}

// Put p to on _Pidle list.
// Sched must be locked.
// May run during STW, so write barriers are not allowed.
//go:nowritebarrier
func Pidleput(_p_ *P) {
	if !Runqempty(_p_) {
		Throw("pidleput: P has non-empty run queue")
	}
	_p_.Link = Sched.Pidle
	Sched.Pidle.Set(_p_)
	Xadd(&Sched.Npidle, 1) // TODO: fast atomic
}

// Try get a p from _Pidle list.
// Sched must be locked.
// May run during STW, so write barriers are not allowed.
//go:nowritebarrier
func Pidleget() *P {
	_p_ := Sched.Pidle.Ptr()
	if _p_ != nil {
		Sched.Pidle = _p_.Link
		Xadd(&Sched.Npidle, -1) // TODO: fast atomic
	}
	return _p_
}

// runqempty returns true if _p_ has no Gs on its local run queue.
// Note that this test is generally racy.
func Runqempty(_p_ *P) bool {
	return _p_.Runqhead == _p_.Runqtail && _p_.Runnext == 0
}

// To shake out latent assumptions about scheduling order,
// we introduce some randomness into scheduling decisions
// when running with the race detector.
// The need for this was made obvious by changing the
// (deterministic) scheduling order in Go 1.5 and breaking
// many poorly-written tests.
// With the randomness here, as long as the tests pass
// consistently with -race, they shouldn't have latent scheduling
// assumptions.
const randomizeScheduler = Raceenabled

// runqput tries to put g on the local runnable queue.
// If next if false, runqput adds g to the tail of the runnable queue.
// If next is true, runqput puts g in the _p_.runnext slot.
// If the run queue is full, runnext puts g on the global queue.
// Executed only by the owner P.
func Runqput(_p_ *P, gp *G, next bool) {
	if randomizeScheduler && next && Fastrand1()%2 == 0 {
		next = false
	}

	if next {
	retryNext:
		oldnext := _p_.Runnext
		if !_p_.Runnext.cas(oldnext, Guintptr(unsafe.Pointer(gp))) {
			goto retryNext
		}
		if oldnext == 0 {
			return
		}
		// Kick the old runnext out to the regular run queue.
		gp = oldnext.Ptr()
	}

retry:
	h := Atomicload(&_p_.Runqhead) // load-acquire, synchronize with consumers
	t := _p_.Runqtail
	if t-h < uint32(len(_p_.Runq)) {
		_p_.Runq[t%uint32(len(_p_.Runq))] = gp
		Atomicstore(&_p_.Runqtail, t+1) // store-release, makes the item available for consumption
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
func runqputslow(_p_ *P, gp *G, h, t uint32) bool {
	var batch [len(_p_.Runq)/2 + 1]*G

	// First, grab a batch from local queue.
	n := t - h
	n = n / 2
	if n != uint32(len(_p_.Runq)/2) {
		Throw("runqputslow: queue is not full")
	}
	for i := uint32(0); i < n; i++ {
		batch[i] = _p_.Runq[(h+i)%uint32(len(_p_.Runq))]
	}
	if !Cas(&_p_.Runqhead, h, h+n) { // cas-release, commits consume
		return false
	}
	batch[n] = gp

	if randomizeScheduler {
		for i := uint32(1); i <= n; i++ {
			j := Fastrand1() % (i + 1)
			batch[i], batch[j] = batch[j], batch[i]
		}
	}

	// Link the goroutines.
	for i := uint32(0); i < n; i++ {
		batch[i].Schedlink.Set(batch[i+1])
	}

	// Now put the batch on global queue.
	Lock(&Sched.Lock)
	globrunqputbatch(batch[0], batch[n], int32(n+1))
	Unlock(&Sched.Lock)
	return true
}

// Get g from local runnable queue.
// If inheritTime is true, gp should inherit the remaining time in the
// current time slice. Otherwise, it should start a new time slice.
// Executed only by the owner P.
func Runqget(_p_ *P) (gp *G, inheritTime bool) {
	// If there's a runnext, it's the next G to run.
	for {
		next := _p_.Runnext
		if next == 0 {
			break
		}
		if _p_.Runnext.cas(next, 0) {
			return next.Ptr(), true
		}
	}

	for {
		h := Atomicload(&_p_.Runqhead) // load-acquire, synchronize with other consumers
		t := _p_.Runqtail
		if t == h {
			return nil, false
		}
		gp := _p_.Runq[h%uint32(len(_p_.Runq))]
		if Cas(&_p_.Runqhead, h, h+1) { // cas-release, commits consume
			return gp, false
		}
	}
}

// Grabs a batch of goroutines from _p_'s runnable queue into batch.
// Batch is a ring buffer starting at batchHead.
// Returns number of grabbed goroutines.
// Can be executed by any P.
func runqgrab(_p_ *P, batch *[256]*G, batchHead uint32, stealRunNextG bool) uint32 {
	for {
		h := Atomicload(&_p_.Runqhead) // load-acquire, synchronize with other consumers
		t := Atomicload(&_p_.Runqtail) // load-acquire, synchronize with the producer
		n := t - h
		n = n - n/2
		if n == 0 {
			if stealRunNextG {
				// Try to steal from _p_.runnext.
				if next := _p_.Runnext; next != 0 {
					// Sleep to ensure that _p_ isn't about to run the g we
					// are about to steal.
					// The important use case here is when the g running on _p_
					// ready()s another g and then almost immediately blocks.
					// Instead of stealing runnext in this window, back off
					// to give _p_ a chance to schedule runnext. This will avoid
					// thrashing gs between different Ps.
					Usleep(100)
					if !_p_.Runnext.cas(next, 0) {
						continue
					}
					batch[batchHead%uint32(len(batch))] = next.Ptr()
					return 1
				}
			}
			return 0
		}
		if n > uint32(len(_p_.Runq)/2) { // read inconsistent h and t
			continue
		}
		for i := uint32(0); i < n; i++ {
			g := _p_.Runq[(h+i)%uint32(len(_p_.Runq))]
			batch[(batchHead+i)%uint32(len(batch))] = g
		}
		if Cas(&_p_.Runqhead, h, h+n) { // cas-release, commits consume
			return n
		}
	}
}

// Steal half of elements from local runnable queue of p2
// and put onto local runnable queue of p.
// Returns one of the stolen elements (or nil if failed).
func Runqsteal(_p_, p2 *P, stealRunNextG bool) *G {
	t := _p_.Runqtail
	n := runqgrab(p2, &_p_.Runq, t, stealRunNextG)
	if n == 0 {
		return nil
	}
	n--
	gp := _p_.Runq[(t+n)%uint32(len(_p_.Runq))]
	if n == 0 {
		return gp
	}
	h := Atomicload(&_p_.Runqhead) // load-acquire, synchronize with consumers
	if t-h+n >= uint32(len(_p_.Runq)) {
		Throw("runqsteal: runq overflow")
	}
	Atomicstore(&_p_.Runqtail, t+n) // store-release, makes the item available for consumption
	return gp
}
