// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	_base "runtime/internal/base"
	_gc "runtime/internal/gc"
	_iface "runtime/internal/iface"
	_lock "runtime/internal/lock"
	_print "runtime/internal/print"
	_race "runtime/internal/race"
	"unsafe"
)

var (
	g0 _base.G
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
// Design doc at https://golang.org/s/go11sched.

const (
	// Number of goroutine ids to grab from sched.goidgen to local per-P cache at once.
	// 16 seems to provide enough amortization, but other than that it's mostly arbitrary number.
	_GoidCacheBatch = 16
)

// The bootstrap sequence is:
//
//	call osinit
//	call schedinit
//	make & queue new G
//	call runtime·mstart
//
// The new G calls runtime·main.
func schedinit() {
	// raceinit must be the first call to race detector.
	// In particular, it must be done before mallocinit below calls racemapshadow.
	_g_ := _base.Getg()
	if _base.Raceenabled {
		_g_.Racectx = _race.Raceinit()
	}

	_base.Sched.Maxmcount = 10000

	// Cache the framepointer experiment.  This affects stack unwinding.
	_base.Framepointer_enabled = haveexperiment("framepointer")

	tracebackinit()
	moduledataverify()
	stackinit()
	mallocinit()
	_base.Mcommoninit(_g_.M)

	goargs()
	goenvs()
	parsedebugvars()
	_gc.Gcinit()

	_base.Sched.Lastpoll = uint64(_base.Nanotime())
	procs := int(_base.Ncpu)
	if n := _gc.Atoi(_gc.Gogetenv("GOMAXPROCS")); n > 0 {
		if n > _base.MaxGomaxprocs {
			n = _base.MaxGomaxprocs
		}
		procs = n
	}
	if _gc.Procresize(int32(procs)) != nil {
		_base.Throw("unknown runnable goroutine during bootstrap")
	}

	if buildVersion == "" {
		// Condition should never trigger.  This code just serves
		// to ensure runtime·buildVersion is kept in the resulting binary.
		buildVersion = "unknown"
	}
}

func isscanstatus(status uint32) bool {
	if status == _base.Gscan {
		_base.Throw("isscanstatus: Bad status Gscan")
	}
	return status&_base.Gscan == _base.Gscan
}

// stopTheWorld stops all P's from executing goroutines, interrupting
// all goroutines at GC safe points and records reason as the reason
// for the stop. On return, only the current goroutine's P is running.
// stopTheWorld must not be called from a system stack and the caller
// must not hold worldsema. The caller must call startTheWorld when
// other P's should resume execution.
//
// stopTheWorld is safe for multiple goroutines to call at the
// same time. Each will execute its own stop, and the stops will
// be serialized.
//
// This is also used by routines that do stack dumps. If the system is
// in panic or being exited, this may not reliably stop all
// goroutines.
func stopTheWorld(reason string) {
	_lock.Semacquire(&_gc.Worldsema, false)
	_base.Getg().M.Preemptoff = reason
	_base.Systemstack(_gc.StopTheWorldWithSema)
}

// startTheWorld undoes the effects of stopTheWorld.
func startTheWorld() {
	_base.Systemstack(_gc.StartTheWorldWithSema)
	// worldsema must be held over startTheWorldWithSema to ensure
	// gomaxprocs cannot change while worldsema is held.
	_lock.Semrelease(&_gc.Worldsema)
	_base.Getg().M.Preemptoff = ""
}

// When running with cgo, we call _cgo_thread_start
// to start threads for us so that we can play nicely with
// foreign code.
var cgoThreadStart unsafe.Pointer

// needm is called when a cgo callback happens on a
// thread without an m (a thread not created by Go).
// In this case, needm is expected to find an m to use
// and return with m, g initialized correctly.
// Since m and g are not set now (likely nil, but see below)
// needm is limited in what routines it can call. In particular
// it can only call nosplit functions (textflag 7) and cannot
// do any scheduling that requires an m.
//
// In order to avoid needing heavy lifting here, we adopt
// the following strategy: there is a stack of available m's
// that can be stolen. Using compare-and-swap
// to pop from the stack has ABA races, so we simulate
// a lock by doing an exchange (via casp) to steal the stack
// head and replace the top pointer with MLOCKED (1).
// This serves as a simple spin lock that we can use even
// without an m. The thread that locks the stack in this way
// unlocks the stack by storing a valid stack head pointer.
//
// In order to make sure that there is always an m structure
// available to be stolen, we maintain the invariant that there
// is always one more than needed. At the beginning of the
// program (if cgo is in use) the list is seeded with a single m.
// If needm finds that it has taken the last m off the list, its job
// is - once it has installed its own m so that it can do things like
// allocate memory - to create a spare m and put it on the list.
//
// Each of these extra m's also has a g0 and a curg that are
// pressed into service as the scheduling stack and current
// goroutine for the duration of the cgo callback.
//
// When the callback is done with the m, it calls dropm to
// put the m back on the list.
//go:nosplit
func needm(x byte) {
	if _base.Iscgo && !_base.CgoHasExtraM {
		// Can happen if C/C++ code calls Go from a global ctor.
		// Can not throw, because scheduler is not initialized yet.
		_print.Write(2, unsafe.Pointer(&earlycgocallback[0]), int32(len(earlycgocallback)))
		_base.Exit(1)
	}

	// Lock extra list, take head, unlock popped list.
	// nilokay=false is safe here because of the invariant above,
	// that the extra list always contains or will soon contain
	// at least one m.
	mp := _base.Lockextra(false)

	// Set needextram when we've just emptied the list,
	// so that the eventual call into cgocallbackg will
	// allocate a new m for the extra list. We delay the
	// allocation until then so that it can be done
	// after exitsyscall makes sure it is okay to be
	// running at all (that is, there's no garbage collection
	// running right now).
	mp.Needextram = mp.Schedlink == 0
	_base.Unlockextra(mp.Schedlink.Ptr())

	// Install g (= m->g0) and set the stack bounds
	// to match the current stack. We don't actually know
	// how big the stack is, like we don't know how big any
	// scheduling stack is, but we assume there's at least 32 kB,
	// which is more than enough for us.
	setg(mp.G0)
	_g_ := _base.Getg()
	_g_.Stack.Hi = uintptr(_base.Noescape(unsafe.Pointer(&x))) + 1024
	_g_.Stack.Lo = uintptr(_base.Noescape(unsafe.Pointer(&x))) - 32*1024
	_g_.Stackguard0 = _g_.Stack.Lo + _base.StackGuard

	_base.Msigsave(mp)
	// Initialize this thread to use the m.
	_base.Asminit()
	_base.Minit()
}

var earlycgocallback = []byte("fatal error: cgo callback before cgo call\n")

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
	mp := _base.Getg().M
	mnext := _base.Lockextra(true)
	mp.Schedlink.Set(mnext)

	setg(nil)
	_base.Unlockextra(mp)
}

func gopreempt_m(gp *_base.G) {
	if _base.Trace.Enabled {
		traceGoPreempt()
	}
	_gc.GoschedImpl(gp)
}

// Finishes execution of the current goroutine.
func goexit1() {
	if _base.Raceenabled {
		_race.Racegoend()
	}
	if _base.Trace.Enabled {
		traceGoEnd()
	}
	_base.Mcall(goexit0)
}

// goexit continuation on g0.
func goexit0(gp *_base.G) {
	_g_ := _base.Getg()

	_base.Casgstatus(gp, _base.Grunning, _base.Gdead)
	gp.M = nil
	gp.Lockedm = nil
	_g_.M.Lockedg = nil
	gp.Paniconfault = false
	gp.Defer = nil // should be true already but just in case.
	gp.Panic = nil // non-nil for Goexit during panic. points at stack-allocated data.
	gp.Writebuf = nil
	gp.Waitreason = ""
	gp.Param = nil

	_base.Dropg()

	if _g_.M.Locked&^_base.LockExternal != 0 {
		print("invalid m->locked = ", _g_.M.Locked, "\n")
		_base.Throw("internal lockOSThread error")
	}
	_g_.M.Locked = 0
	gfput(_g_.M.P.Ptr(), gp)
	_base.Schedule()
}

// The goroutine g is about to enter a system call.
// Record that it's not using the cpu anymore.
// This is called only from the go syscall library and cgocall,
// not from the low-level system calls used by the
//
// Entersyscall cannot split the stack: the gosave must
// make g->sched refer to the caller's stack segment, because
// entersyscall is going to return immediately after.
//
// Nothing entersyscall calls can split the stack either.
// We cannot safely move the stack during an active call to syscall,
// because we do not know which of the uintptr arguments are
// really pointers (back into the stack).
// In practice, this means that we make the fast path run through
// entersyscall doing no-split things, and the slow path has to use systemstack
// to run bigger things on the system stack.
//
// reentersyscall is the entry point used by cgo callbacks, where explicitly
// saved SP and PC are restored. This is needed when exitsyscall will be called
// from a function further up in the call stack than the parent, as g->syscallsp
// must always point to a valid stack frame. entersyscall below is the normal
// entry point for syscalls, which obtains the SP and PC from the caller.
//
// Syscall tracing:
// At the start of a syscall we emit traceGoSysCall to capture the stack trace.
// If the syscall does not block, that is it, we do not emit any other events.
// If the syscall blocks (that is, P is retaken), retaker emits traceGoSysBlock;
// when syscall returns we emit traceGoSysExit and when the goroutine starts running
// (potentially instantly, if exitsyscallfast returns true) we emit traceGoStart.
// To ensure that traceGoSysExit is emitted strictly after traceGoSysBlock,
// we remember current value of syscalltick in m (_g_.m.syscalltick = _g_.m.p.ptr().syscalltick),
// whoever emits traceGoSysBlock increments p.syscalltick afterwards;
// and we wait for the increment before emitting traceGoSysExit.
// Note that the increment is done even if tracing is not enabled,
// because tracing can be enabled in the middle of syscall. We don't want the wait to hang.
//
//go:nosplit
func reentersyscall(pc, sp uintptr) {
	_g_ := _base.Getg()

	// Disable preemption because during this function g is in Gsyscall status,
	// but can have inconsistent g->sched, do not let GC observe it.
	_g_.M.Locks++

	if _base.Trace.Enabled {
		_base.Systemstack(_base.TraceGoSysCall)
	}

	// Entersyscall must not call any function that might split/grow the stack.
	// (See details in comment above.)
	// Catch calls that might, by replacing the stack guard with something that
	// will trip any stack check and leaving a flag to tell newstack to die.
	_g_.Stackguard0 = _base.StackPreempt
	_g_.Throwsplit = true

	// Leave SP around for GC and traceback.
	_base.Save(pc, sp)
	_g_.Syscallsp = sp
	_g_.Syscallpc = pc
	_base.Casgstatus(_g_, _base.Grunning, _base.Gsyscall)
	if _g_.Syscallsp < _g_.Stack.Lo || _g_.Stack.Hi < _g_.Syscallsp {
		_base.Systemstack(func() {
			print("entersyscall inconsistent ", _base.Hex(_g_.Syscallsp), " [", _base.Hex(_g_.Stack.Lo), ",", _base.Hex(_g_.Stack.Hi), "]\n")
			_base.Throw("entersyscall")
		})
	}

	if _base.Atomicload(&_base.Sched.Sysmonwait) != 0 { // TODO: fast atomic
		_base.Systemstack(entersyscall_sysmon)
		_base.Save(pc, sp)
	}

	if _g_.M.P.Ptr().RunSafePointFn != 0 {
		// runSafePointFn may stack split if run on this stack
		_base.Systemstack(_base.RunSafePointFn)
		_base.Save(pc, sp)
	}

	_g_.M.Syscalltick = _g_.M.P.Ptr().Syscalltick
	_g_.Sysblocktraced = true
	_g_.M.Mcache = nil
	_g_.M.P.Ptr().M = 0
	_base.Atomicstore(&_g_.M.P.Ptr().Status, _base.Psyscall)
	if _base.Sched.Gcwaiting != 0 {
		_base.Systemstack(entersyscall_gcwait)
		_base.Save(pc, sp)
	}

	// Goroutines must not split stacks in Gsyscall status (it would corrupt g->sched).
	// We set _StackGuard to StackPreempt so that first split stack check calls morestack.
	// Morestack detects this case and throws.
	_g_.Stackguard0 = _base.StackPreempt
	_g_.M.Locks--
}

// Standard syscall entry used by the go syscall library and normal cgo calls.
//go:nosplit
func entersyscall(dummy int32) {
	reentersyscall(_base.Getcallerpc(unsafe.Pointer(&dummy)), _base.Getcallersp(unsafe.Pointer(&dummy)))
}

func entersyscall_sysmon() {
	_base.Lock(&_base.Sched.Lock)
	if _base.Atomicload(&_base.Sched.Sysmonwait) != 0 {
		_base.Atomicstore(&_base.Sched.Sysmonwait, 0)
		_base.Notewakeup(&_base.Sched.Sysmonnote)
	}
	_base.Unlock(&_base.Sched.Lock)
}

func entersyscall_gcwait() {
	_g_ := _base.Getg()
	_p_ := _g_.M.P.Ptr()

	_base.Lock(&_base.Sched.Lock)
	if _base.Sched.Stopwait > 0 && _base.Cas(&_p_.Status, _base.Psyscall, _base.Pgcstop) {
		if _base.Trace.Enabled {
			_base.TraceGoSysBlock(_p_)
			_base.TraceProcStop(_p_)
		}
		_p_.Syscalltick++
		if _base.Sched.Stopwait--; _base.Sched.Stopwait == 0 {
			_base.Notewakeup(&_base.Sched.Stopnote)
		}
	}
	_base.Unlock(&_base.Sched.Lock)
}

func beforefork() {
	gp := _base.Getg().M.Curg

	// Fork can hang if preempted with signals frequently enough (see issue 5517).
	// Ensure that we stay on the same M where we disable profiling.
	gp.M.Locks++
	if gp.M.Profilehz != 0 {
		_base.Resetcpuprofiler(0)
	}

	// This function is called before fork in syscall package.
	// Code between fork and exec must not allocate memory nor even try to grow stack.
	// Here we spoil g->_StackGuard to reliably detect any attempts to grow stack.
	// runtime_AfterFork will undo this in parent process, but not in child.
	gp.Stackguard0 = _base.StackFork
}

// Called from syscall package before fork.
//go:linkname syscall_runtime_BeforeFork syscall.runtime_BeforeFork
//go:nosplit
func syscall_runtime_BeforeFork() {
	_base.Systemstack(beforefork)
}

func afterfork() {
	gp := _base.Getg().M.Curg

	// See the comment in beforefork.
	gp.Stackguard0 = gp.Stack.Lo + _base.StackGuard

	hz := _base.Sched.Profilehz
	if hz != 0 {
		_base.Resetcpuprofiler(hz)
	}
	gp.M.Locks--
}

// Called from syscall package after fork in parent.
//go:linkname syscall_runtime_AfterFork syscall.runtime_AfterFork
//go:nosplit
func syscall_runtime_AfterFork() {
	_base.Systemstack(afterfork)
}

// Create a new g running fn with siz bytes of arguments.
// Put it on the queue of g's waiting to run.
// The compiler turns a go statement into a call to this.
// Cannot split the stack because it assumes that the arguments
// are available sequentially after &fn; they would not be
// copied if a stack split occurred.
//go:nosplit
func newproc(siz int32, fn *_base.Funcval) {
	argp := _base.Add(unsafe.Pointer(&fn), _base.PtrSize)
	pc := _base.Getcallerpc(unsafe.Pointer(&siz))
	_base.Systemstack(func() {
		newproc1(fn, (*uint8)(argp), siz, 0, pc)
	})
}

// Create a new g running fn with narg bytes of arguments starting
// at argp and returning nret bytes of results.  callerpc is the
// address of the go statement that created this.  The new g is put
// on the queue of g's waiting to run.
func newproc1(fn *_base.Funcval, argp *uint8, narg int32, nret int32, callerpc uintptr) *_base.G {
	_g_ := _base.Getg()

	if fn == nil {
		_g_.M.Throwing = -1 // do not dump full stacks
		_base.Throw("go of nil func value")
	}
	_g_.M.Locks++ // disable preemption because it can be holding p in a local var
	siz := narg + nret
	siz = (siz + 7) &^ 7

	// We could allocate a larger initial stack if necessary.
	// Not worth it: this is almost always an error.
	// 4*sizeof(uintreg): extra space added below
	// sizeof(uintreg): caller's LR (arm) or return address (x86, in gostartcall).
	if siz >= _base.StackMin-4*_base.RegSize-_base.RegSize {
		_base.Throw("newproc: function arguments too large for new goroutine")
	}

	_p_ := _g_.M.P.Ptr()
	newg := gfget(_p_)
	if newg == nil {
		newg = _base.Malg(_base.StackMin)
		_base.Casgstatus(newg, _base.Gidle, _base.Gdead)
		_base.Allgadd(newg) // publishes with a g->status of Gdead so GC scanner doesn't look at uninitialized stack.
	}
	if newg.Stack.Hi == 0 {
		_base.Throw("newproc1: newg missing stack")
	}

	if _base.Readgstatus(newg) != _base.Gdead {
		_base.Throw("newproc1: new g is not Gdead")
	}

	totalSize := 4*_base.RegSize + uintptr(siz) // extra space in case of reads slightly beyond frame
	if hasLinkRegister {
		totalSize += _base.PtrSize
	}
	totalSize += -totalSize & (_gc.SpAlign - 1) // align to spAlign
	sp := newg.Stack.Hi - totalSize
	spArg := sp
	if hasLinkRegister {
		// caller's LR
		*(*unsafe.Pointer)(unsafe.Pointer(sp)) = nil
		spArg += _base.PtrSize
	}
	_base.Memmove(unsafe.Pointer(spArg), unsafe.Pointer(argp), uintptr(narg))

	_base.Memclr(unsafe.Pointer(&newg.Sched), unsafe.Sizeof(newg.Sched))
	newg.Sched.Sp = sp
	newg.Sched.Pc = _base.FuncPC(_base.Goexit) + _base.PCQuantum // +PCQuantum so that previous instruction is in same function
	newg.Sched.G = _base.Guintptr(unsafe.Pointer(newg))
	gostartcallfn(&newg.Sched, fn)
	newg.Gopc = callerpc
	newg.Startpc = fn.Fn
	_base.Casgstatus(newg, _base.Gdead, _base.Grunnable)

	if _p_.Goidcache == _p_.Goidcacheend {
		// Sched.goidgen is the last allocated id,
		// this batch must be [sched.goidgen+1, sched.goidgen+GoidCacheBatch].
		// At startup sched.goidgen=0, so main goroutine receives goid=1.
		_p_.Goidcache = _base.Xadd64(&_base.Sched.Goidgen, _GoidCacheBatch)
		_p_.Goidcache -= _GoidCacheBatch - 1
		_p_.Goidcacheend = _p_.Goidcache + _GoidCacheBatch
	}
	newg.Goid = int64(_p_.Goidcache)
	_p_.Goidcache++
	if _base.Raceenabled {
		newg.Racectx = _base.Racegostart(callerpc)
	}
	if _base.Trace.Enabled {
		traceGoCreate(newg, newg.Startpc)
	}
	_base.Runqput(_p_, newg, true)

	if _base.Atomicload(&_base.Sched.Npidle) != 0 && _base.Atomicload(&_base.Sched.Nmspinning) == 0 && unsafe.Pointer(fn.Fn) != unsafe.Pointer(_base.FuncPC(main)) { // TODO: fast atomic
		_base.Wakep()
	}
	_g_.M.Locks--
	if _g_.M.Locks == 0 && _g_.Preempt { // restore the preemption request in case we've cleared it in newstack
		_g_.Stackguard0 = _base.StackPreempt
	}
	return newg
}

// Put on gfree list.
// If local list is too long, transfer a batch to the global list.
func gfput(_p_ *_base.P, gp *_base.G) {
	if _base.Readgstatus(gp) != _base.Gdead {
		_base.Throw("gfput: bad status (not Gdead)")
	}

	stksize := gp.StackAlloc

	if stksize != _base.FixedStack {
		// non-standard stack size - free it.
		_gc.Stackfree(gp.Stack, gp.StackAlloc)
		gp.Stack.Lo = 0
		gp.Stack.Hi = 0
		gp.Stackguard0 = 0
		gp.Stkbar = nil
		gp.StkbarPos = 0
	} else {
		// Reset stack barriers.
		gp.Stkbar = gp.Stkbar[:0]
		gp.StkbarPos = 0
	}

	gp.Schedlink.Set(_p_.Gfree)
	_p_.Gfree = gp
	_p_.Gfreecnt++
	if _p_.Gfreecnt >= 64 {
		_base.Lock(&_base.Sched.Gflock)
		for _p_.Gfreecnt >= 32 {
			_p_.Gfreecnt--
			gp = _p_.Gfree
			_p_.Gfree = gp.Schedlink.Ptr()
			gp.Schedlink.Set(_base.Sched.Gfree)
			_base.Sched.Gfree = gp
			_base.Sched.Ngfree++
		}
		_base.Unlock(&_base.Sched.Gflock)
	}
}

// Get from gfree list.
// If local list is empty, grab a batch from global list.
func gfget(_p_ *_base.P) *_base.G {
retry:
	gp := _p_.Gfree
	if gp == nil && _base.Sched.Gfree != nil {
		_base.Lock(&_base.Sched.Gflock)
		for _p_.Gfreecnt < 32 && _base.Sched.Gfree != nil {
			_p_.Gfreecnt++
			gp = _base.Sched.Gfree
			_base.Sched.Gfree = gp.Schedlink.Ptr()
			_base.Sched.Ngfree--
			gp.Schedlink.Set(_p_.Gfree)
			_p_.Gfree = gp
		}
		_base.Unlock(&_base.Sched.Gflock)
		goto retry
	}
	if gp != nil {
		_p_.Gfree = gp.Schedlink.Ptr()
		_p_.Gfreecnt--
		if gp.Stack.Lo == 0 {
			// Stack was deallocated in gfput.  Allocate a new one.
			_base.Systemstack(func() {
				gp.Stack, gp.Stkbar = _base.Stackalloc(_base.FixedStack)
			})
			gp.Stackguard0 = gp.Stack.Lo + _base.StackGuard
			gp.StackAlloc = _base.FixedStack
		} else {
			if _base.Raceenabled {
				_base.Racemalloc(unsafe.Pointer(gp.Stack.Lo), gp.StackAlloc)
			}
		}
	}
	return gp
}

// Breakpoint executes a breakpoint trap.
func Breakpoint() {
	breakpoint()
}

// dolockOSThread is called by LockOSThread and lockOSThread below
// after they modify m.locked. Do not allow preemption during this call,
// or else the m might be different in this function than in the caller.
//go:nosplit
func dolockOSThread() {
	_g_ := _base.Getg()
	_g_.M.Lockedg = _g_
	_g_.Lockedm = _g_.M
}

//go:nosplit

// LockOSThread wires the calling goroutine to its current operating system thread.
// Until the calling goroutine exits or calls UnlockOSThread, it will always
// execute in that thread, and no other goroutine can.
func LockOSThread() {
	_base.Getg().M.Locked |= _base.LockExternal
	dolockOSThread()
}

//go:nosplit
func lockOSThread() {
	_base.Getg().M.Locked += _base.LockInternal
	dolockOSThread()
}

// dounlockOSThread is called by UnlockOSThread and unlockOSThread below
// after they update m->locked. Do not allow preemption during this call,
// or else the m might be in different in this function than in the caller.
//go:nosplit
func dounlockOSThread() {
	_g_ := _base.Getg()
	if _g_.M.Locked != 0 {
		return
	}
	_g_.M.Lockedg = nil
	_g_.Lockedm = nil
}

//go:nosplit

// UnlockOSThread unwires the calling goroutine from its fixed operating system thread.
// If the calling goroutine has not called LockOSThread, UnlockOSThread is a no-op.
func UnlockOSThread() {
	_base.Getg().M.Locked &^= _base.LockExternal
	dounlockOSThread()
}

//go:nosplit
func unlockOSThread() {
	_g_ := _base.Getg()
	if _g_.M.Locked < _base.LockInternal {
		_base.Systemstack(badunlockosthread)
	}
	_g_.M.Locked -= _base.LockInternal
	dounlockOSThread()
}

func badunlockosthread() {
	_base.Throw("runtime: internal error: misuse of lockOSThread/unlockOSThread")
}

func gcount() int32 {
	n := int32(_base.Allglen) - _base.Sched.Ngfree
	for i := 0; ; i++ {
		_p_ := _base.Allp[i]
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

// Arrange to call fn with a traceback hz times a second.
func setcpuprofilerate_m(hz int32) {
	// Force sane arguments.
	if hz < 0 {
		hz = 0
	}

	// Disable preemption, otherwise we can be rescheduled to another thread
	// that has profiling enabled.
	_g_ := _base.Getg()
	_g_.M.Locks++

	// Stop profiler on this thread so that it is safe to lock prof.
	// if a profiling signal came in while we had prof locked,
	// it would deadlock.
	_base.Resetcpuprofiler(0)

	for !_base.Cas(&_base.Prof.Lock, 0, 1) {
		_base.Osyield()
	}
	_base.Prof.Hz = hz
	_base.Atomicstore(&_base.Prof.Lock, 0)

	_base.Lock(&_base.Sched.Lock)
	_base.Sched.Profilehz = hz
	_base.Unlock(&_base.Sched.Lock)

	if hz != 0 {
		_base.Resetcpuprofiler(hz)
	}

	_g_.M.Locks--
}

func sysmon() {
	// If we go two minutes without a garbage collection, force one to run.
	forcegcperiod := int64(2 * 60 * 1e9)

	// If a heap span goes unused for 5 minutes after a garbage collection,
	// we hand it back to the operating system.
	scavengelimit := int64(5 * 60 * 1e9)

	if _base.Debug.Scavenge > 0 {
		// Scavenge-a-lot for testing.
		forcegcperiod = 10 * 1e6
		scavengelimit = 20 * 1e6
	}

	lastscavenge := _base.Nanotime()
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
		_base.Usleep(delay)
		if _base.Debug.Schedtrace <= 0 && (_base.Sched.Gcwaiting != 0 || _base.Atomicload(&_base.Sched.Npidle) == uint32(_base.Gomaxprocs)) { // TODO: fast atomic
			_base.Lock(&_base.Sched.Lock)
			if _base.Atomicload(&_base.Sched.Gcwaiting) != 0 || _base.Atomicload(&_base.Sched.Npidle) == uint32(_base.Gomaxprocs) {
				_base.Atomicstore(&_base.Sched.Sysmonwait, 1)
				_base.Unlock(&_base.Sched.Lock)
				_gc.Notetsleep(&_base.Sched.Sysmonnote, maxsleep)
				_base.Lock(&_base.Sched.Lock)
				_base.Atomicstore(&_base.Sched.Sysmonwait, 0)
				_base.Noteclear(&_base.Sched.Sysmonnote)
				idle = 0
				delay = 20
			}
			_base.Unlock(&_base.Sched.Lock)
		}
		// poll network if not polled for more than 10ms
		lastpoll := int64(_base.Atomicload64(&_base.Sched.Lastpoll))
		now := _base.Nanotime()
		unixnow := _gc.Unixnanotime()
		if lastpoll != 0 && lastpoll+10*1000*1000 < now {
			_base.Cas64(&_base.Sched.Lastpoll, uint64(lastpoll), uint64(now))
			gp := _base.Netpoll(false) // non-blocking - returns list of goroutines
			if gp != nil {
				// Need to decrement number of idle locked M's
				// (pretending that one more is running) before injectglist.
				// Otherwise it can lead to the following situation:
				// injectglist grabs all P's but before it starts M's to run the P's,
				// another M returns from syscall, finishes running its G,
				// observes that there is no work to do and no other running M's
				// and reports deadlock.
				_base.Incidlelocked(-1)
				_base.Injectglist(gp)
				_base.Incidlelocked(1)
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
		lastgc := int64(_base.Atomicload64(&_base.Memstats.Last_gc))
		if lastgc != 0 && unixnow-lastgc > forcegcperiod && _base.Atomicload(&forcegc.idle) != 0 && _iface.Atomicloaduint(&_iface.Bggc.Working) == 0 {
			_base.Lock(&forcegc.lock)
			forcegc.idle = 0
			forcegc.g.Schedlink = 0
			_base.Injectglist(forcegc.g)
			_base.Unlock(&forcegc.lock)
		}
		// scavenge heap once in a while
		if lastscavenge+scavengelimit/2 < now {
			mHeap_Scavenge(int32(nscavenge), uint64(now), uint64(scavengelimit))
			lastscavenge = now
			nscavenge++
		}
		if _base.Debug.Schedtrace > 0 && lasttrace+int64(_base.Debug.Schedtrace*1000000) <= now {
			lasttrace = now
			_base.Schedtrace(_base.Debug.Scheddetail > 0)
		}
	}
}

var pdesc [_base.MaxGomaxprocs]struct {
	schedtick   uint32
	schedwhen   int64
	syscalltick uint32
	syscallwhen int64
}

// forcePreemptNS is the time slice given to a G before it is
// preempted.
const forcePreemptNS = 10 * 1000 * 1000 // 10ms

func retake(now int64) uint32 {
	n := 0
	for i := int32(0); i < _base.Gomaxprocs; i++ {
		_p_ := _base.Allp[i]
		if _p_ == nil {
			continue
		}
		pd := &pdesc[i]
		s := _p_.Status
		if s == _base.Psyscall {
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
			if _base.Runqempty(_p_) && _base.Atomicload(&_base.Sched.Nmspinning)+_base.Atomicload(&_base.Sched.Npidle) > 0 && pd.syscallwhen+10*1000*1000 > now {
				continue
			}
			// Need to decrement number of idle locked M's
			// (pretending that one more is running) before the CAS.
			// Otherwise the M from which we retake can exit the syscall,
			// increment nmidle and report deadlock.
			_base.Incidlelocked(-1)
			if _base.Cas(&_p_.Status, s, _base.Pidle) {
				if _base.Trace.Enabled {
					_base.TraceGoSysBlock(_p_)
					_base.TraceProcStop(_p_)
				}
				n++
				_p_.Syscalltick++
				_base.Handoffp(_p_)
			}
			_base.Incidlelocked(1)
		} else if s == _base.Prunning {
			// Preempt G if it's running for too long.
			t := int64(_p_.Schedtick)
			if int64(pd.schedtick) != t {
				pd.schedtick = uint32(t)
				pd.schedwhen = now
				continue
			}
			if pd.schedwhen+forcePreemptNS > now {
				continue
			}
			_base.Preemptone(_p_)
		}
	}
	return uint32(n)
}

func testSchedLocalQueue() {
	_p_ := new(_base.P)
	gs := make([]_base.G, len(_p_.Runq))
	for i := 0; i < len(_p_.Runq); i++ {
		if g, _ := _base.Runqget(_p_); g != nil {
			_base.Throw("runq is not empty initially")
		}
		for j := 0; j < i; j++ {
			_base.Runqput(_p_, &gs[i], false)
		}
		for j := 0; j < i; j++ {
			if g, _ := _base.Runqget(_p_); g != &gs[i] {
				print("bad element at iter ", i, "/", j, "\n")
				_base.Throw("bad element")
			}
		}
		if g, _ := _base.Runqget(_p_); g != nil {
			_base.Throw("runq is not empty afterwards")
		}
	}
}

func testSchedLocalQueueSteal() {
	p1 := new(_base.P)
	p2 := new(_base.P)
	gs := make([]_base.G, len(p1.Runq))
	for i := 0; i < len(p1.Runq); i++ {
		for j := 0; j < i; j++ {
			gs[j].Sig = 0
			_base.Runqput(p1, &gs[j], false)
		}
		gp := _base.Runqsteal(p2, p1, true)
		s := 0
		if gp != nil {
			s++
			gp.Sig++
		}
		for {
			gp, _ = _base.Runqget(p2)
			if gp == nil {
				break
			}
			s++
			gp.Sig++
		}
		for {
			gp, _ = _base.Runqget(p1)
			if gp == nil {
				break
			}
			gp.Sig++
		}
		for j := 0; j < i; j++ {
			if gs[j].Sig != 1 {
				print("bad element ", j, "(", gs[j].Sig, ") at iter ", i, "\n")
				_base.Throw("bad element")
			}
		}
		if s != i/2 && s != i/2+1 {
			print("bad steal ", s, ", want ", i/2, " or ", i/2+1, ", iter ", i, "\n")
			_base.Throw("bad steal")
		}
	}
}

func setMaxThreads(in int) (out int) {
	_base.Lock(&_base.Sched.Lock)
	out = int(_base.Sched.Maxmcount)
	_base.Sched.Maxmcount = int32(in)
	_base.Checkmcount()
	_base.Unlock(&_base.Sched.Lock)
	return
}

func haveexperiment(name string) bool {
	x := goexperiment
	for x != "" {
		xname := ""
		i := _base.Index(x, ",")
		if i < 0 {
			xname, x = x, ""
		} else {
			xname, x = x[:i], x[i+1:]
		}
		if xname == name {
			return true
		}
	}
	return false
}

//go:nosplit
func procPin() int {
	_g_ := _base.Getg()
	mp := _g_.M

	mp.Locks++
	return int(mp.P.Ptr().Id)
}

//go:nosplit
func procUnpin() {
	_g_ := _base.Getg()
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

// Active spinning for sync.Mutex.
//go:linkname sync_runtime_canSpin sync.runtime_canSpin
//go:nosplit
func sync_runtime_canSpin(i int) bool {
	// sync.Mutex is cooperative, so we are conservative with spinning.
	// Spin only few times and only if running on a multicore machine and
	// GOMAXPROCS>1 and there is at least one other running P and local runq is empty.
	// As opposed to runtime mutex we don't do passive spinning here,
	// because there can be work on global runq on on other Ps.
	if i >= _base.Active_spin || _base.Ncpu <= 1 || _base.Gomaxprocs <= int32(_base.Sched.Npidle+_base.Sched.Nmspinning)+1 {
		return false
	}
	if p := _base.Getg().M.P.Ptr(); !_base.Runqempty(p) {
		return false
	}
	return true
}

//go:linkname sync_runtime_doSpin sync.runtime_doSpin
//go:nosplit
func sync_runtime_doSpin() {
	_base.Procyield(_base.Active_spin_cnt)
}
