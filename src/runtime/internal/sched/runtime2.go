// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sched

import (
	"unsafe"
)

const (
	SigNotify   = 1 << 0 // let signal.Notify have signal, even if from kernel
	SigKill     = 1 << 1 // if signal.Notify doesn't take it, exit quietly
	SigThrow    = 1 << 2 // if signal.Notify doesn't take it, exit loudly
	SigPanic    = 1 << 3 // if the signal is from the kernel, panic
	SigDefault  = 1 << 4 // if the signal isn't explicitly requested, don't monitor it
	SigHandling = 1 << 5 // our signal handler is registered
	SigIgnored  = 1 << 6 // the signal was ignored before we registered for it
	SigGoExit   = 1 << 7 // cause all runtime procs to exit (only used on Plan 9).
	SigSetStack = 1 << 8 // add SA_ONSTACK to libc handler
)

// Lock-free stack node.
// // Also known to export_test.go.
type lfnode struct {
	next    uint64
	pushcnt uintptr
}

// Parallel for descriptor.
type Parfor struct {
	Body    unsafe.Pointer // go func(*parfor, uint32), executed for each element
	Done    uint32         // number of idle threads
	Nthr    uint32         // total number of threads
	Nthrmax uint32         // maximum number of threads
	Thrseq  uint32         // thread id sequencer
	Cnt     uint32         // iteration space [0, cnt)
	Ctx     unsafe.Pointer // arbitrary user context
	Wait    bool           // if true, wait while all threads finish processing,
	// otherwise parfor may return while other threads are still working
	Thr *Parforthread // array of thread descriptors
	pad uint32        // to align parforthread.pos for 64-bit atomic operations
	// stats
	Nsteal     uint64
	Nstealcnt  uint64
	Nprocyield uint64
	Nosyield   uint64
	Nsleep     uint64
}

// Indicates to write barrier and sychronization task to preform.
const (
	GCoff             = iota // GC not running, write barrier disabled
	GCquiesce                // unused state
	GCstw                    // unused state
	GCscan                   // GC collecting roots into workbufs, write barrier disabled
	GCmark                   // GC marking from workbufs, write barrier ENABLED
	GCmarktermination        // GC mark termination: allocate black, P's help GC, write barrier ENABLED
	GCsweep                  // GC mark completed; sweeping in background, write barrier disabled
)

var Gcphase uint32

var (
	Iscgo bool
)
