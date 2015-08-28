// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package base

import (
	"unsafe"
)

/*
 * defined constants
 */
const (
	// G status
	//
	// If you add to this list, add to the list
	// of "okay during garbage collection" status
	// in mgcmark.go too.
	Gidle            = iota // 0
	Grunnable               // 1 runnable and on a run queue
	Grunning                // 2
	Gsyscall                // 3
	Gwaiting                // 4
	Gmoribund_unused        // 5 currently unused, but hardcoded in gdb scripts
	Gdead                   // 6
	Genqueue                // 7 Only the Gscanenqueue is used.
	Gcopystack              // 8 in this state when newstack is moving the stack
	// the following encode that the GC is scanning the stack and what to do when it is done
	Gscan = 0x1000 // atomicstatus&~Gscan = the non-scan state,
	// _Gscanidle =     _Gscan + _Gidle,      // Not used. Gidle only used with newly malloced gs
	Gscanrunnable = Gscan + Grunnable //  0x1001 When scanning completes make Grunnable (it is already on run queue)
	Gscanrunning  = Gscan + Grunning  //  0x1002 Used to tell preemption newstack routine to scan preempted stack.
	Gscansyscall  = Gscan + Gsyscall  //  0x1003 When scanning completes make it Gsyscall
	Gscanwaiting  = Gscan + Gwaiting  //  0x1004 When scanning completes make it Gwaiting
	// _Gscanmoribund_unused,               //  not possible
	// _Gscandead,                          //  not possible
	Gscanenqueue = Gscan + Genqueue //  When scanning completes make it Grunnable and put on runqueue
)

const (
	// P status
	Pidle    = iota
	Prunning // Only this P is allowed to change from _Prunning.
	Psyscall
	Pgcstop
	Pdead
)

// The next line makes 'go generate' write the zgen_*.go files with
// per-OS and per-arch information, including constants
// named goos_$GOOS and goarch_$GOARCH for every
// known GOOS and GOARCH. The constant is 1 on the
// current system, 0 otherwise; multiplying by them is
// useful for defining GOOS- or GOARCH-specific constants.
//go:generate go run gengoos.go

type Mutex struct {
	// Futex-based impl treats it as uint32 key,
	// while sema-based impl as M* waitm.
	// Used to be a union, but unions break precise GC.
	key uintptr
}

type Note struct {
	// Futex-based impl treats it as uint32 key,
	// while sema-based impl as M* waitm.
	// Used to be a union, but unions break precise GC.
	key uintptr
}

type Funcval struct {
	Fn uintptr
	// variable-size, fn-specific data here
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
type Guintptr uintptr

func (gp Guintptr) Ptr() *G   { return (*G)(unsafe.Pointer(gp)) }
func (gp *Guintptr) Set(g *G) { *gp = Guintptr(unsafe.Pointer(g)) }
func (gp *Guintptr) cas(old, new Guintptr) bool {
	return Casuintptr((*uintptr)(unsafe.Pointer(gp)), uintptr(old), uintptr(new))
}

type puintptr uintptr

func (pp puintptr) Ptr() *P   { return (*P)(unsafe.Pointer(pp)) }
func (pp *puintptr) Set(p *P) { *pp = puintptr(unsafe.Pointer(p)) }

type muintptr uintptr

func (mp muintptr) Ptr() *M   { return (*M)(unsafe.Pointer(mp)) }
func (mp *muintptr) Set(m *M) { *mp = muintptr(unsafe.Pointer(m)) }

type Gobuf struct {
	// The offsets of sp, pc, and g are known to (hard-coded in) libmach.
	Sp   uintptr
	Pc   uintptr
	G    Guintptr
	Ctxt unsafe.Pointer // this has to be a pointer so that gc scans it
	Ret  Uintreg
	Lr   uintptr
	bp   uintptr // for GOEXPERIMENT=framepointer
}

// Known to compiler.
// Changes here must also be made in src/cmd/internal/gc/select.go's selecttype.
type Sudog struct {
	G           *G
	Selectdone  *uint32
	Next        *Sudog
	Prev        *Sudog
	Elem        unsafe.Pointer // data element
	Releasetime int64
	Nrelease    int32  // -1 for acquire
	Waitlink    *Sudog // g.waiting list
}

type Gcstats struct {
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

// Stack describes a Go execution stack.
// The bounds of the stack are exactly [lo, hi),
// with no implicit data structures on either side.
type Stack struct {
	Lo uintptr
	Hi uintptr
}

// stkbar records the state of a G's stack barrier.
type Stkbar struct {
	SavedLRPtr uintptr // location overwritten by stack barrier PC
	SavedLRVal uintptr // value overwritten at savedLRPtr
}

type G struct {
	// Stack parameters.
	// stack describes the actual stack memory: [stack.lo, stack.hi).
	// stackguard0 is the stack pointer compared in the Go stack growth prologue.
	// It is stack.lo+StackGuard normally, but can be StackPreempt to trigger a preemption.
	// stackguard1 is the stack pointer compared in the C stack growth prologue.
	// It is stack.lo+StackGuard on g0 and gsignal stacks.
	// It is ~0 on other goroutine stacks, to trigger a call to morestackc (and crash).
	Stack       Stack   // offset known to runtime/cgo
	Stackguard0 uintptr // offset known to liblink
	stackguard1 uintptr // offset known to liblink

	Panic          *Panic  // innermost panic - offset known to liblink
	Defer          *Defer  // innermost defer
	M              *M      // current m; offset known to arm liblink
	StackAlloc     uintptr // stack allocation is [stack.lo,stack.lo+stackAlloc)
	Sched          Gobuf
	Syscallsp      uintptr        // if status==Gsyscall, syscallsp = sched.sp to use during gc
	Syscallpc      uintptr        // if status==Gsyscall, syscallpc = sched.pc to use during gc
	Stkbar         []Stkbar       // stack barriers, from low to high
	StkbarPos      uintptr        // index of lowest stack barrier not hit
	Param          unsafe.Pointer // passed parameter on wakeup
	Atomicstatus   uint32
	StackLock      uint32 // sigprof/scang lock; TODO: fold in to atomicstatus
	Goid           int64
	Waitsince      int64  // approx time when the g become blocked
	Waitreason     string // if status==Gwaiting
	Schedlink      Guintptr
	Preempt        bool   // preemption signal, duplicates stackguard0 = stackpreempt
	Paniconfault   bool   // panic (instead of crash) on unexpected fault address
	Preemptscan    bool   // preempted g does scan for gc
	Gcscandone     bool   // g has scanned stack; protected by _Gscan bit in status
	Gcscanvalid    bool   // false at start of gc cycle, true if G has not run since last scan
	Throwsplit     bool   // must not split stack
	raceignore     int8   // ignore race detection events
	Sysblocktraced bool   // StartTrace has emitted EvGoInSyscall about this goroutine
	sysexitticks   int64  // cputicks when syscall has returned (for tracing)
	sysexitseq     uint64 // trace seq when syscall has returned (for tracing)
	Lockedm        *M
	Sig            uint32
	Writebuf       []byte
	Sigcode0       uintptr
	Sigcode1       uintptr
	Sigpc          uintptr
	Gopc           uintptr // pc of go statement that created this goroutine
	Startpc        uintptr // pc of goroutine function
	Racectx        uintptr
	Waiting        *Sudog // sudog structures this g is waiting on (that have a valid elem ptr)
	readyg         *G     // scratch for readyExecute

	// Per-G gcController state
	Gcalloc    uintptr // bytes allocated during this GC cycle
	Gcscanwork int64   // scan work done (or stolen) this GC cycle
}

type mts struct {
	tv_sec  int64
	tv_nsec int64
}

type mscratch struct {
	v [6]uintptr
}

type M struct {
	G0      *G     // goroutine with scheduling stack
	Morebuf Gobuf  // gobuf arg to morestack
	divmod  uint32 // div/mod denominator for arm - known to liblink

	// Fields not known to debuggers.
	Procid        uint64     // for debuggers, but offset not hard-coded
	Gsignal       *G         // signal-handling g
	Sigmask       [4]uintptr // storage for saved signal mask
	tls           [4]uintptr // thread-local storage (for x86 extern register)
	mstartfn      func()
	Curg          *G       // current running goroutine
	caughtsig     Guintptr // goroutine running during fatal signal
	P             puintptr // attached p for executing go code (nil if not executing go code)
	Nextp         puintptr
	Id            int32
	Mallocing     int32
	Throwing      int32
	Preemptoff    string // if != "", keep curg running on this m
	Locks         int32
	Softfloat     int32
	dying         int32
	Profilehz     int32
	Helpgc        int32
	spinning      bool // m is out of work and is actively looking for work
	blocked       bool // m is blocked on a note
	inwb          bool // m is executing a write barrier
	printlock     int8
	fastrand      uint32
	Ncgocall      uint64 // number of cgo calls in total
	Ncgo          int32  // number of cgo calls currently in progress
	Park          Note
	Alllink       *M // on allm
	Schedlink     muintptr
	machport      uint32 // return address for mach ipc (os x)
	Mcache        *Mcache
	Lockedg       *G
	Createstack   [32]uintptr // stack that created this thread.
	freglo        [16]uint32  // d[i] lsb and f[i]
	freghi        [16]uint32  // d[i] msb and f[i+16]
	fflag         uint32      // floating point compare flags
	Locked        uint32      // tracking for lockosthread
	nextwaitm     uintptr     // next m waiting for lock
	Waitsema      uintptr     // semaphore for parking on locks
	waitsemacount uint32
	waitsemalock  uint32
	Gcstats       Gcstats
	Needextram    bool
	Traceback     uint8
	waitunlockf   unsafe.Pointer // todo go func(*g, unsafe.pointer) bool
	waitlock      unsafe.Pointer
	waittraceev   byte
	waittraceskip int
	Startingtrace bool
	Syscalltick   uint32
	//#ifdef GOOS_windows
	thread uintptr // thread handle
	// these are here because they are too large to be on the stack
	// of low-level NOSPLIT functions.
	libcall   libcall
	libcallpc uintptr // for cpu profiler
	Libcallsp uintptr
	libcallg  Guintptr
	Syscall   libcall // stores syscall parameters on windows
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

type P struct {
	lock Mutex

	Id          int32
	Status      uint32 // one of pidle/prunning/...
	Link        puintptr
	Schedtick   uint32   // incremented on every scheduler call
	Syscalltick uint32   // incremented on every system call
	M           muintptr // back-link to associated m (nil if idle)
	Mcache      *Mcache

	Deferpool    [5][]*Defer // pool of available defer structs of different sizes (see panic.go)
	Deferpoolbuf [5][32]*Defer

	// Cache of goroutine ids, amortizes accesses to runtime·sched.goidgen.
	Goidcache    uint64
	Goidcacheend uint64

	// Queue of runnable goroutines. Accessed without lock.
	Runqhead uint32
	Runqtail uint32
	Runq     [256]*G
	// runnext, if non-nil, is a runnable G that was ready'd by
	// the current G and should be run next instead of what's in
	// runq if there's time remaining in the running G's time
	// slice. It will inherit the time left in the current time
	// slice. If a set of goroutines is locked in a
	// communicate-and-wait pattern, this schedules that set as a
	// unit and eliminates the (potentially large) scheduling
	// latency that otherwise arises from adding the ready'd
	// goroutines to the end of the run queue.
	Runnext Guintptr

	// Available G's (status == Gdead)
	Gfree    *G
	Gfreecnt int32

	Sudogcache []*Sudog
	Sudogbuf   [128]*Sudog

	Tracebuf *TraceBuf

	palloc persistentAlloc // per-P to avoid mutex

	// Per-P GC state
	GcAssistTime     int64 // Nanoseconds in assistAlloc
	GcBgMarkWorker   *G
	GcMarkWorkerMode gcMarkWorkerMode

	// gcw is this P's GC work buffer cache. The work buffer is
	// filled by write barriers, drained by mutator assists, and
	// disposed on certain GC state transitions.
	Gcw GcWork

	RunSafePointFn uint32 // if 1, run sched.safePointFn at next safe point

	pad [64]byte
}

const (
	// The max value of GOMAXPROCS.
	// There are no fundamental restrictions on the value.
	MaxGomaxprocs = 1 << 8
)

type Schedt struct {
	Lock Mutex

	Goidgen uint64

	midle        muintptr // idle m's waiting for work
	Nmidle       int32    // number of idle m's waiting for work
	nmidlelocked int32    // number of locked m's waiting for work
	mcount       int32    // number of m's that have been created
	Maxmcount    int32    // maximum number of m's allowed (or die)

	Pidle      puintptr // idle p's
	Npidle     uint32
	Nmspinning uint32

	// Global runnable queue.
	Runqhead Guintptr
	Runqtail Guintptr
	Runqsize int32

	// Global cache of dead G's.
	Gflock Mutex
	Gfree  *G
	Ngfree int32

	// Central cache of sudog structs.
	Sudoglock  Mutex
	Sudogcache *Sudog

	// Central pool of available defer structs of different sizes.
	Deferlock Mutex
	Deferpool [5]*Defer

	Gcwaiting  uint32 // gc is waiting to run
	Stopwait   int32
	Stopnote   Note
	Sysmonwait uint32
	Sysmonnote Note
	Lastpoll   uint64

	// safepointFn should be called on each P at the next GC
	// safepoint if p.runSafePointFn is set.
	SafePointFn   func(*P)
	SafePointWait int32
	SafePointNote Note

	Profilehz int32 // cpu profiling rate

	Procresizetime int64 // nanotime() of last change to gomaxprocs
	Totaltime      int64 // ∫gomaxprocs dt up to procresizetime
}

// The m->locked word holds two pieces of state counting active calls to LockOSThread/lockOSThread.
// The low bit (LockExternal) is a boolean reporting whether any LockOSThread call is active.
// External locks are not recursive; a second lock is silently ignored.
// The upper bits of m->locked record the nesting depth of calls to lockOSThread
// (counting up by LockInternal), popped by unlockOSThread (counting down by LockInternal).
// Internal locks can be recursive. For instance, a lock for cgo can occur while the main
// goroutine is holding the lock during the initialization phase.
const (
	LockExternal = 1
	LockInternal = 2
)

const (
	SigNotify   = 1 << iota // let signal.Notify have signal, even if from kernel
	SigKill                 // if signal.Notify doesn't take it, exit quietly
	SigThrow                // if signal.Notify doesn't take it, exit loudly
	SigPanic                // if the signal is from the kernel, panic
	SigDefault              // if the signal isn't explicitly requested, don't monitor it
	SigHandling             // our signal handler is registered
	SigIgnored              // the signal was ignored before we registered for it
	SigGoExit               // cause all runtime procs to exit (only used on Plan 9).
	SigSetStack             // add SA_ONSTACK to libc handler
	SigUnblock              // unblocked in minit
)

// Layout of in-memory per-function information prepared by linker
// See https://golang.org/s/go12symtab.
// Keep in sync with linker
// and with package debug/gosym and with symtab.go in package runtime.
type Func struct {
	Entry   uintptr // start pc
	nameoff int32   // function name

	args  int32 // in/out args size
	frame int32 // legacy frame size; use pcsp if possible

	Pcsp      int32
	Pcfile    int32
	Pcln      int32
	Npcdata   int32
	Nfuncdata int32
}

// Lock-free stack node.
// // Also known to export_test.go.
type lfnode struct {
	next    uint64
	pushcnt uintptr
}

/*
 * deferred subroutine calls
 */
type Defer struct {
	Siz     int32
	Started bool
	Sp      uintptr // sp at time of defer
	Pc      uintptr
	Fn      *Funcval
	Panic   *Panic // panic that is running defer
	Link    *Defer
}

/*
 * panics
 */
type Panic struct {
	Argp      unsafe.Pointer // pointer to arguments of deferred call run during panic; cannot move - known to liblink
	Arg       interface{}    // argument to panic
	Link      *Panic         // link to earlier panic
	Recovered bool           // whether this panic is over
	Aborted   bool           // the panic was aborted
}

/*
 * stack traces
 */

type Stkframe struct {
	Fn       *Func      // function being run
	Pc       uintptr    // program counter within fn
	Continpc uintptr    // program counter where execution can continue, or 0 if not
	Lr       uintptr    // program counter at caller aka link register
	Sp       uintptr    // stack pointer at pc
	Fp       uintptr    // stack pointer at caller aka frame pointer
	Varp     uintptr    // top of local variables
	Argp     uintptr    // pointer to function arguments
	Arglen   uintptr    // number of bytes at argp
	Argmap   *Bitvector // force use of this argmap
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
	allg       **G
	Allglen    uintptr
	Allm       *M
	Allp       [MaxGomaxprocs + 1]*P
	Gomaxprocs int32
	Panicking  uint32
	Ncpu       int32
	Sched      Schedt
)

// Set by the linker so the runtime can determine the buildmode.
var (
	Islibrary bool // -buildmode=c-shared
	Isarchive bool // -buildmode=c-archive
)
