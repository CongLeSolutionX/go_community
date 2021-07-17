// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	"runtime/internal/atomic"
	"runtime/internal/sys"
	"unsafe"
)

type mOS struct {
	// profileTimer is accessed atomically. It holds the ID of the POSIX
	// interval timer for profiling CPU usage on this thread.
	//
	// Bits 2 to 33 hold the 32 bits of an interval timer's id, which is valid
	// when the "profileTimerValid" flag (bit 0) is set. A thread creates its
	// own timer, so all transitions that set the "profileTimerValid" flag
	// originate in the thread itself.
	//
	// The "profileTimerShutdown" flag (bit 1) coordinates deletion of the
	// timer.
	profileTimer int64
}

const (
	profileTimerValid    = 0x1
	profileTimerShutdown = 0x2
)

//go:noescape
func futex(addr unsafe.Pointer, op int32, val uint32, ts, addr2 unsafe.Pointer, val3 uint32) int32

// Linux futex.
//
//	futexsleep(uint32 *addr, uint32 val)
//	futexwakeup(uint32 *addr)
//
// Futexsleep atomically checks if *addr == val and if so, sleeps on addr.
// Futexwakeup wakes up threads sleeping on addr.
// Futexsleep is allowed to wake up spuriously.

const (
	_FUTEX_PRIVATE_FLAG = 128
	_FUTEX_WAIT_PRIVATE = 0 | _FUTEX_PRIVATE_FLAG
	_FUTEX_WAKE_PRIVATE = 1 | _FUTEX_PRIVATE_FLAG
)

// Atomically,
//	if(*addr == val) sleep
// Might be woken up spuriously; that's allowed.
// Don't sleep longer than ns; ns < 0 means forever.
//go:nosplit
func futexsleep(addr *uint32, val uint32, ns int64) {
	// Some Linux kernels have a bug where futex of
	// FUTEX_WAIT returns an internal error code
	// as an errno. Libpthread ignores the return value
	// here, and so can we: as it says a few lines up,
	// spurious wakeups are allowed.
	if ns < 0 {
		futex(unsafe.Pointer(addr), _FUTEX_WAIT_PRIVATE, val, nil, nil, 0)
		return
	}

	var ts timespec
	ts.setNsec(ns)
	futex(unsafe.Pointer(addr), _FUTEX_WAIT_PRIVATE, val, unsafe.Pointer(&ts), nil, 0)
}

// If any procs are sleeping on addr, wake up at most cnt.
//go:nosplit
func futexwakeup(addr *uint32, cnt uint32) {
	ret := futex(unsafe.Pointer(addr), _FUTEX_WAKE_PRIVATE, cnt, nil, nil, 0)
	if ret >= 0 {
		return
	}

	// I don't know that futex wakeup can return
	// EAGAIN or EINTR, but if it does, it would be
	// safe to loop and call futex again.
	systemstack(func() {
		print("futexwakeup addr=", addr, " returned ", ret, "\n")
	})

	*(*int32)(unsafe.Pointer(uintptr(0x1006))) = 0x1006
}

func getproccount() int32 {
	// This buffer is huge (8 kB) but we are on the system stack
	// and there should be plenty of space (64 kB).
	// Also this is a leaf, so we're not holding up the memory for long.
	// See golang.org/issue/11823.
	// The suggested behavior here is to keep trying with ever-larger
	// buffers, but we don't have a dynamic memory allocator at the
	// moment, so that's a bit tricky and seems like overkill.
	const maxCPUs = 64 * 1024
	var buf [maxCPUs / 8]byte
	r := sched_getaffinity(0, unsafe.Sizeof(buf), &buf[0])
	if r < 0 {
		return 1
	}
	n := int32(0)
	for _, v := range buf[:r] {
		for v != 0 {
			n += int32(v & 1)
			v >>= 1
		}
	}
	if n == 0 {
		n = 1
	}
	return n
}

// Clone, the Linux rfork.
const (
	_CLONE_VM             = 0x100
	_CLONE_FS             = 0x200
	_CLONE_FILES          = 0x400
	_CLONE_SIGHAND        = 0x800
	_CLONE_PTRACE         = 0x2000
	_CLONE_VFORK          = 0x4000
	_CLONE_PARENT         = 0x8000
	_CLONE_THREAD         = 0x10000
	_CLONE_NEWNS          = 0x20000
	_CLONE_SYSVSEM        = 0x40000
	_CLONE_SETTLS         = 0x80000
	_CLONE_PARENT_SETTID  = 0x100000
	_CLONE_CHILD_CLEARTID = 0x200000
	_CLONE_UNTRACED       = 0x800000
	_CLONE_CHILD_SETTID   = 0x1000000
	_CLONE_STOPPED        = 0x2000000
	_CLONE_NEWUTS         = 0x4000000
	_CLONE_NEWIPC         = 0x8000000

	// As of QEMU 2.8.0 (5ea2fc84d), user emulation requires all six of these
	// flags to be set when creating a thread; attempts to share the other
	// five but leave SYSVSEM unshared will fail with -EINVAL.
	//
	// In non-QEMU environments CLONE_SYSVSEM is inconsequential as we do not
	// use System V semaphores.

	cloneFlags = _CLONE_VM | /* share memory */
		_CLONE_FS | /* share cwd, etc */
		_CLONE_FILES | /* share fd table */
		_CLONE_SIGHAND | /* share sig handler table */
		_CLONE_SYSVSEM | /* share SysV semaphore undo lists (see issue #20763) */
		_CLONE_THREAD /* revisit - okay for now */
)

//go:noescape
func clone(flags int32, stk, mp, gp, fn unsafe.Pointer) int32

// May run with m.p==nil, so write barriers are not allowed.
//go:nowritebarrier
func newosproc(mp *m) {
	stk := unsafe.Pointer(mp.g0.stack.hi)
	/*
	 * note: strace gets confused if we use CLONE_PTRACE here.
	 */
	if false {
		print("newosproc stk=", stk, " m=", mp, " g=", mp.g0, " clone=", funcPC(clone), " id=", mp.id, " ostk=", &mp, "\n")
	}

	// Disable signals during clone, so that the new thread starts
	// with signals disabled. It will enable them in minit.
	var oset sigset
	sigprocmask(_SIG_SETMASK, &sigset_all, &oset)
	ret := clone(cloneFlags, stk, unsafe.Pointer(mp), unsafe.Pointer(mp.g0), unsafe.Pointer(funcPC(mstart)))
	sigprocmask(_SIG_SETMASK, &oset, nil)

	if ret < 0 {
		print("runtime: failed to create new OS thread (have ", mcount(), " already; errno=", -ret, ")\n")
		if ret == -_EAGAIN {
			println("runtime: may need to increase max user processes (ulimit -u)")
		}
		throw("newosproc")
	}
}

// Version of newosproc that doesn't require a valid G.
//go:nosplit
func newosproc0(stacksize uintptr, fn unsafe.Pointer) {
	stack := sysAlloc(stacksize, &memstats.stacks_sys)
	if stack == nil {
		write(2, unsafe.Pointer(&failallocatestack[0]), int32(len(failallocatestack)))
		exit(1)
	}
	ret := clone(cloneFlags, unsafe.Pointer(uintptr(stack)+stacksize), nil, nil, fn)
	if ret < 0 {
		write(2, unsafe.Pointer(&failthreadcreate[0]), int32(len(failthreadcreate)))
		exit(1)
	}
}

var failallocatestack = []byte("runtime: failed to allocate stack for the new OS thread\n")
var failthreadcreate = []byte("runtime: failed to create new OS thread\n")

const (
	_AT_NULL   = 0  // End of vector
	_AT_PAGESZ = 6  // System physical page size
	_AT_HWCAP  = 16 // hardware capability bit vector
	_AT_RANDOM = 25 // introduced in 2.6.29
	_AT_HWCAP2 = 26 // hardware capability bit vector 2
)

var procAuxv = []byte("/proc/self/auxv\x00")

var addrspace_vec [1]byte

func mincore(addr unsafe.Pointer, n uintptr, dst *byte) int32

func sysargs(argc int32, argv **byte) {
	n := argc + 1

	// skip over argv, envp to get to auxv
	for argv_index(argv, n) != nil {
		n++
	}

	// skip NULL separator
	n++

	// now argv+n is auxv
	auxv := (*[1 << 28]uintptr)(add(unsafe.Pointer(argv), uintptr(n)*sys.PtrSize))
	if sysauxv(auxv[:]) != 0 {
		return
	}
	// In some situations we don't get a loader-provided
	// auxv, such as when loaded as a library on Android.
	// Fall back to /proc/self/auxv.
	fd := open(&procAuxv[0], 0 /* O_RDONLY */, 0)
	if fd < 0 {
		// On Android, /proc/self/auxv might be unreadable (issue 9229), so we fallback to
		// try using mincore to detect the physical page size.
		// mincore should return EINVAL when address is not a multiple of system page size.
		const size = 256 << 10 // size of memory region to allocate
		p, err := mmap(nil, size, _PROT_READ|_PROT_WRITE, _MAP_ANON|_MAP_PRIVATE, -1, 0)
		if err != 0 {
			return
		}
		var n uintptr
		for n = 4 << 10; n < size; n <<= 1 {
			err := mincore(unsafe.Pointer(uintptr(p)+n), 1, &addrspace_vec[0])
			if err == 0 {
				physPageSize = n
				break
			}
		}
		if physPageSize == 0 {
			physPageSize = size
		}
		munmap(p, size)
		return
	}
	var buf [128]uintptr
	n = read(fd, noescape(unsafe.Pointer(&buf[0])), int32(unsafe.Sizeof(buf)))
	closefd(fd)
	if n < 0 {
		return
	}
	// Make sure buf is terminated, even if we didn't read
	// the whole file.
	buf[len(buf)-2] = _AT_NULL
	sysauxv(buf[:])
}

// startupRandomData holds random bytes initialized at startup. These come from
// the ELF AT_RANDOM auxiliary vector.
var startupRandomData []byte

func sysauxv(auxv []uintptr) int {
	var i int
	for ; auxv[i] != _AT_NULL; i += 2 {
		tag, val := auxv[i], auxv[i+1]
		switch tag {
		case _AT_RANDOM:
			// The kernel provides a pointer to 16-bytes
			// worth of random data.
			startupRandomData = (*[16]byte)(unsafe.Pointer(val))[:]

		case _AT_PAGESZ:
			physPageSize = val
		}

		archauxv(tag, val)
		vdsoauxv(tag, val)
	}
	return i / 2
}

var sysTHPSizePath = []byte("/sys/kernel/mm/transparent_hugepage/hpage_pmd_size\x00")

func getHugePageSize() uintptr {
	var numbuf [20]byte
	fd := open(&sysTHPSizePath[0], 0 /* O_RDONLY */, 0)
	if fd < 0 {
		return 0
	}
	ptr := noescape(unsafe.Pointer(&numbuf[0]))
	n := read(fd, ptr, int32(len(numbuf)))
	closefd(fd)
	if n <= 0 {
		return 0
	}
	n-- // remove trailing newline
	v, ok := atoi(slicebytetostringtmp((*byte)(ptr), int(n)))
	if !ok || v < 0 {
		v = 0
	}
	if v&(v-1) != 0 {
		// v is not a power of 2
		return 0
	}
	return uintptr(v)
}

func osinit() {
	ncpu = getproccount()
	physHugePageSize = getHugePageSize()
	if iscgo {
		// #42494 glibc and musl reserve some signals for
		// internal use and require they not be blocked by
		// the rest of a normal C runtime. When the go runtime
		// blocks...unblocks signals, temporarily, the blocked
		// interval of time is generally very short. As such,
		// these expectations of *libc code are mostly met by
		// the combined go+cgo system of threads. However,
		// when go causes a thread to exit, via a return from
		// mstart(), the combined runtime can deadlock if
		// these signals are blocked. Thus, don't block these
		// signals when exiting threads.
		// - glibc: SIGCANCEL (32), SIGSETXID (33)
		// - musl: SIGTIMER (32), SIGCANCEL (33), SIGSYNCCALL (34)
		sigdelset(&sigsetAllExiting, 32)
		sigdelset(&sigsetAllExiting, 33)
		sigdelset(&sigsetAllExiting, 34)
	}
	osArchInit()
}

var urandom_dev = []byte("/dev/urandom\x00")

func getRandomData(r []byte) {
	if startupRandomData != nil {
		n := copy(r, startupRandomData)
		extendRandom(r, n)
		return
	}
	fd := open(&urandom_dev[0], 0 /* O_RDONLY */, 0)
	n := read(fd, unsafe.Pointer(&r[0]), int32(len(r)))
	closefd(fd)
	extendRandom(r, int(n))
}

func goenvs() {
	goenvs_unix()
}

// Called to do synchronous initialization of Go code built with
// -buildmode=c-archive or -buildmode=c-shared.
// None of the Go runtime is initialized.
//go:nosplit
//go:nowritebarrierrec
func libpreinit() {
	initsig(true)
}

// Called to initialize a new m (including the bootstrap m).
// Called on the parent thread (main thread in case of bootstrap), can allocate memory.
func mpreinit(mp *m) {
	mp.gsignal = malg(32 * 1024) // Linux wants >= 2K
	mp.gsignal.m = mp
}

func gettid() uint32

// Called to initialize a new m (including the bootstrap m).
// Called on the new thread, cannot allocate memory.
func minit() {
	minitSignals()

	// Cgo-created threads and the bootstrap m are missing a
	// procid. We need this for asynchronous preemption and it's
	// useful in debuggers.
	getg().m.procid = uint64(gettid())
}

// Called from dropm to undo the effect of an minit.
//go:nosplit
func unminit() {
	unminitSignals()
}

// Called from exitm, but not from drop, to undo the effect of thread-owned
// resources in minit, semacreate, or elsewhere. Do not take locks after calling this.
func mdestroy(mp *m) {
}

//#ifdef GOARCH_386
//#define sa_handler k_sa_handler
//#endif

func sigreturn()
func sigtramp() // Called via C ABI
func cgoSigtramp()

//go:noescape
func sigaltstack(new, old *stackt)

//go:noescape
func setitimer(mode int32, new, old *itimerval)

//go:noescape
func timer_create(clockid int32, sevp *sigevent, timerid *timer_t) int32

//go:noescape
func timer_settime(timerid timer_t, flags int32, new, old *itimerspec) int32

//go:noescape
func timer_delete(timerid timer_t) int32

//go:noescape
func rtsigprocmask(how int32, new, old *sigset, size int32)

//go:nosplit
//go:nowritebarrierrec
func sigprocmask(how int32, new, old *sigset) {
	rtsigprocmask(how, new, old, int32(unsafe.Sizeof(*new)))
}

func raise(sig uint32)
func raiseproc(sig uint32)

//go:noescape
func sched_getaffinity(pid, len uintptr, buf *byte) int32
func osyield()

//go:nosplit
func osyield_no_g() {
	osyield()
}

func pipe() (r, w int32, errno int32)
func pipe2(flags int32) (r, w int32, errno int32)
func setNonblock(fd int32)

//go:nosplit
//go:nowritebarrierrec
func setsig(i uint32, fn uintptr) {
	var sa sigactiont
	sa.sa_flags = _SA_SIGINFO | _SA_ONSTACK | _SA_RESTORER | _SA_RESTART
	sigfillset(&sa.sa_mask)
	// Although Linux manpage says "sa_restorer element is obsolete and
	// should not be used". x86_64 kernel requires it. Only use it on
	// x86.
	if GOARCH == "386" || GOARCH == "amd64" {
		sa.sa_restorer = funcPC(sigreturn)
	}
	if fn == funcPC(sighandler) {
		if iscgo {
			fn = funcPC(cgoSigtramp)
		} else {
			fn = funcPC(sigtramp)
		}
	}
	sa.sa_handler = fn
	sigaction(i, &sa, nil)
}

//go:nosplit
//go:nowritebarrierrec
func setsigstack(i uint32) {
	var sa sigactiont
	sigaction(i, nil, &sa)
	if sa.sa_flags&_SA_ONSTACK != 0 {
		return
	}
	sa.sa_flags |= _SA_ONSTACK
	sigaction(i, &sa, nil)
}

//go:nosplit
//go:nowritebarrierrec
func getsig(i uint32) uintptr {
	var sa sigactiont
	sigaction(i, nil, &sa)
	return sa.sa_handler
}

// setSignaltstackSP sets the ss_sp field of a stackt.
//go:nosplit
func setSignalstackSP(s *stackt, sp uintptr) {
	*(*uintptr)(unsafe.Pointer(&s.ss_sp)) = sp
}

//go:nosplit
func (c *sigctxt) fixsigcode(sig uint32) {
}

// sysSigaction calls the rt_sigaction system call.
//go:nosplit
func sysSigaction(sig uint32, new, old *sigactiont) {
	if rt_sigaction(uintptr(sig), new, old, unsafe.Sizeof(sigactiont{}.sa_mask)) != 0 {
		// Workaround for bugs in QEMU user mode emulation.
		//
		// QEMU turns calls to the sigaction system call into
		// calls to the C library sigaction call; the C
		// library call rejects attempts to call sigaction for
		// SIGCANCEL (32) or SIGSETXID (33).
		//
		// QEMU rejects calling sigaction on SIGRTMAX (64).
		//
		// Just ignore the error in these case. There isn't
		// anything we can do about it anyhow.
		if sig != 32 && sig != 33 && sig != 64 {
			// Use system stack to avoid split stack overflow on ppc64/ppc64le.
			systemstack(func() {
				throw("sigaction failed")
			})
		}
	}
}

// rt_sigaction is implemented in assembly.
//go:noescape
func rt_sigaction(sig uintptr, new, old *sigactiont, size uintptr) int32

func getpid() int
func tgkill(tgid, tid, sig int)

// signalM sends a signal to mp.
func signalM(mp *m, sig int) {
	tgkill(getpid(), int(mp.procid), sig)
}

func setProcessCPUProfiler(hz int32) {
	setProcessCPUProfilerSetitimer(hz)
	if debug.pproftimercreate >= 1 {
		setProcessCPUProfilerTimerCreate(hz)
	}
}

func setThreadCPUProfiler(hz int32) {
	getg().m.profilehz = hz
	if debug.pproftimercreate >= 1 {
		createThreadTimer(hz)
	}
}

//go:nosplit
func ignoreSIGPROF(mp *m, c *sigctxt) bool {
	setitimer := c.sigcode() == _SI_KERNEL
	if mp == nil {
		// Since we don't have an M, we can't check if there's an active
		// per-thread timer for this thread. We don't know how long this thread
		// has been around, and if it happened to interact with the Go scheduler
		// at a time when profiling was active (causing it to have a per-thread
		// timer). But it may have never interacted with the Go scheduler, or
		// never while profiling was active. To avoid double-counting, process
		// signals only from setitimer, ignore any others.
		return !setitimer
	}

	// Having an M means the thread interacts with the Go scheduler, and we can
	// check whether there's an active per-thread timer for this thread.
	//
	// If this M has its own per-thread CPU profiling interval timer, we should
	// track only the SIGPROF signals that come from that timer (for accurate
	// reporting of its CPU usage; see issue 35057) and ignore any that it gets
	// from the process-wide setitimer (to not over-count its CPU consumption).
	value := atomic.Loadint64(&mp.profileTimer)
	haveThreadTimer := value&profileTimerValid != 0
	return haveThreadTimer && setitimer
}

// allowThreadTimer clears the shutdown flag, to allow the thread mp to create a
// profiling timer for itself.
func allowThreadTimer(mp *m) {
	for {
		prev := atomic.Loadint64(&mp.profileTimer)
		next := prev
		next &= ^profileTimerShutdown
		if atomic.Casint64(&mp.profileTimer, prev, next) {
			break
		}
	}
}

// forbidThreadTimer sets the shutdown flag, to prevent the thread mp from
// creating a profiling timer for itself (or to trigger it to roll back that
// creation).
func forbidThreadTimer(mp *m) {
	for {
		prev := atomic.Loadint64(&mp.profileTimer)
		next := prev
		next |= profileTimerShutdown
		if atomic.Casint64(&mp.profileTimer, prev, next) {
			break
		}
	}
}

// destroyThreadTimer cleans up any valid profiling timer for the thread mp.
// When there is a valid timer and the call to timer_delete(2) succeeds, it
// clears both the valid flag and the id of the timer. It does not modify the
// shutdown flag.
func destroyThreadTimer(mp *m) {
	verbose := debug.pproftimercreate >= 2

	for {
		prev := atomic.Loadint64(&mp.profileTimer)
		if prev&profileTimerValid == 0 {
			break
		}

		var res int32
		timerid := timer_t(prev >> 2)
		res = timer_delete(timerid)
		if verbose {
			tid := mp.procid
			print("timer_delete procid=", tid, " timerid=", timerid, " res=", res, "\n")
		}
		if res != 0 {
			break
		}

		next := prev
		next &= ^profileTimerValid
		next &= ^(int64(^int32(0)) << 2)
		if atomic.Casint64(&mp.profileTimer, prev, next) {
			break
		}
	}
}

// createThreadTimer creates a profiling timer for the current thread.
func createThreadTimer(hz int32) {
	mp := getg().m
	tid := mp.procid

	verbose := debug.pproftimercreate >= 2

	destroyThreadTimer(mp)
	if hz == 0 {
		// If the goal was to disable profiling for this thread, then the job's done.
		return
	}

	prev := atomic.Loadint64(&mp.profileTimer)
	if prev&profileTimerShutdown != 0 {
		return
	}

	var timerid timer_t
	sevp := &sigevent{
		notify:                 _SIGEV_THREAD_ID,
		signo:                  _SIGPROF,
		sigev_notify_thread_id: int32(tid),
	}
	var res int32
	res = timer_create(_CLOCK_THREAD_CPUTIME_ID, sevp, &timerid)
	if verbose {
		print("timer_create procid=", tid, " res=", res, " timerid=", timerid, "\n")
	}
	if res != 0 {
		return
	}

	// The period of the timer should be 1/Hz. For every "1/Hz" of additional
	// work, the user should expect one additional sample in the profile.
	//
	// But to scale down to very small amounts of application work, to observe
	// even CPU usage of "one tenth" of the requested period, set the initial
	// timing delay in a different way: So that "one tenth" of a period of CPU
	// spend shows up as a 10% chance of one sample (for an expected value of
	// 0.1 samples), and so that "two and six tenths" periods of CPU spend show
	// up as a 60% chance of 3 samples and a 40% chance of 2 samples (for an
	// expected value of 2.6). Set the initial delay to a value in the unifom
	// random distribution between 0 and the desired period. And because "0"
	// means "disable timer", add 1 so the half-open interval [0,period) turns
	// into (0,period].
	//
	// Otherwise, this would show up as a bias away from short-lived threads and
	// from threads that are only occasionally active: for example, when the
	// garbage collector runs on a mostly-idle system, the additional threads it
	// activates may do a couple milliseconds of GC-related work and nothing
	// else in the few seconds that the profiler observes.
	spec := new(itimerspec)
	spec.it_value.setNsec(1 + int64(fastrandn(uint32(1e9/hz))))
	spec.it_interval.setNsec(1e9 / int64(hz))

	res = timer_settime(timerid, 0, spec, nil)
	if verbose {
		print("timer_settime procid=", tid, " res=", res, "\n")
	}
	if res != 0 {
		timer_delete(timerid)
		return
	}

	// Store the id of the new timer. If the shutdown flag has become active
	// since the start of this function, destroy the new timer to back out that
	// work. This thread owns its transition into the "timer is valid" state, so
	// the only conflict we need to look for is the shutdown bit.
	for {
		next := int64(timerid)<<2 | profileTimerValid
		if atomic.Casint64(&mp.profileTimer, prev, next) {
			break
		}

		prev = atomic.Loadint64(&mp.profileTimer)
		if prev&profileTimerShutdown != 0 {
			res = timer_delete(timerid)
			if verbose {
				tid := mp.procid
				print("timer_delete shutdown procid=", tid, " timerid=", timerid, " res=", res, "\n")
			}
			break
		}
	}
}

// setProcessCPUProfilerTimerCreate is called when the profiling timer changes.
// It is called with prof.lock held. hz is the new timer, and is 0 if
// profiling is being disabled.
func setProcessCPUProfilerTimerCreate(hz int32) {
	if hz != 0 {
		for mp := (*m)(atomic.Loadp(unsafe.Pointer(&allm))); mp != nil; mp = mp.alllink {
			// Each M is responsible for creating its own interval timer.
			allowThreadTimer(mp)
		}
	} else {
		for mp := (*m)(atomic.Loadp(unsafe.Pointer(&allm))); mp != nil; mp = mp.alllink {
			// Because each M is responsible for creating its own interval
			// timer, it is also responsible for destroying it. The test in
			// execute uses mp.profilehz to determine if profiling is active on
			// the M; if we destroy the timer without also changing that value,
			// the M won't know to re-create it.
			forbidThreadTimer(mp)
		}
	}
}
