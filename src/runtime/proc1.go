// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	_cgo "runtime/internal/cgo"
	_core "runtime/internal/core"
	_gc "runtime/internal/gc"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
	_schedinit "runtime/internal/schedinit"
	"unsafe"
)

// Goroutine scheduler
// The scheduler's job is to distribute ready-to-run goroutines over worker threads.
//
// The main concepts are:
// G - goroutine.
// M - worker thread, or machine.
// P - processor, a resource that is required to execute Go code.
//     M must have an associated P to execute Go code, however it can be
//     blocked or in a syscall w/o an associated P.
//
// Design doc at http://golang.org/s/go11sched.

const (
	// Number of goroutine ids to grab from sched.goidgen to local per-P cache at once.
	// 16 seems to provide enough amortization, but other than that it's mostly arbitrary number.
	_GoidCacheBatch = 16
)

func newsysmon() {
	_sched.Newm(sysmon, nil)
}

func isscanstatus(status uint32) bool {
	if status == _lock.Gscan {
		_lock.Throw("isscanstatus: Bad status Gscan")
	}
	return status&_lock.Gscan == _lock.Gscan
}

func stopscanstart(gp *_core.G) {
	_g_ := _core.Getg()
	if _g_ == gp {
		_lock.Throw("GC not moved to G0")
	}
	if _gc.Stopg(gp) {
		if !isscanstatus(_lock.Readgstatus(gp)) {
			_sched.Dumpgstatus(gp)
			_lock.Throw("GC not in scan state")
		}
		_gc.Restartg(gp)
	}
}

// Runs on g0 and does the actual work after putting the g back on the run queue.
func mquiesce(gpmaster *_core.G) {
	// enqueue the calling goroutine.
	_gc.Restartg(gpmaster)

	activeglen := len(_lock.Allgs)
	for i := 0; i < activeglen; i++ {
		gp := _lock.Allgs[i]
		if _lock.Readgstatus(gp) == _lock.Gdead {
			gp.Gcworkdone = true // noop scan.
		} else {
			gp.Gcworkdone = false
		}
		stopscanstart(gp)
	}

	// Check that the G's gcwork (such as scanning) has been done. If not do it now.
	// You can end up doing work here if the page trap on a Grunning Goroutine has
	// not been sprung or in some race situations. For example a runnable goes dead
	// and is started up again with a gp->gcworkdone set to false.
	for i := 0; i < activeglen; i++ {
		gp := _lock.Allgs[i]
		for !gp.Gcworkdone {
			status := _lock.Readgstatus(gp)
			if status == _lock.Gdead {
				//do nothing, scan not needed.
				gp.Gcworkdone = true // scan is a noop
				break
			}
			if status == _lock.Grunning && gp.Stackguard0 == uintptr(_lock.StackPreempt) && _gc.Notetsleep(&_core.Sched.Stopnote, 100*1000) { // nanosecond arg
				_sched.Noteclear(&_core.Sched.Stopnote)
			} else {
				stopscanstart(gp)
			}
		}
	}

	for i := 0; i < activeglen; i++ {
		gp := _lock.Allgs[i]
		status := _lock.Readgstatus(gp)
		if isscanstatus(status) {
			print("mstopandscang:bottom: post scan bad status gp=", gp, " has status ", _core.Hex(status), "\n")
			_sched.Dumpgstatus(gp)
		}
		if !gp.Gcworkdone && status != _lock.Gdead {
			print("mstopandscang:bottom: post scan gp=", gp, "->gcworkdone still false\n")
			_sched.Dumpgstatus(gp)
		}
	}

	_sched.Schedule() // Never returns.
}

// quiesce moves all the goroutines to a GC safepoint which for now is a at preemption point.
// If the global gcphase is GCmark quiesce will ensure that all of the goroutine's stacks
// have been scanned before it returns.
func quiesce(mastergp *_core.G) {
	_gc.Castogscanstatus(mastergp, _lock.Grunning, _lock.Gscanenqueue)
	// Now move this to the g0 (aka m) stack.
	// g0 will potentially scan this thread and put mastergp on the runqueue
	_sched.Mcall(mquiesce)
}

// When running with cgo, we call _cgo_thread_start
// to start threads for us so that we can play nicely with
// foreign code.
var cgoThreadStart unsafe.Pointer

// dropm is called when a cgo callback has called needm but is now
// done with the callback and returning back into the non-Go thread.
// It puts the current m back onto the extra list.
//
// The main expense here is the call to signalstack to release the
// m's signal stack, and then the call to needm on the next callback
// from this thread. It is tempting to try to save the m for next time,
// which would eliminate both these costs, but there might not be
// a next time: the current thread (which Go does not control) might exit.
// If we saved the m for that thread, there would be an m leak each time
// such a thread exited. Instead, we acquire and release an m on each
// call. These should typically not be scheduling operations, just a few
// atomics, so the cost should be small.
//
// TODO(rsc): An alternative would be to allocate a dummy pthread per-thread
// variable using pthread_key_create. Unlike the pthread keys we already use
// on OS X, this dummy key would never be read by Go code. It would exist
// only so that we could register at thread-exit-time destructor.
// That destructor would put the m back onto the extra list.
// This is purely a performance optimization. The current version,
// in which dropm happens on each cgo call, is still correct too.
// We may have to keep the current version on systems with cgo
// but without pthreads, like Windows.
func dropm() {
	// Undo whatever initialization minit did during needm.
	unminit()

	// Clear m and g, and return m to the extra list.
	// After the call to setg we can only call nosplit functions
	// with no pointer manipulation.
	mp := _core.Getg().M
	mnext := _core.Lockextra(true)
	mp.Schedlink = mnext

	_core.Setg(nil)
	_core.Unlockextra(mp)
}

// Puts the current goroutine into a waiting state and calls unlockf.
// If unlockf returns false, the goroutine is resumed.
func park(unlockf func(*_core.G, unsafe.Pointer) bool, lock unsafe.Pointer, reason string, traceev byte) {
	_g_ := _core.Getg()

	_g_.M.Waitlock = lock
	_g_.M.Waitunlockf = *(*unsafe.Pointer)(unsafe.Pointer(&unlockf))
	_g_.M.Waittraceev = traceev
	_g_.Waitreason = reason
	_sched.Mcall(_sched.Park_m)
}

// Puts the current goroutine into a waiting state and unlocks the lock.
// The goroutine can be made runnable again by calling ready(gp).
func parkunlock(lock *_core.Mutex, reason string, traceev byte) {
	park(_sched.Parkunlock_c, unsafe.Pointer(lock), reason, traceev)
}

func gopreempt_m(gp *_core.G) {
	if _sched.Trace.Enabled {
		traceGoPreempt()
	}
	_sched.GoschedImpl(gp)
}

// Finishes execution of the current goroutine.
// Must be NOSPLIT because it is called from Go. (TODO - probably not anymore)
//go:nosplit
func goexit1() {
	if _sched.Raceenabled {
		racegoend()
	}
	if _sched.Trace.Enabled {
		traceGoEnd()
	}
	_sched.Mcall(goexit0)
}

// goexit continuation on g0.
func goexit0(gp *_core.G) {
	_g_ := _core.Getg()

	_sched.Casgstatus(gp, _lock.Grunning, _lock.Gdead)
	gp.M = nil
	gp.Lockedm = nil
	_g_.M.Lockedg = nil
	gp.Paniconfault = false
	gp.Defer = nil // should be true already but just in case.
	gp.Panic = nil // non-nil for Goexit during panic. points at stack-allocated data.
	gp.Writebuf = nil
	gp.Waitreason = ""
	gp.Param = nil

	_sched.Dropg()

	if _g_.M.Locked&^_cgo.LockExternal != 0 {
		print("invalid m->locked = ", _g_.M.Locked, "\n")
		_lock.Throw("internal lockOSThread error")
	}
	_g_.M.Locked = 0
	gfput(_g_.M.P, gp)
	_sched.Schedule()
}

func beforefork() {
	gp := _core.Getg().M.Curg

	// Fork can hang if preempted with signals frequently enough (see issue 5517).
	// Ensure that we stay on the same M where we disable profiling.
	gp.M.Locks++
	if gp.M.Profilehz != 0 {
		_sched.Resetcpuprofiler(0)
	}

	// This function is called before fork in syscall package.
	// Code between fork and exec must not allocate memory nor even try to grow stack.
	// Here we spoil g->_StackGuard to reliably detect any attempts to grow stack.
	// runtime_AfterFork will undo this in parent process, but not in child.
	gp.Stackguard0 = _lock.StackFork
}

// Called from syscall package before fork.
//go:linkname syscall_runtime_BeforeFork syscall.runtime_BeforeFork
//go:nosplit
func syscall_runtime_BeforeFork() {
	_lock.Systemstack(beforefork)
}

func afterfork() {
	gp := _core.Getg().M.Curg

	// See the comment in beforefork.
	gp.Stackguard0 = gp.Stack.Lo + _core.StackGuard

	hz := _core.Sched.Profilehz
	if hz != 0 {
		_sched.Resetcpuprofiler(hz)
	}
	gp.M.Locks--
}

// Called from syscall package after fork in parent.
//go:linkname syscall_runtime_AfterFork syscall.runtime_AfterFork
//go:nosplit
func syscall_runtime_AfterFork() {
	_lock.Systemstack(afterfork)
}

// Create a new g running fn with siz bytes of arguments.
// Put it on the queue of g's waiting to run.
// The compiler turns a go statement into a call to this.
// Cannot split the stack because it assumes that the arguments
// are available sequentially after &fn; they would not be
// copied if a stack split occurred.
//go:nosplit
func newproc(siz int32, fn *_core.Funcval) {
	argp := _core.Add(unsafe.Pointer(&fn), _core.PtrSize)
	pc := _lock.Getcallerpc(unsafe.Pointer(&siz))
	_lock.Systemstack(func() {
		newproc1(fn, (*uint8)(argp), siz, 0, pc)
	})
}

// Create a new g running fn with narg bytes of arguments starting
// at argp and returning nret bytes of results.  callerpc is the
// address of the go statement that created this.  The new g is put
// on the queue of g's waiting to run.
func newproc1(fn *_core.Funcval, argp *uint8, narg int32, nret int32, callerpc uintptr) *_core.G {
	_g_ := _core.Getg()

	if fn == nil {
		_g_.M.Throwing = -1 // do not dump full stacks
		_lock.Throw("go of nil func value")
	}
	_g_.M.Locks++ // disable preemption because it can be holding p in a local var
	siz := narg + nret
	siz = (siz + 7) &^ 7

	// We could allocate a larger initial stack if necessary.
	// Not worth it: this is almost always an error.
	// 4*sizeof(uintreg): extra space added below
	// sizeof(uintreg): caller's LR (arm) or return address (x86, in gostartcall).
	if siz >= _core.StackMin-4*_lock.RegSize-_lock.RegSize {
		_lock.Throw("newproc: function arguments too large for new goroutine")
	}

	_p_ := _g_.M.P
	newg := gfget(_p_)
	if newg == nil {
		newg = _sched.Malg(_core.StackMin)
		_sched.Casgstatus(newg, _lock.Gidle, _lock.Gdead)
		_cgo.Allgadd(newg) // publishes with a g->status of Gdead so GC scanner doesn't look at uninitialized stack.
	}
	if newg.Stack.Hi == 0 {
		_lock.Throw("newproc1: newg missing stack")
	}

	if _lock.Readgstatus(newg) != _lock.Gdead {
		_lock.Throw("newproc1: new g is not Gdead")
	}

	sp := newg.Stack.Hi
	sp -= 4 * _lock.RegSize // extra space in case of reads slightly beyond frame
	sp -= uintptr(siz)
	_sched.Memmove(unsafe.Pointer(sp), unsafe.Pointer(argp), uintptr(narg))
	if hasLinkRegister {
		// caller's LR
		sp -= _core.PtrSize
		*(*unsafe.Pointer)(unsafe.Pointer(sp)) = nil
	}

	_core.Memclr(unsafe.Pointer(&newg.Sched), unsafe.Sizeof(newg.Sched))
	newg.Sched.Sp = sp
	newg.Sched.Pc = _lock.FuncPC(_schedinit.Goexit) + _lock.PCQuantum // +PCQuantum so that previous instruction is in same function
	newg.Sched.G = _core.Guintptr(unsafe.Pointer(newg))
	gostartcallfn(&newg.Sched, fn)
	newg.Gopc = callerpc
	newg.Startpc = fn.Fn
	_sched.Casgstatus(newg, _lock.Gdead, _lock.Grunnable)

	if _p_.Goidcache == _p_.Goidcacheend {
		// Sched.goidgen is the last allocated id,
		// this batch must be [sched.goidgen+1, sched.goidgen+GoidCacheBatch].
		// At startup sched.goidgen=0, so main goroutine receives goid=1.
		_p_.Goidcache = _lock.Xadd64(&_core.Sched.Goidgen, _GoidCacheBatch)
		_p_.Goidcache -= _GoidCacheBatch - 1
		_p_.Goidcacheend = _p_.Goidcache + _GoidCacheBatch
	}
	newg.Goid = int64(_p_.Goidcache)
	_p_.Goidcache++
	if _sched.Raceenabled {
		newg.Racectx = _cgo.Racegostart(callerpc)
	}
	if _sched.Trace.Enabled {
		traceGoCreate(newg, newg.Startpc)
	}
	_sched.Runqput(_p_, newg)

	if _lock.Atomicload(&_core.Sched.Npidle) != 0 && _lock.Atomicload(&_core.Sched.Nmspinning) == 0 && unsafe.Pointer(fn.Fn) != unsafe.Pointer(_lock.FuncPC(main)) { // TODO: fast atomic
		_sched.Wakep()
	}
	_g_.M.Locks--
	if _g_.M.Locks == 0 && _g_.Preempt { // restore the preemption request in case we've cleared it in newstack
		_g_.Stackguard0 = _lock.StackPreempt
	}
	return newg
}

// Put on gfree list.
// If local list is too long, transfer a batch to the global list.
func gfput(_p_ *_core.P, gp *_core.G) {
	if _lock.Readgstatus(gp) != _lock.Gdead {
		_lock.Throw("gfput: bad status (not Gdead)")
	}

	stksize := gp.Stack.Hi - gp.Stack.Lo

	if stksize != _core.FixedStack {
		// non-standard stack size - free it.
		_gc.Stackfree(gp.Stack)
		gp.Stack.Lo = 0
		gp.Stack.Hi = 0
		gp.Stackguard0 = 0
	}

	gp.Schedlink = _p_.Gfree
	_p_.Gfree = gp
	_p_.Gfreecnt++
	if _p_.Gfreecnt >= 64 {
		_lock.Lock(&_core.Sched.Gflock)
		for _p_.Gfreecnt >= 32 {
			_p_.Gfreecnt--
			gp = _p_.Gfree
			_p_.Gfree = gp.Schedlink
			gp.Schedlink = _core.Sched.Gfree
			_core.Sched.Gfree = gp
			_core.Sched.Ngfree++
		}
		_lock.Unlock(&_core.Sched.Gflock)
	}
}

// Get from gfree list.
// If local list is empty, grab a batch from global list.
func gfget(_p_ *_core.P) *_core.G {
retry:
	gp := _p_.Gfree
	if gp == nil && _core.Sched.Gfree != nil {
		_lock.Lock(&_core.Sched.Gflock)
		for _p_.Gfreecnt < 32 && _core.Sched.Gfree != nil {
			_p_.Gfreecnt++
			gp = _core.Sched.Gfree
			_core.Sched.Gfree = gp.Schedlink
			_core.Sched.Ngfree--
			gp.Schedlink = _p_.Gfree
			_p_.Gfree = gp
		}
		_lock.Unlock(&_core.Sched.Gflock)
		goto retry
	}
	if gp != nil {
		_p_.Gfree = gp.Schedlink
		_p_.Gfreecnt--
		if gp.Stack.Lo == 0 {
			// Stack was deallocated in gfput.  Allocate a new one.
			_lock.Systemstack(func() {
				gp.Stack = _sched.Stackalloc(_core.FixedStack)
			})
			gp.Stackguard0 = gp.Stack.Lo + _core.StackGuard
		} else {
			if _sched.Raceenabled {
				_sched.Racemalloc(unsafe.Pointer(gp.Stack.Lo), gp.Stack.Hi-gp.Stack.Lo)
			}
		}
	}
	return gp
}

// Breakpoint executes a breakpoint trap.
func Breakpoint() {
	breakpoint()
}

//go:nosplit

// LockOSThread wires the calling goroutine to its current operating system thread.
// Until the calling goroutine exits or calls UnlockOSThread, it will always
// execute in that thread, and no other goroutine can.
func LockOSThread() {
	_core.Getg().M.Locked |= _cgo.LockExternal
	_cgo.DolockOSThread()
}

//go:nosplit

// UnlockOSThread unwires the calling goroutine from its fixed operating system thread.
// If the calling goroutine has not called LockOSThread, UnlockOSThread is a no-op.
func UnlockOSThread() {
	_core.Getg().M.Locked &^= _cgo.LockExternal
	_cgo.DounlockOSThread()
}

func sysmon() {
	// If we go two minutes without a garbage collection, force one to run.
	forcegcperiod := int64(2 * 60 * 1e9)

	// If a heap span goes unused for 5 minutes after a garbage collection,
	// we hand it back to the operating system.
	scavengelimit := int64(5 * 60 * 1e9)

	if _lock.Debug.Scavenge > 0 {
		// Scavenge-a-lot for testing.
		forcegcperiod = 10 * 1e6
		scavengelimit = 20 * 1e6
	}

	lastscavenge := _lock.Nanotime()
	nscavenge := 0

	// Make wake-up period small enough for the sampling to be correct.
	maxsleep := forcegcperiod / 2
	if scavengelimit < forcegcperiod {
		maxsleep = scavengelimit / 2
	}

	lasttrace := int64(0)
	idle := 0 // how many cycles in succession we had not wokeup somebody
	delay := uint32(0)
	for {
		if idle == 0 { // start with 20us sleep...
			delay = 20
		} else if idle > 50 { // start doubling the sleep after 1ms...
			delay *= 2
		}
		if delay > 10*1000 { // up to 10ms
			delay = 10 * 1000
		}
		_core.Usleep(delay)
		if _lock.Debug.Schedtrace <= 0 && (_core.Sched.Gcwaiting != 0 || _lock.Atomicload(&_core.Sched.Npidle) == uint32(_lock.Gomaxprocs)) { // TODO: fast atomic
			_lock.Lock(&_core.Sched.Lock)
			if _lock.Atomicload(&_core.Sched.Gcwaiting) != 0 || _lock.Atomicload(&_core.Sched.Npidle) == uint32(_lock.Gomaxprocs) {
				_lock.Atomicstore(&_core.Sched.Sysmonwait, 1)
				_lock.Unlock(&_core.Sched.Lock)
				_gc.Notetsleep(&_core.Sched.Sysmonnote, maxsleep)
				_lock.Lock(&_core.Sched.Lock)
				_lock.Atomicstore(&_core.Sched.Sysmonwait, 0)
				_sched.Noteclear(&_core.Sched.Sysmonnote)
				idle = 0
				delay = 20
			}
			_lock.Unlock(&_core.Sched.Lock)
		}
		// poll network if not polled for more than 10ms
		lastpoll := int64(_sched.Atomicload64(&_core.Sched.Lastpoll))
		now := _lock.Nanotime()
		unixnow := _gc.Unixnanotime()
		if lastpoll != 0 && lastpoll+10*1000*1000 < now {
			_sched.Cas64(&_core.Sched.Lastpoll, uint64(lastpoll), uint64(now))
			gp := _sched.Netpoll(false) // non-blocking - returns list of goroutines
			if gp != nil {
				// Need to decrement number of idle locked M's
				// (pretending that one more is running) before injectglist.
				// Otherwise it can lead to the following situation:
				// injectglist grabs all P's but before it starts M's to run the P's,
				// another M returns from syscall, finishes running its G,
				// observes that there is no work to do and no other running M's
				// and reports deadlock.
				_sched.Incidlelocked(-1)
				_sched.Injectglist(gp)
				_sched.Incidlelocked(1)
			}
		}
		// retake P's blocked in syscalls
		// and preempt long running G's
		if retake(now) != 0 {
			idle = 0
		} else {
			idle++
		}
		// check if we need to force a GC
		lastgc := int64(_sched.Atomicload64(&_lock.Memstats.Last_gc))
		if lastgc != 0 && unixnow-lastgc > forcegcperiod && _lock.Atomicload(&forcegc.idle) != 0 {
			_lock.Lock(&forcegc.lock)
			forcegc.idle = 0
			forcegc.g.Schedlink = nil
			_sched.Injectglist(forcegc.g)
			_lock.Unlock(&forcegc.lock)
		}
		// scavenge heap once in a while
		if lastscavenge+scavengelimit/2 < now {
			mHeap_Scavenge(int32(nscavenge), uint64(now), uint64(scavengelimit))
			lastscavenge = now
			nscavenge++
		}
		if _lock.Debug.Schedtrace > 0 && lasttrace+int64(_lock.Debug.Schedtrace*1000000) <= now {
			lasttrace = now
			_lock.Schedtrace(_lock.Debug.Scheddetail > 0)
		}
	}
}

var pdesc [_lock.MaxGomaxprocs]struct {
	schedtick   uint32
	schedwhen   int64
	syscalltick uint32
	syscallwhen int64
}

func retake(now int64) uint32 {
	n := 0
	for i := int32(0); i < _lock.Gomaxprocs; i++ {
		_p_ := _lock.Allp[i]
		if _p_ == nil {
			continue
		}
		pd := &pdesc[i]
		s := _p_.Status
		if s == _lock.Psyscall {
			// Retake P from syscall if it's there for more than 1 sysmon tick (at least 20us).
			t := int64(_p_.Syscalltick)
			if int64(pd.syscalltick) != t {
				pd.syscalltick = uint32(t)
				pd.syscallwhen = now
				continue
			}
			// On the one hand we don't want to retake Ps if there is no other work to do,
			// but on the other hand we want to retake them eventually
			// because they can prevent the sysmon thread from deep sleep.
			if _p_.Runqhead == _p_.Runqtail && _lock.Atomicload(&_core.Sched.Nmspinning)+_lock.Atomicload(&_core.Sched.Npidle) > 0 && pd.syscallwhen+10*1000*1000 > now {
				continue
			}
			// Need to decrement number of idle locked M's
			// (pretending that one more is running) before the CAS.
			// Otherwise the M from which we retake can exit the syscall,
			// increment nmidle and report deadlock.
			_sched.Incidlelocked(-1)
			if _sched.Cas(&_p_.Status, s, _lock.Pidle) {
				if _sched.Trace.Enabled {
					_sched.TraceGoSysBlock(_p_)
					_sched.TraceProcStop(_p_)
				}
				n++
				_p_.Syscalltick++
				_sched.Handoffp(_p_)
			}
			_sched.Incidlelocked(1)
		} else if s == _lock.Prunning {
			// Preempt G if it's running for more than 10ms.
			t := int64(_p_.Schedtick)
			if int64(pd.schedtick) != t {
				pd.schedtick = uint32(t)
				pd.schedwhen = now
				continue
			}
			if pd.schedwhen+10*1000*1000 > now {
				continue
			}
			_lock.Preemptone(_p_)
		}
	}
	return uint32(n)
}

func testSchedLocalQueue() {
	_p_ := new(_core.P)
	gs := make([]_core.G, len(_p_.Runq))
	for i := 0; i < len(_p_.Runq); i++ {
		if _sched.Runqget(_p_) != nil {
			_lock.Throw("runq is not empty initially")
		}
		for j := 0; j < i; j++ {
			_sched.Runqput(_p_, &gs[i])
		}
		for j := 0; j < i; j++ {
			if _sched.Runqget(_p_) != &gs[i] {
				print("bad element at iter ", i, "/", j, "\n")
				_lock.Throw("bad element")
			}
		}
		if _sched.Runqget(_p_) != nil {
			_lock.Throw("runq is not empty afterwards")
		}
	}
}

func testSchedLocalQueueSteal() {
	p1 := new(_core.P)
	p2 := new(_core.P)
	gs := make([]_core.G, len(p1.Runq))
	for i := 0; i < len(p1.Runq); i++ {
		for j := 0; j < i; j++ {
			gs[j].Sig = 0
			_sched.Runqput(p1, &gs[j])
		}
		gp := _sched.Runqsteal(p2, p1)
		s := 0
		if gp != nil {
			s++
			gp.Sig++
		}
		for {
			gp = _sched.Runqget(p2)
			if gp == nil {
				break
			}
			s++
			gp.Sig++
		}
		for {
			gp = _sched.Runqget(p1)
			if gp == nil {
				break
			}
			gp.Sig++
		}
		for j := 0; j < i; j++ {
			if gs[j].Sig != 1 {
				print("bad element ", j, "(", gs[j].Sig, ") at iter ", i, "\n")
				_lock.Throw("bad element")
			}
		}
		if s != i/2 && s != i/2+1 {
			print("bad steal ", s, ", want ", i/2, " or ", i/2+1, ", iter ", i, "\n")
			_lock.Throw("bad steal")
		}
	}
}

func setMaxThreads(in int) (out int) {
	_lock.Lock(&_core.Sched.Lock)
	out = int(_core.Sched.Maxmcount)
	_core.Sched.Maxmcount = int32(in)
	_sched.Checkmcount()
	_lock.Unlock(&_core.Sched.Lock)
	return
}

//go:nosplit
func procPin() int {
	_g_ := _core.Getg()
	mp := _g_.M

	mp.Locks++
	return int(mp.P.Id)
}

//go:nosplit
func procUnpin() {
	_g_ := _core.Getg()
	_g_.M.Locks--
}

//go:linkname sync_runtime_procPin sync.runtime_procPin
//go:nosplit
func sync_runtime_procPin() int {
	return procPin()
}

//go:linkname sync_runtime_procUnpin sync.runtime_procUnpin
//go:nosplit
func sync_runtime_procUnpin() {
	procUnpin()
}

//go:linkname sync_atomic_runtime_procPin sync/atomic.runtime_procPin
//go:nosplit
func sync_atomic_runtime_procPin() int {
	return procPin()
}

//go:linkname sync_atomic_runtime_procUnpin sync/atomic.runtime_procUnpin
//go:nosplit
func sync_atomic_runtime_procUnpin() {
	procUnpin()
}
