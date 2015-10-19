// Copyright 2009 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import _ "unsafe" // for go:linkname

//go:generate go run wincallback.go
//go:generate go run mkduff.go
//go:generate go run mkfastlog2table.go

var ticks struct {
	lock mutex
	pad  uint32 // ensure 8-byte alignment of val on 386
	val  uint64
}

var tls0 [8]uintptr // available storage for m0's TLS; not necessarily used; opaque to GC

// Note: Called by runtime/pprof in addition to runtime code.
func tickspersecond() int64 {
	r := int64(atomicload64(&ticks.val))
	if r != 0 {
		return r
	}
	lock(&ticks.lock)
	r = int64(ticks.val)
	if r == 0 {
		t0 := nanotime()
		c0 := cputicks()
		usleep(100 * 1000)
		t1 := nanotime()
		c1 := cputicks()
		if t1 == t0 {
			t1++
		}
		r = (c1 - c0) * 1000 * 1000 * 1000 / (t1 - t0)
		if r == 0 {
			r++
		}
		atomicstore64(&ticks.val, uint64(r))
	}
	unlock(&ticks.lock)
	return r
}

var envs []string
var argslice []string

//go:linkname syscall_runtime_envs syscall.runtime_envs
func syscall_runtime_envs() []string { return append([]string{}, envs...) }

//go:linkname os_runtime_args os.runtime_args
func os_runtime_args() []string { return append([]string{}, argslice...) }

// Keep a cached value to make gotraceback fast,
// since we call it on every call to gentraceback.
// The cached value is a uint32 in which the low bit
// is the "crash" setting and the top 31 bits are the
// gotraceback value.
var traceback_cache uint32 = 2 << 1

// The GOTRACEBACK environment variable controls the
// behavior of a Go program that is crashing and exiting.
//	GOTRACEBACK=0   suppress all tracebacks
//	GOTRACEBACK=1   default behavior - show tracebacks but exclude runtime frames
//	GOTRACEBACK=2   show tracebacks including runtime frames
//	GOTRACEBACK=crash   show tracebacks including runtime frames, then crash (core dump etc)
//go:nosplit
func gotraceback(crash *bool) int32 {
	_g_ := getg()
	if crash != nil {
		*crash = false
	}
	if _g_.m.traceback != 0 {
		return int32(_g_.m.traceback)
	}
	if crash != nil {
		*crash = traceback_cache&1 != 0
	}
	return int32(traceback_cache >> 1)
}

var (
	argc int32
	argv **byte
)

// nosplit for use in linux/386 startup linux_setup_vdso
//go:nosplit
func argv_index(argv **byte, i int32) *byte {
	return *(**byte)(add(unsafe.Pointer(argv), uintptr(i)*ptrSize))
}

func args(c int32, v **byte) {
	argc = c
	argv = v
	sysargs(c, v)
}

var (
	// TODO: Retire in favor of GOOS== checks.
	isplan9   int32
	issolaris int32
	iswindows int32
)

func goargs() {
	if GOOS == "windows" {
		return
	}

	argslice = make([]string, argc)
	for i := int32(0); i < argc; i++ {
		argslice[i] = gostringnocopy(argv_index(argv, i))
	}
}

func goenvs_unix() {
	// TODO(austin): ppc64 in dynamic linking mode doesn't
	// guarantee env[] will immediately follow argv.  Might cause
	// problems.
	n := int32(0)
	for argv_index(argv, argc+1+n) != nil {
		n++
	}

	envs = make([]string, n)
	for i := int32(0); i < n; i++ {
		envs[i] = gostring(argv_index(argv, argc+1+i))
	}
}

func environ() []string {
	return envs
}

// TODO: These should be locals in testAtomic64, but we don't 8-byte
// align stack variables on 386.
var test_z64, test_x64 uint64

func testAtomic64() {
	test_z64 = 42
	test_x64 = 0
	prefetcht0(uintptr(unsafe.Pointer(&test_z64)))
	prefetcht1(uintptr(unsafe.Pointer(&test_z64)))
	prefetcht2(uintptr(unsafe.Pointer(&test_z64)))
	prefetchnta(uintptr(unsafe.Pointer(&test_z64)))
	if cas64(&test_z64, test_x64, 1) {
		throw("cas64 failed")
	}
	if test_x64 != 0 {
		throw("cas64 failed")
	}
	test_x64 = 42
	if !cas64(&test_z64, test_x64, 1) {
		throw("cas64 failed")
	}
	if test_x64 != 42 || test_z64 != 1 {
		throw("cas64 failed")
	}
	if atomicload64(&test_z64) != 1 {
		throw("load64 failed")
	}
	atomicstore64(&test_z64, (1<<40)+1)
	if atomicload64(&test_z64) != (1<<40)+1 {
		throw("store64 failed")
	}
	if xadd64(&test_z64, (1<<40)+1) != (2<<40)+2 {
		throw("xadd64 failed")
	}
	if atomicload64(&test_z64) != (2<<40)+2 {
		throw("xadd64 failed")
	}
	if xchg64(&test_z64, (3<<40)+3) != (2<<40)+2 {
		throw("xchg64 failed")
	}
	if atomicload64(&test_z64) != (3<<40)+3 {
		throw("xchg64 failed")
	}
}

func check() {
	var (
		a     int8
		b     uint8
		c     int16
		d     uint16
		e     int32
		f     uint32
		g     int64
		h     uint64
		i, i1 float32
		j, j1 float64
		k, k1 unsafe.Pointer
		l     *uint16
		m     [4]byte
	)
	type x1t struct {
		x uint8
	}
	type y1t struct {
		x1 x1t
		y  uint8
	}
	var x1 x1t
	var y1 y1t

	if unsafe.Sizeof(a) != 1 {
		throw("bad a")
	}
	if unsafe.Sizeof(b) != 1 {
		throw("bad b")
	}
	if unsafe.Sizeof(c) != 2 {
		throw("bad c")
	}
	if unsafe.Sizeof(d) != 2 {
		throw("bad d")
	}
	if unsafe.Sizeof(e) != 4 {
		throw("bad e")
	}
	if unsafe.Sizeof(f) != 4 {
		throw("bad f")
	}
	if unsafe.Sizeof(g) != 8 {
		throw("bad g")
	}
	if unsafe.Sizeof(h) != 8 {
		throw("bad h")
	}
	if unsafe.Sizeof(i) != 4 {
		throw("bad i")
	}
	if unsafe.Sizeof(j) != 8 {
		throw("bad j")
	}
	if unsafe.Sizeof(k) != ptrSize {
		throw("bad k")
	}
	if unsafe.Sizeof(l) != ptrSize {
		throw("bad l")
	}
	if unsafe.Sizeof(x1) != 1 {
		throw("bad unsafe.Sizeof x1")
	}
	if unsafe.Offsetof(y1.y) != 1 {
		throw("bad offsetof y1.y")
	}
	if unsafe.Sizeof(y1) != 2 {
		throw("bad unsafe.Sizeof y1")
	}

	if timediv(12345*1000000000+54321, 1000000000, &e) != 12345 || e != 54321 {
		throw("bad timediv")
	}

	var z uint32
	z = 1
	if !cas(&z, 1, 2) {
		throw("cas1")
	}
	if z != 2 {
		throw("cas2")
	}

	z = 4
	if cas(&z, 5, 6) {
		throw("cas3")
	}
	if z != 4 {
		throw("cas4")
	}

	z = 0xffffffff
	if !cas(&z, 0xffffffff, 0xfffffffe) {
		throw("cas5")
	}
	if z != 0xfffffffe {
		throw("cas6")
	}

	k = unsafe.Pointer(uintptr(0xfedcb123))
	if ptrSize == 8 {
		k = unsafe.Pointer(uintptr(unsafe.Pointer(k)) << 10)
	}
	if casp(&k, nil, nil) {
		throw("casp1")
	}
	k1 = add(k, 1)
	if !casp(&k, k, k1) {
		throw("casp2")
	}
	if k != k1 {
		throw("casp3")
	}

	m = [4]byte{1, 1, 1, 1}
	atomicor8(&m[1], 0xf0)
	if m[0] != 1 || m[1] != 0xf1 || m[2] != 1 || m[3] != 1 {
		throw("atomicor8")
	}

	*(*uint64)(unsafe.Pointer(&j)) = ^uint64(0)
	if j == j {
		throw("float64nan")
	}
	if !(j != j) {
		throw("float64nan1")
	}

	*(*uint64)(unsafe.Pointer(&j1)) = ^uint64(1)
	if j == j1 {
		throw("float64nan2")
	}
	if !(j != j1) {
		throw("float64nan3")
	}

	*(*uint32)(unsafe.Pointer(&i)) = ^uint32(0)
	if i == i {
		throw("float32nan")
	}
	if i == i {
		throw("float32nan1")
	}

	*(*uint32)(unsafe.Pointer(&i1)) = ^uint32(1)
	if i == i1 {
		throw("float32nan2")
	}
	if i == i1 {
		throw("float32nan3")
	}

	testAtomic64()

	if _FixedStack != round2(_FixedStack) {
		throw("FixedStack is not power-of-2")
	}
}

type dbgVar struct {
	name  string
	value *int32
}

// Holds variables parsed from GODEBUG env var,
// except for "memprofilerate" since there is an
// existing int var for that value, which may
// already have an initial value.
var debug struct {
	allocfreetrace    int32
	efence            int32
	gccheckmark       int32
	gcpacertrace      int32
	gcshrinkstackoff  int32
	gcstackbarrieroff int32
	gcstackbarrierall int32
	gcstoptheworld    int32
	gctrace           int32
	invalidptr        int32
	sbrk              int32
	scavenge          int32
	scheddetail       int32
	schedtrace        int32
	wbshadow          int32
}

var dbgvars = []dbgVar{
	{"allocfreetrace", &debug.allocfreetrace},
	{"efence", &debug.efence},
	{"gccheckmark", &debug.gccheckmark},
	{"gcpacertrace", &debug.gcpacertrace},
	{"gcshrinkstackoff", &debug.gcshrinkstackoff},
	{"gcstackbarrieroff", &debug.gcstackbarrieroff},
	{"gcstackbarrierall", &debug.gcstackbarrierall},
	{"gcstoptheworld", &debug.gcstoptheworld},
	{"gctrace", &debug.gctrace},
	{"invalidptr", &debug.invalidptr},
	{"sbrk", &debug.sbrk},
	{"scavenge", &debug.scavenge},
	{"scheddetail", &debug.scheddetail},
	{"schedtrace", &debug.schedtrace},
	{"wbshadow", &debug.wbshadow},
}

func parsedebugvars() {
	// defaults
	debug.invalidptr = 1

	for p := gogetenv("GODEBUG"); p != ""; {
		field := ""
		i := index(p, ",")
		if i < 0 {
			field, p = p, ""
		} else {
			field, p = p[:i], p[i+1:]
		}
		i = index(field, "=")
		if i < 0 {
			continue
		}
		key, value := field[:i], field[i+1:]

		// Update MemProfileRate directly here since it
		// is int, not int32, and should only be updated
		// if specified in GODEBUG.
		if key == "memprofilerate" {
			MemProfileRate = atoi(value)
		} else {
			for _, v := range dbgvars {
				if v.name == key {
					*v.value = int32(atoi(value))
				}
			}
		}
	}

	switch p := gogetenv("GOTRACEBACK"); p {
	case "":
		traceback_cache = 1 << 1
	case "crash":
		traceback_cache = 2<<1 | 1
	default:
		traceback_cache = uint32(atoi(p)) << 1
	}
	// when C owns the process, simply exit'ing the process on fatal errors
	// and panics is surprising. Be louder and abort instead.
	if islibrary || isarchive {
		traceback_cache |= 1
	}

	if debug.gcstackbarrierall > 0 {
		firstStackBarrierOffset = 0
	}
}

// Poor mans 64-bit division.
// This is a very special function, do not use it if you are not sure what you are doing.
// int64 division is lowered into _divv() call on 386, which does not fit into nosplit functions.
// Handles overflow in a time-specific manner.
//go:nosplit
func timediv(v int64, div int32, rem *int32) int32 {
	res := int32(0)
	for bit := 30; bit >= 0; bit-- {
		if v >= int64(div)<<uint(bit) {
			v = v - (int64(div) << uint(bit))
			res += 1 << uint(bit)
		}
	}
	if v >= int64(div) {
		if rem != nil {
			*rem = 0
		}
		return 0x7fffffff
	}
	if rem != nil {
		*rem = int32(v)
	}
	return res
}

// Helpers for Go. Must be NOSPLIT, must only call NOSPLIT functions, and must not block.

//go:nosplit
func acquirem() *m {
	_g_ := getg()
	_g_.m.locks++
	return _g_.m
}

//go:nosplit
func releasem(mp *m) {
	_g_ := getg()
	mp.locks--
	if mp.locks == 0 && _g_.preempt {
		// restore the preemption request in case we've cleared it in newstack
		_g_.stackguard0 = stackPreempt
	}
}

//go:nosplit
func gomcache() *mcache {
	return getg().m.mcache
}

//go:linkname reflect_typelinks reflect.typelinks
//go:nosplit
func reflect_typelinks() [][]*_type {
	ret := [][]*_type{firstmoduledata.typelinks}
	for datap := firstmoduledata.next; datap != nil; datap = datap.next {
		ret = append(ret, datap.typelinks)
	}
	return ret
}

/*
 * defined constants
 */
const (
	// G status
	//
	// If you add to this list, add to the list
	// of "okay during garbage collection" status
	// in mgcmark.go too.
	_Gidle            = iota // 0
	_Grunnable               // 1 runnable and on a run queue
	_Grunning                // 2
	_Gsyscall                // 3
	_Gwaiting                // 4
	_Gmoribund_unused        // 5 currently unused, but hardcoded in gdb scripts
	_Gdead                   // 6
	_Genqueue                // 7 Only the Gscanenqueue is used.
	_Gcopystack              // 8 in this state when newstack is moving the stack
	// the following encode that the GC is scanning the stack and what to do when it is done
	_Gscan = 0x1000 // atomicstatus&~Gscan = the non-scan state,
	// _Gscanidle =     _Gscan + _Gidle,      // Not used. Gidle only used with newly malloced gs
	_Gscanrunnable = _Gscan + _Grunnable //  0x1001 When scanning completes make Grunnable (it is already on run queue)
	_Gscanrunning  = _Gscan + _Grunning  //  0x1002 Used to tell preemption newstack routine to scan preempted stack.
	_Gscansyscall  = _Gscan + _Gsyscall  //  0x1003 When scanning completes make it Gsyscall
	_Gscanwaiting  = _Gscan + _Gwaiting  //  0x1004 When scanning completes make it Gwaiting
	// _Gscanmoribund_unused,               //  not possible
	// _Gscandead,                          //  not possible
	_Gscanenqueue = _Gscan + _Genqueue //  When scanning completes make it Grunnable and put on runqueue
)

const (
	// P status
	_Pidle    = iota
	_Prunning // Only this P is allowed to change from _Prunning.
	_Psyscall
	_Pgcstop
	_Pdead
)

// The next line makes 'go generate' write the zgen_*.go files with
// per-OS and per-arch information, including constants
// named goos_$GOOS and goarch_$GOARCH for every
// known GOOS and GOARCH. The constant is 1 on the
// current system, 0 otherwise; multiplying by them is
// useful for defining GOOS- or GOARCH-specific constants.
//go:generate go run gengoos.go

type mutex struct {
	// Futex-based impl treats it as uint32 key,
	// while sema-based impl as M* waitm.
	// Used to be a union, but unions break precise GC.
	key uintptr
}

type note struct {
	// Futex-based impl treats it as uint32 key,
	// while sema-based impl as M* waitm.
	// Used to be a union, but unions break precise GC.
	key uintptr
}

type _string struct {
	str *byte
	len int
}

type funcval struct {
	fn uintptr
	// variable-size, fn-specific data here
}

type iface struct {
	tab  *itab
	data unsafe.Pointer
}

type eface struct {
	_type *_type
	data  unsafe.Pointer
}

// The guintptr, muintptr, and puintptr are all used to bypass write barriers.
// It is particularly important to avoid write barriers when the current P has
// been released, because the GC thinks the world is stopped, and an
// unexpected write barrier would not be synchronized with the GC,
// which can lead to a half-executed write barrier that has marked the object
// but not queued it. If the GC skips the object and completes before the
// queuing can occur, it will incorrectly free the object.
//
// We tried using special assignment functions invoked only when not
// holding a running P, but then some updates to a particular memory
// word went through write barriers and some did not. This breaks the
// write barrier shadow checking mode, and it is also scary: better to have
// a word that is completely ignored by the GC than to have one for which
// only a few updates are ignored.
//
// Gs, Ms, and Ps are always reachable via true pointers in the
// allgs, allm, and allp lists or (during allocation before they reach those lists)
// from stack variables.

// A guintptr holds a goroutine pointer, but typed as a uintptr
// to bypass write barriers. It is used in the Gobuf goroutine state
// and in scheduling lists that are manipulated without a P.
//
// The Gobuf.g goroutine pointer is almost always updated by assembly code.
// In one of the few places it is updated by Go code - func save - it must be
// treated as a uintptr to avoid a write barrier being emitted at a bad time.
// Instead of figuring out how to emit the write barriers missing in the
// assembly manipulation, we change the type of the field to uintptr,
// so that it does not require write barriers at all.
//
// Goroutine structs are published in the allg list and never freed.
// That will keep the goroutine structs from being collected.
// There is never a time that Gobuf.g's contain the only references
// to a goroutine: the publishing of the goroutine in allg comes first.
// Goroutine pointers are also kept in non-GC-visible places like TLS,
// so I can't see them ever moving. If we did want to start moving data
// in the GC, we'd need to allocate the goroutine structs from an
// alternate arena. Using guintptr doesn't make that problem any worse.
type guintptr uintptr

func (gp guintptr) ptr() *g   { return (*g)(unsafe.Pointer(gp)) }
func (gp *guintptr) set(g *g) { *gp = guintptr(unsafe.Pointer(g)) }
func (gp *guintptr) cas(old, new guintptr) bool {
	return casuintptr((*uintptr)(unsafe.Pointer(gp)), uintptr(old), uintptr(new))
}

type puintptr uintptr

func (pp puintptr) ptr() *p   { return (*p)(unsafe.Pointer(pp)) }
func (pp *puintptr) set(p *p) { *pp = puintptr(unsafe.Pointer(p)) }

type muintptr uintptr

func (mp muintptr) ptr() *m   { return (*m)(unsafe.Pointer(mp)) }
func (mp *muintptr) set(m *m) { *mp = muintptr(unsafe.Pointer(m)) }

type gobuf struct {
	// The offsets of sp, pc, and g are known to (hard-coded in) libmach.
	sp   uintptr
	pc   uintptr
	g    guintptr
	ctxt unsafe.Pointer // this has to be a pointer so that gc scans it
	ret  uintreg
	lr   uintptr
	bp   uintptr // for GOEXPERIMENT=framepointer
}

// Known to compiler.
// Changes here must also be made in src/cmd/internal/gc/select.go's selecttype.
type sudog struct {
	g           *g
	selectdone  *uint32
	next        *sudog
	prev        *sudog
	elem        unsafe.Pointer // data element
	releasetime int64
	nrelease    int32  // -1 for acquire
	waitlink    *sudog // g.waiting list
}

type gcstats struct {
	// the struct must consist of only uint64's,
	// because it is casted to uint64[].
	nhandoff    uint64
	nhandoffcnt uint64
	nprocyield  uint64
	nosyield    uint64
	nsleep      uint64
}

type libcall struct {
	fn   uintptr
	n    uintptr // number of parameters
	args uintptr // parameters
	r1   uintptr // return values
	r2   uintptr
	err  uintptr // error number
}

// describes how to handle callback
type wincallbackcontext struct {
	gobody       unsafe.Pointer // go function to call
	argsize      uintptr        // callback arguments size (in bytes)
	restorestack uintptr        // adjust stack on return by (in bytes) (386 only)
	cleanstack   bool
}

// Stack describes a Go execution stack.
// The bounds of the stack are exactly [lo, hi),
// with no implicit data structures on either side.
type stack struct {
	lo uintptr
	hi uintptr
}

// stkbar records the state of a G's stack barrier.
type stkbar struct {
	savedLRPtr uintptr // location overwritten by stack barrier PC
	savedLRVal uintptr // value overwritten at savedLRPtr
}

type g struct {
	// Stack parameters.
	// stack describes the actual stack memory: [stack.lo, stack.hi).
	// stackguard0 is the stack pointer compared in the Go stack growth prologue.
	// It is stack.lo+StackGuard normally, but can be StackPreempt to trigger a preemption.
	// stackguard1 is the stack pointer compared in the C stack growth prologue.
	// It is stack.lo+StackGuard on g0 and gsignal stacks.
	// It is ~0 on other goroutine stacks, to trigger a call to morestackc (and crash).
	stack       stack   // offset known to runtime/cgo
	stackguard0 uintptr // offset known to liblink
	stackguard1 uintptr // offset known to liblink

	_panic         *_panic // innermost panic - offset known to liblink
	_defer         *_defer // innermost defer
	m              *m      // current m; offset known to arm liblink
	stackAlloc     uintptr // stack allocation is [stack.lo,stack.lo+stackAlloc)
	sched          gobuf
	syscallsp      uintptr        // if status==Gsyscall, syscallsp = sched.sp to use during gc
	syscallpc      uintptr        // if status==Gsyscall, syscallpc = sched.pc to use during gc
	stkbar         []stkbar       // stack barriers, from low to high
	stkbarPos      uintptr        // index of lowest stack barrier not hit
	stktopsp       uintptr        // expected sp at top of stack, to check in traceback
	param          unsafe.Pointer // passed parameter on wakeup
	atomicstatus   uint32
	stackLock      uint32 // sigprof/scang lock; TODO: fold in to atomicstatus
	goid           int64
	waitsince      int64  // approx time when the g become blocked
	waitreason     string // if status==Gwaiting
	schedlink      guintptr
	preempt        bool   // preemption signal, duplicates stackguard0 = stackpreempt
	paniconfault   bool   // panic (instead of crash) on unexpected fault address
	preemptscan    bool   // preempted g does scan for gc
	gcscandone     bool   // g has scanned stack; protected by _Gscan bit in status
	gcscanvalid    bool   // false at start of gc cycle, true if G has not run since last scan
	throwsplit     bool   // must not split stack
	raceignore     int8   // ignore race detection events
	sysblocktraced bool   // StartTrace has emitted EvGoInSyscall about this goroutine
	sysexitticks   int64  // cputicks when syscall has returned (for tracing)
	sysexitseq     uint64 // trace seq when syscall has returned (for tracing)
	lockedm        *m
	sig            uint32
	writebuf       []byte
	sigcode0       uintptr
	sigcode1       uintptr
	sigpc          uintptr
	gopc           uintptr // pc of go statement that created this goroutine
	startpc        uintptr // pc of goroutine function
	racectx        uintptr
	waiting        *sudog // sudog structures this g is waiting on (that have a valid elem ptr)

	// Per-G gcController state

	// gcAssistBytes is this G's GC assist credit in terms of
	// bytes allocated. If this is positive, then the G has credit
	// to allocate gcAssistBytes bytes without assisting. If this
	// is negative, then the G must correct this by performing
	// scan work. We track this in bytes to make it fast to update
	// and check for debt in the malloc hot path. The assist ratio
	// determines how this corresponds to scan work debt.
	gcAssistBytes int64
}

type mts struct {
	tv_sec  int64
	tv_nsec int64
}

type mscratch struct {
	v [6]uintptr
}

type m struct {
	g0      *g     // goroutine with scheduling stack
	morebuf gobuf  // gobuf arg to morestack
	divmod  uint32 // div/mod denominator for arm - known to liblink

	// Fields not known to debuggers.
	procid        uint64     // for debuggers, but offset not hard-coded
	gsignal       *g         // signal-handling g
	sigmask       [4]uintptr // storage for saved signal mask
	tls           [4]uintptr // thread-local storage (for x86 extern register)
	mstartfn      func()
	curg          *g       // current running goroutine
	caughtsig     guintptr // goroutine running during fatal signal
	p             puintptr // attached p for executing go code (nil if not executing go code)
	nextp         puintptr
	id            int32
	mallocing     int32
	throwing      int32
	preemptoff    string // if != "", keep curg running on this m
	locks         int32
	softfloat     int32
	dying         int32
	profilehz     int32
	helpgc        int32
	spinning      bool // m is out of work and is actively looking for work
	blocked       bool // m is blocked on a note
	inwb          bool // m is executing a write barrier
	printlock     int8
	fastrand      uint32
	ncgocall      uint64 // number of cgo calls in total
	ncgo          int32  // number of cgo calls currently in progress
	park          note
	alllink       *m // on allm
	schedlink     muintptr
	machport      uint32 // return address for mach ipc (os x)
	mcache        *mcache
	lockedg       *g
	createstack   [32]uintptr // stack that created this thread.
	freglo        [16]uint32  // d[i] lsb and f[i]
	freghi        [16]uint32  // d[i] msb and f[i+16]
	fflag         uint32      // floating point compare flags
	locked        uint32      // tracking for lockosthread
	nextwaitm     uintptr     // next m waiting for lock
	waitsema      uintptr     // semaphore for parking on locks
	waitsemacount uint32
	waitsemalock  uint32
	gcstats       gcstats
	needextram    bool
	traceback     uint8
	waitunlockf   unsafe.Pointer // todo go func(*g, unsafe.pointer) bool
	waitlock      unsafe.Pointer
	waittraceev   byte
	waittraceskip int
	startingtrace bool
	syscalltick   uint32
	//#ifdef GOOS_windows
	thread uintptr // thread handle
	// these are here because they are too large to be on the stack
	// of low-level NOSPLIT functions.
	libcall   libcall
	libcallpc uintptr // for cpu profiler
	libcallsp uintptr
	libcallg  guintptr
	syscall   libcall // stores syscall parameters on windows
	//#endif
	//#ifdef GOOS_solaris
	perrno *int32 // pointer to tls errno
	// these are here because they are too large to be on the stack
	// of low-level NOSPLIT functions.
	//LibCall	libcall;
	ts      mts
	scratch mscratch
	//#endif
	//#ifdef GOOS_plan9
	notesig *int8
	errstr  *byte
	//#endif
}

type p struct {
	lock mutex

	id          int32
	status      uint32 // one of pidle/prunning/...
	link        puintptr
	schedtick   uint32   // incremented on every scheduler call
	syscalltick uint32   // incremented on every system call
	m           muintptr // back-link to associated m (nil if idle)
	mcache      *mcache

	deferpool    [5][]*_defer // pool of available defer structs of different sizes (see panic.go)
	deferpoolbuf [5][32]*_defer

	// Cache of goroutine ids, amortizes accesses to runtime·sched.goidgen.
	goidcache    uint64
	goidcacheend uint64

	// Queue of runnable goroutines. Accessed without lock.
	runqhead uint32
	runqtail uint32
	runq     [256]*g
	// runnext, if non-nil, is a runnable G that was ready'd by
	// the current G and should be run next instead of what's in
	// runq if there's time remaining in the running G's time
	// slice. It will inherit the time left in the current time
	// slice. If a set of goroutines is locked in a
	// communicate-and-wait pattern, this schedules that set as a
	// unit and eliminates the (potentially large) scheduling
	// latency that otherwise arises from adding the ready'd
	// goroutines to the end of the run queue.
	runnext guintptr

	// Available G's (status == Gdead)
	gfree    *g
	gfreecnt int32

	sudogcache []*sudog
	sudogbuf   [128]*sudog

	tracebuf *traceBuf

	palloc persistentAlloc // per-P to avoid mutex

	// Per-P GC state
	gcAssistTime     int64 // Nanoseconds in assistAlloc
	gcBgMarkWorker   *g
	gcMarkWorkerMode gcMarkWorkerMode

	// gcw is this P's GC work buffer cache. The work buffer is
	// filled by write barriers, drained by mutator assists, and
	// disposed on certain GC state transitions.
	gcw gcWork

	runSafePointFn uint32 // if 1, run sched.safePointFn at next safe point

	pad [64]byte
}

const (
	// The max value of GOMAXPROCS.
	// There are no fundamental restrictions on the value.
	_MaxGomaxprocs = 1 << 8
)

type schedt struct {
	lock mutex

	goidgen uint64

	midle        muintptr // idle m's waiting for work
	nmidle       int32    // number of idle m's waiting for work
	nmidlelocked int32    // number of locked m's waiting for work
	mcount       int32    // number of m's that have been created
	maxmcount    int32    // maximum number of m's allowed (or die)

	pidle      puintptr // idle p's
	npidle     uint32
	nmspinning uint32 // limited to [0, 2^31-1]

	// Global runnable queue.
	runqhead guintptr
	runqtail guintptr
	runqsize int32

	// Global cache of dead G's.
	gflock mutex
	gfree  *g
	ngfree int32

	// Central cache of sudog structs.
	sudoglock  mutex
	sudogcache *sudog

	// Central pool of available defer structs of different sizes.
	deferlock mutex
	deferpool [5]*_defer

	gcwaiting  uint32 // gc is waiting to run
	stopwait   int32
	stopnote   note
	sysmonwait uint32
	sysmonnote note
	lastpoll   uint64

	// safepointFn should be called on each P at the next GC
	// safepoint if p.runSafePointFn is set.
	safePointFn   func(*p)
	safePointWait int32
	safePointNote note

	profilehz int32 // cpu profiling rate

	procresizetime int64 // nanotime() of last change to gomaxprocs
	totaltime      int64 // ∫gomaxprocs dt up to procresizetime
}

// The m->locked word holds two pieces of state counting active calls to LockOSThread/lockOSThread.
// The low bit (LockExternal) is a boolean reporting whether any LockOSThread call is active.
// External locks are not recursive; a second lock is silently ignored.
// The upper bits of m->locked record the nesting depth of calls to lockOSThread
// (counting up by LockInternal), popped by unlockOSThread (counting down by LockInternal).
// Internal locks can be recursive. For instance, a lock for cgo can occur while the main
// goroutine is holding the lock during the initialization phase.
const (
	_LockExternal = 1
	_LockInternal = 2
)

type sigtabtt struct {
	flags int32
	name  *int8
}

const (
	_SigNotify   = 1 << iota // let signal.Notify have signal, even if from kernel
	_SigKill                 // if signal.Notify doesn't take it, exit quietly
	_SigThrow                // if signal.Notify doesn't take it, exit loudly
	_SigPanic                // if the signal is from the kernel, panic
	_SigDefault              // if the signal isn't explicitly requested, don't monitor it
	_SigHandling             // our signal handler is registered
	_SigIgnored              // the signal was ignored before we registered for it
	_SigGoExit               // cause all runtime procs to exit (only used on Plan 9).
	_SigSetStack             // add SA_ONSTACK to libc handler
	_SigUnblock              // unblocked in minit
)

// Layout of in-memory per-function information prepared by linker
// See https://golang.org/s/go12symtab.
// Keep in sync with linker
// and with package debug/gosym and with symtab.go in package runtime.
type _func struct {
	entry   uintptr // start pc
	nameoff int32   // function name

	args  int32 // in/out args size
	frame int32 // legacy frame size; use pcsp if possible

	pcsp      int32
	pcfile    int32
	pcln      int32
	npcdata   int32
	nfuncdata int32
}

// layout of Itab known to compilers
// allocated in non-garbage-collected memory
type itab struct {
	inter  *interfacetype
	_type  *_type
	link   *itab
	bad    int32
	unused int32
	fun    [1]uintptr // variable sized
}

// Lock-free stack node.
// // Also known to export_test.go.
type lfnode struct {
	next    uint64
	pushcnt uintptr
}

type forcegcstate struct {
	lock mutex
	g    *g
	idle uint32
}

/*
 * known to compiler
 */
const (
	_Structrnd = regSize
)

// startup_random_data holds random bytes initialized at startup.  These come from
// the ELF AT_RANDOM auxiliary vector (vdso_linux_amd64.go or os_linux_386.go).
var startupRandomData []byte

// extendRandom extends the random numbers in r[:n] to the whole slice r.
// Treats n<0 as n==0.
func extendRandom(r []byte, n int) {
	if n < 0 {
		n = 0
	}
	for n < len(r) {
		// Extend random bits using hash function & time seed
		w := n
		if w > 16 {
			w = 16
		}
		h := memhash(unsafe.Pointer(&r[n-w]), uintptr(nanotime()), uintptr(w))
		for i := 0; i < ptrSize && n < len(r); i++ {
			r[n] = byte(h)
			n++
			h >>= 8
		}
	}
}

/*
 * deferred subroutine calls
 */
type _defer struct {
	siz     int32
	started bool
	sp      uintptr // sp at time of defer
	pc      uintptr
	fn      *funcval
	_panic  *_panic // panic that is running defer
	link    *_defer
}

/*
 * panics
 */
type _panic struct {
	argp      unsafe.Pointer // pointer to arguments of deferred call run during panic; cannot move - known to liblink
	arg       interface{}    // argument to panic
	link      *_panic        // link to earlier panic
	recovered bool           // whether this panic is over
	aborted   bool           // the panic was aborted
}

/*
 * stack traces
 */

type stkframe struct {
	fn       *_func     // function being run
	pc       uintptr    // program counter within fn
	continpc uintptr    // program counter where execution can continue, or 0 if not
	lr       uintptr    // program counter at caller aka link register
	sp       uintptr    // stack pointer at pc
	fp       uintptr    // stack pointer at caller aka frame pointer
	varp     uintptr    // top of local variables
	argp     uintptr    // pointer to function arguments
	arglen   uintptr    // number of bytes at argp
	argmap   *bitvector // force use of this argmap
}

const (
	_TraceRuntimeFrames = 1 << iota // include frames for internal runtime functions.
	_TraceTrap                      // the initial PC, SP are from a trap, not a return PC from a call
	_TraceJumpStack                 // if traceback is on a systemstack, resume trace at g that called into it
)

const (
	// The maximum number of frames we print for a traceback
	_TracebackMaxFrames = 100
)

var (
	emptystring string
	allg        **g
	allglen     uintptr
	lastg       *g
	allm        *m
	allp        [_MaxGomaxprocs + 1]*p
	gomaxprocs  int32
	panicking   uint32
	goos        *int8
	ncpu        int32
	signote     note
	forcegc     forcegcstate
	sched       schedt
	newprocs    int32

	// Information about what cpu features are available.
	// Set on startup in asm_{x86,amd64}.s.
	cpuid_ecx         uint32
	cpuid_edx         uint32
	lfenceBeforeRdtsc bool

	goarm uint8 // set by cmd/link on arm systems
)

// Set by the linker so the runtime can determine the buildmode.
var (
	islibrary bool // -buildmode=c-shared
	isarchive bool // -buildmode=c-archive
)

/*
 * mutual exclusion locks.  in the uncontended case,
 * as fast as spin locks (just a few user-level instructions),
 * but on the contention path they sleep in the kernel.
 * a zeroed Mutex is unlocked (no need to initialize each lock).
 */

/*
 * sleep and wakeup on one-time events.
 * before any calls to notesleep or notewakeup,
 * must call noteclear to initialize the Note.
 * then, exactly one thread can call notesleep
 * and exactly one thread can call notewakeup (once).
 * once notewakeup has been called, the notesleep
 * will return.  future notesleep will return immediately.
 * subsequent noteclear must be called only after
 * previous notesleep has returned, e.g. it's disallowed
 * to call noteclear straight after notewakeup.
 *
 * notetsleep is like notesleep but wakes up after
 * a given number of nanoseconds even if the event
 * has not yet happened.  if a goroutine uses notetsleep to
 * wake up early, it must wait to call noteclear until it
 * can be sure that no other goroutine is calling
 * notewakeup.
 *
 * notesleep/notetsleep are generally called on g0,
 * notetsleepg is similar to notetsleep but is called on user g.
 */
// bool	runtime·notetsleep(Note*, int64);  // false - timeout
// bool	runtime·notetsleepg(Note*, int64);  // false - timeout

/*
 * Lock-free stack.
 * Initialize uint64 head to 0, compare with 0 to test for emptiness.
 * The stack does not keep pointers to nodes,
 * so they can be garbage collected if there are no other pointers to nodes.
 */

// for mmap, we only pass the lower 32 bits of file offset to the
// assembly routine; the higher bits (if required), should be provided
// by the assembly routine as 0.
