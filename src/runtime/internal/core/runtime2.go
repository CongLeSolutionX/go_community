// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	"unsafe"
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
	Key uintptr
}

type Note struct {
	// Futex-based impl treats it as uint32 key,
	// while sema-based impl as M* waitm.
	// Used to be a union, but unions break precise GC.
	Key uintptr
}

type Funcval struct {
	Fn uintptr
	// variable-size, fn-specific data here
}

type Eface struct {
	Type *Type
	Data unsafe.Pointer
}

type Slice struct {
	Array *byte // actual data
	Len   uint  // number of elements
	Cap   uint  // allocated number of elements
}

// A guintptr holds a goroutine pointer, but typed as a uintptr
// to bypass write barriers. It is used in the Gobuf goroutine state.
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

func (gp Guintptr) Ptr() *G {
	return (*G)(unsafe.Pointer(gp))
}

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
// Changes here must also be made in src/cmd/gc/select.c's selecttype.
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
	Nhandoff    uint64
	Nhandoffcnt uint64
	Nprocyield  uint64
	Nosyield    uint64
	Nsleep      uint64
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
	Stackguard1 uintptr // offset known to liblink

	Panic        *Panic // innermost panic - offset known to liblink
	Defer        *Defer // innermost defer
	Sched        Gobuf
	Syscallsp    uintptr        // if status==gsyscall, syscallsp = sched.sp to use during gc
	Syscallpc    uintptr        // if status==gsyscall, syscallpc = sched.pc to use during gc
	Param        unsafe.Pointer // passed parameter on wakeup
	Atomicstatus uint32
	Goid         int64
	Waitsince    int64  // approx time when the g become blocked
	Waitreason   string // if status==gwaiting
	Schedlink    *G
	Issystem     bool // do not output in stack dump, ignore in deadlock detector
	Preempt      bool // preemption signal, duplicates stackguard0 = stackpreempt
	Paniconfault bool // panic (instead of crash) on unexpected fault address
	Preemptscan  bool // preempted g does scan for gc
	Gcworkdone   bool // debug: cleared at begining of gc work phase cycle, set by gcphasework, tested at end of cycle
	Gcscanvalid  bool // false at start of gc cycle, true if G has not run since last scan
	Throwsplit   bool // must not split stack
	raceignore   int8 // ignore race detection events
	M            *M   // for debuggers, but offset not hard-coded
	Lockedm      *M
	Sig          uint32
	Writebuf     []byte
	Sigcode0     uintptr
	Sigcode1     uintptr
	Sigpc        uintptr
	Gopc         uintptr // pc of go statement that created this goroutine
	Startpc      uintptr // pc of goroutine function
	Racectx      uintptr
	Waiting      *Sudog // sudog structures this g is waiting on (that have a valid elem ptr)
}

type mts struct {
	tv_sec  int64
	tv_nsec int64
}

type mscratch struct {
	v [6]uintptr
}

type M struct {
	G0      *G    // goroutine with scheduling stack
	Morebuf Gobuf // gobuf arg to morestack

	// Fields not known to debuggers.
	Procid        uint64         // for debuggers, but offset not hard-coded
	Gsignal       *G             // signal-handling g
	Tls           [4]uintptr     // thread-local storage (for x86 extern register)
	Mstartfn      unsafe.Pointer // todo go func()
	Curg          *G             // current running goroutine
	Caughtsig     *G             // goroutine running during fatal signal
	P             *P             // attached p for executing go code (nil if not executing go code)
	Nextp         *P
	Id            int32
	Mallocing     int32
	Throwing      int32
	Preemptoff    string // if != "", keep curg running on this m
	Locks         int32
	Softfloat     int32
	Dying         int32
	Profilehz     int32
	Helpgc        int32
	Spinning      bool // m is out of work and is actively looking for work
	Blocked       bool // m is blocked on a note
	Inwb          bool // m is executing a write barrier
	Printlock     int8
	Fastrand      uint32
	Ncgocall      uint64 // number of cgo calls in total
	Ncgo          int32  // number of cgo calls currently in progress
	Cgomal        *Cgomal
	Park          Note
	Alllink       *M // on allm
	Schedlink     *M
	Machport      uint32 // return address for mach ipc (os x)
	Mcache        *Mcache
	Lockedg       *G
	Createstack   [32]uintptr // stack that created this thread.
	freglo        [16]uint32  // d[i] lsb and f[i]
	freghi        [16]uint32  // d[i] msb and f[i+16]
	fflag         uint32      // floating point compare flags
	Locked        uint32      // tracking for lockosthread
	Nextwaitm     *M          // next m waiting for lock
	Waitsema      uintptr     // semaphore for parking on locks
	waitsemacount uint32
	waitsemalock  uint32
	Gcstats       Gcstats
	Needextram    bool
	Traceback     uint8
	Waitunlockf   unsafe.Pointer // todo go func(*g, unsafe.pointer) bool
	Waitlock      unsafe.Pointer
	Waittraceev   byte
	Syscalltick   uint32
	//#ifdef GOOS_windows
	thread uintptr // thread handle
	// these are here because they are too large to be on the stack
	// of low-level NOSPLIT functions.
	libcall   libcall
	Libcallpc uintptr // for cpu profiler
	Libcallsp uintptr
	Libcallg  *G
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
	Link        *P
	Schedtick   uint32 // incremented on every scheduler call
	Syscalltick uint32 // incremented on every system call
	M           *M     // back-link to associated m (nil if idle)
	Mcache      *Mcache
	Deferpool   [5]*Defer // pool of available defer structs of different sizes (see panic.c)

	// Cache of goroutine ids, amortizes accesses to runtime·sched.goidgen.
	Goidcache    uint64
	Goidcacheend uint64

	// Queue of runnable goroutines.
	Runqhead uint32
	Runqtail uint32
	Runq     [256]*G

	// Available G's (status == Gdead)
	Gfree    *G
	Gfreecnt int32

	Tracebuf *TraceBuf

	pad [64]byte
}

type Schedt struct {
	Lock Mutex

	Goidgen uint64

	Midle        *M    // idle m's waiting for work
	Nmidle       int32 // number of idle m's waiting for work
	Nmidlelocked int32 // number of locked m's waiting for work
	Mcount       int32 // number of m's that have been created
	Maxmcount    int32 // maximum number of m's allowed (or die)

	Pidle      *P // idle p's
	Npidle     uint32
	Nmspinning uint32

	// Global runnable queue.
	Runqhead *G
	Runqtail *G
	Runqsize int32

	// Global cache of dead G's.
	Gflock Mutex
	Gfree  *G
	Ngfree int32

	Gcwaiting  uint32 // gc is waiting to run
	Stopwait   int32
	Stopnote   Note
	Sysmonwait uint32
	Sysmonnote Note
	Lastpoll   uint64

	Profilehz int32 // cpu profiling rate
}

// layout of Itab known to compilers
// allocated in non-garbage-collected memory
type Itab struct {
	Inter  *Interfacetype
	Type   *Type
	Link   *Itab
	Bad    int32
	unused int32
	Fun    [1]uintptr // variable sized
}

// Track memory allocated by code not written in Go during a cgo call,
// so that the garbage collector can see them.
type Cgomal struct {
	Next  *Cgomal
	Alloc unsafe.Pointer
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

var (
	lastg      *G
	Needextram uint32
	signote    Note
	Sched      Schedt
)
