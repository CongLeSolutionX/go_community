// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lock

import (
	_core "runtime/internal/core"
)

/*
 * defined constants
 */
const (
	// G status
	//
	// If you add to this list, add to the list
	// of "okay during garbage collection" status
	// in mgc0.c too.
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
	Gscanrunnable = Gscan + Grunnable //  0x1001 When scanning complets make Grunnable (it is already on run queue)
	Gscanrunning  = Gscan + Grunning  //  0x1002 Used to tell preemption newstack routine to scan preempted stack.
	Gscansyscall  = Gscan + Gsyscall  //  0x1003 When scanning completes make is Gsyscall
	Gscanwaiting  = Gscan + Gwaiting  //  0x1004 When scanning completes make it Gwaiting
	// _Gscanmoribund_unused,               //  not possible
	// _Gscandead,                          //  not possible
	Gscanenqueue = Gscan + Genqueue //  When scanning completes make it Grunnable and put on runqueue
)

const (
	// P status
	Pidle = iota
	Prunning
	Psyscall
	Pgcstop
	Pdead
)

const (
	// The max value of GOMAXPROCS.
	// There are no fundamental restrictions on the value.
	MaxGomaxprocs = 1 << 8
)

// Layout of in-memory per-function information prepared by linker
// See http://golang.org/s/go12symtab.
// Keep in sync with linker and with ../../libmach/sym.c
// and with package debug/gosym and with symtab.go in package runtime.
type Func struct {
	Entry   uintptr // start pc
	nameoff int32   // function name

	args  int32 // in/out args size
	frame int32 // legacy frame size; use pcsp if possible

	pcsp      int32
	pcfile    int32
	pcln      int32
	Npcdata   int32
	Nfuncdata int32
}

// Holds variables parsed from GODEBUG env var.
type Debugvars struct {
	Allocfreetrace int32
	Efence         int32
	Gctrace        int32
	Gcdead         int32
	Scheddetail    int32
	Schedtrace     int32
	Scavenge       int32
}

/*
 * stack traces
 */

type Stkframe struct {
	Fn       *Func      // function being run
	Pc       uintptr    // program counter within fn
	Continpc uintptr    // program counter where execution can continue, or 0 if not
	lr       uintptr    // program counter at caller aka link register
	Sp       uintptr    // stack pointer at pc
	Fp       uintptr    // stack pointer at caller aka frame pointer
	Varp     uintptr    // top of local variables
	Argp     uintptr    // pointer to function arguments
	Arglen   uintptr    // number of bytes at argp
	Argmap   *Bitvector // force use of this argmap
}

const (
	TraceRuntimeFrames = 1 << 0 // include frames for internal runtime functions.
	TraceTrap          = 1 << 1 // the initial PC, SP are from a trap, not a return PC from a call
)

const (
	// The maximum number of frames we print for a traceback
	_TracebackMaxFrames = 100
)

var (
	Allm       *_core.M
	Allp       [MaxGomaxprocs + 1]*_core.P
	Gomaxprocs int32
	Panicking  uint32
	Ncpu       int32
	Debug      Debugvars
)
