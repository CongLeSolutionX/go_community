// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file implements tracing of lock operations for constructing
// the runtime lock order and checking for cycles. To use this
// facility, build with
//
//  -tags=locklog -gcflags=all=-d=maymorestack=runtime.lockLogMoreStack
//
// The "maymorestack" flag may be omitted if you don't care to track
// potential lock acquisitions caused by stack growths.

package runtime

import (
	"unsafe"
)

const LEAFRANK = 1000

var RANKCHECKS = false

//go:linkname lockLogInit os.runtime_lockLogInit
func lockLogInit() {
	RANKCHECKS = true
}

const (
	_Ldummy = iota
	_Lscavenge
	_Lforcegc
	_LsweepWaiters
	_LassistQueue
	_Lcpuprof
	_Lsweep

	_Lsched
	_Ldeadlock
	_Lpanic
	_Lallg
	_Lallp
	_LpollDesc

	_Ltimers
	_Litab
	_LreflectOffs
	_Lhchan
	_Lfin
	_LnotifyList
	_LtraceBuf
	_LtraceStrings
	_LmspanSpecial
	_Lprof
	_LgcBitsArenas
	_Lroot
	_Ltrace
	_LnetpollInit
	_Lstackpool
	_LstackLarge
	_Ldefer
	_LwbufSpans
	_Lsudog

	_Lmheap
	_LmheapSpecial
	_Lmcentral
	_LtraceTab
	_Lspine
	_LgFree

	_LrwmutexW
	_LrwmutexR

	// Leaf locks with no dependencies, so not actually used anywhere
	_LglobalAlloc
	_LnewmHandoff
	_LdebugPtrmask
	_LfaketimeState
	_Lticks
	_LraceFini
	_LpollCache
)

var lockNames []string = []string{
	"",

	// Locks held above sched
	"scavenge",
	"forcegc",
	"sweepWaiters",
	"assistQueue",
	"cpuprof",
	"sweep",

	"sched", // below scavenge, forcegc, sweepWaiters, assistQueue, cpuprof, sweep
	"deadlock",
	"panic", // below deadlock*
	"allg",  // below sched, panic
	"allp",  // below sched
	"pollDesc",

	"timers", // below scavenge, sched, allp, pollDesc, timers ; multiple locked simultaneously in destroy()
	"itab",
	"reflectOffs", // below itab

	"hchan",      // below hchan XXX Multiple hchans acquired in lock order in syncadjustsudogs, anywhere else?
	"fin",        // below sched, timers, hchan
	"notifyList", // below fin?
	"traceBuf",
	"traceStrings", // below traceBuf
	"mspanSpecial", // below scavenge, cpuprof, sched, allg, allp, timers, itab, reflectOffs, hchan, notifyList, traceBuf, traceStrings
	"prof",         // below scavenge, assistQueue, cpuprof, sched, allg, allp, timers, itab, reflectOffs, notifyList, traceBuf, traceStrings, hchan
	"gcBitsArenas", // below scavenge, cpuprof, sched, allg, timers, itab, reflectOffs, notifyList, traceBuf, traceStrings, hchan
	"root",
	"trace",       // below scavenge, assistQueue, sweep, sched, hchan, traceBuf, traceStrings, root
	"netpollInit", // below timers
	"stackpool",   // below scavenge, sweepWaiters, assistQueue, cpuprof, sched, pollDesc, timers, itab, reflectOffs, notifyList, traceBuf, traceStrings, hchan, prof, gcBitsArenas, root, trace, gcBitsArenas?, traceTab?, netPollInit
	"stackLarge",  // below sched, itab, hchan, prof
	"defer",
	"wbufSpans", // below sweepWaiters, sched, allg, timers, itab, notifyList, hchan, mspanSpecial, root, defer
	"sudog",     // below notifyList, hchan

	// Memory-related locks
	"mheap", // below scavenge, assistQueue, cpuprof, sched, allg, allp, timers, itab, reflectOffs, notifyList, traceBuf, traceStrings, hchan, mspanSpecial, prof, root, stackPool, stackLarge, defer, wbufSpans, sudog?

	// Memory-related leaf locks
	"mheapSpecial", // below scavenge, cpuprof, sched, allg, allp, timers, itab, reflectOffs, notifyList, traceBuf, traceStrings, hchan
	"mcentral",     // below scavenge, forcegc, assistQueue, cpuprof, sched, allg, allp, timers, itab, reflectOffs, notifyList, traceBuf, traceStrings, hchan

	"traceTab", // below scavenge, forcegc, sweepWaiters, assistQueue, sweep, sched, allg, timers, fin, notifyList, traceBuf, traceStrings, hchan, root, trace, mheap

	// Other leaf locks
	"spine", // below scavenge, cpuprof, sched, allg, timers, reflectOffs, notifyList, traceStrings, hchan
	"gFree", // below sched

	// Independent pair of locks
	"rwmutexW",
	"rwmutexR", // below rwmutexW

	// Leaf locks with no dependencies, so not actually used anywhere
	// There are other architecture-dependent leaf locks as well.
	"globalAlloc.mutex",
	"newmHandoff.lock",
	"debugPtrmask.lock",
	"faketimeState.lock",
	"ticks.lock",
	"raceFiniLock",
	"pollCache.lock",
}

var arcs [][20]int = [][20]int{
	{}, // _Ldummy
	{}, // _Lscavenge
	{}, // _Lforcegc
	{}, // _LsweepWaiters
	{}, // _LassistQueue
	{}, // _Lcpuprof
	{}, // _Lsweep
	{_Lscavenge, _Lforcegc, _LsweepWaiters, _LassistQueue, _Lcpuprof, _Lsweep}, // _Lsched
	{_Ldeadlock},       // _Ldeadlock
	{_Ldeadlock},       // _Lpanic
	{_Lsched, _Lpanic}, // _Lallg
	{_Lsched},          // _Lallp
	{},                 // _LpollDesc
	{_Lscavenge, _Lsched, _Lallp, _LpollDesc, _Ltimers}, // _Ltimers
	{},                           // _Litab
	{_Litab},                     // _LreflectOffs
	{_Lhchan},                    // _Lhchan
	{_Lsched, _Ltimers, _Lhchan}, // _Lfin
	{},                           // _LnotifyList
	{},                           // _LtraceBuf
	{_LtraceBuf},                 // _LtraceStrings
	{_Lscavenge, _Lcpuprof, _Lsched, _Lallg, _Lallp, _Ltimers, _Litab, _LreflectOffs, _Lhchan, _LnotifyList, _LtraceBuf, _LtraceStrings},                // _LmspanSpecial
	{_Lscavenge, _LassistQueue, _Lcpuprof, _Lsched, _Lallg, _Lallp, _Ltimers, _Litab, _LreflectOffs, _LnotifyList, _LtraceBuf, _LtraceStrings, _Lhchan}, // _Lprof
	{_Lscavenge, _Lcpuprof, _Lsched, _Lallg, _Ltimers, _Litab, _LreflectOffs, _LnotifyList, _LtraceBuf, _LtraceStrings, _Lhchan},                        // _LgcBitsArenas
	{}, // _Lroot
	{_Lscavenge, _LassistQueue, _Lsched, _Lhchan, _LtraceBuf, _LtraceStrings, _Lroot, _Lsweep}, // _Ltrace
	{_Ltimers}, // _LnetpollInit
	{_Lscavenge, _LsweepWaiters, _LassistQueue, _Lcpuprof, _Lsched, _LpollDesc, _Ltimers, _Litab, _LreflectOffs, _LnotifyList, _LtraceBuf, _LtraceStrings, _Lhchan, _Lprof, _LgcBitsArenas, _Lroot, _Ltrace, _LnetpollInit}, // _Lstackpool
	{_Lsched, _Litab, _Lhchan, _Lprof}, // _LstackLarge
	{},                                 // _Ldefer
	{_LsweepWaiters, _Lsched, _Lallg, _Ltimers, _Litab, _LnotifyList, _Lhchan, _LmspanSpecial, _Lroot, _Ldefer}, // _LwbufSpans
	{_LnotifyList, _Lhchan}, // _Lsudog
	{_Lscavenge, _LassistQueue, _Lcpuprof, _Lsched, _Lallg, _Lallp, _Ltimers, _Litab, _LreflectOffs, _LnotifyList, _LtraceBuf, _LtraceStrings, _Lhchan, _LmspanSpecial, _Lprof, _Lroot, _Lstackpool, _LstackLarge, _Ldefer, _LwbufSpans}, // _Lmheap
	{_Lscavenge, _Lcpuprof, _Lsched, _Lallg, _Lallp, _Ltimers, _Litab, _LreflectOffs, _LnotifyList, _LtraceBuf, _LtraceStrings, _Lhchan},                                                                                                 // _LmheapSpecial
	{_Lscavenge, _Lforcegc, _LassistQueue, _Lcpuprof, _Lsched, _Lallg, _Lallp, _Ltimers, _Litab, _LreflectOffs, _LnotifyList, _LtraceBuf, _LtraceStrings, _Lhchan},                                                                       // _Lmcentral
	{_Lscavenge, _Lforcegc, _LsweepWaiters, _LassistQueue, _Lsched, _Lallg, _Ltimers, _Lfin, _LnotifyList, _LtraceBuf, _LtraceStrings, _Lhchan, _Lroot, _Ltrace, _Lsweep, _Lmheap},                                                       // _LtraceTab
	{_Lscavenge, _Lcpuprof, _Lsched, _Lallg, _Ltimers, _LreflectOffs, _LnotifyList, _LtraceStrings, _Lhchan},                                                                                                                             // _Lspine
	{_Lsched}, // _LgFree

	{},           // _LrwmutexW
	{_LrwmutexW}, // _LrwmutexR

	// Leaf locks with no dependencies, so not actually used anywhere
	{}, // _LglobalAlloc
	{}, // _LnewmHandoff
	{}, // _LdebugPtrmask
	{}, // _LfaketimeState
	{}, // _Lticks
	{}, // _LraceFini
	{}, // _LpollCache
}

func lockInit(l *mutex, rank int) {
	l.rank = rank
}

// lockLabeled is like lock(l), but records the lock class and rank
// for a non-static lock acquisition.
func lockLabeled(l *mutex, rank int) {
	if !RANKCHECKS || l == &debuglock {
		lock2(l)
		return
	}
	if rank == 0 {
		rank = LEAFRANK
	}
	gp := getg()
	// Log the new class.
	systemstack(func() {
		i := gp.m.lockIndex
		if i >= 10 {
			throw("overflow")
		}
		gp.m.locksHeld[i].rank = rank
		gp.m.locksHeld[i].l = uintptr(unsafe.Pointer(l))
		gp.m.lockIndex++
		i++
		if i > 1 && rank != LEAFRANK /* && !(gp.m.locksHeld[i-2].rank == _LtraceTab && gp.m.locksHeld[i-1].rank == _Lstackpool) */ {
			found := false
			list := arcs[gp.m.locksHeld[i-1].rank]
			for j := 0; j < 20; j++ {
				if list[j] == 0 {
					break
				}
				if list[j] == gp.m.locksHeld[i-2].rank {
					found = true
					break
				}
			}
			if !found {
				println(gp.m.procid, " ======")
				for j := 0; j < i; j++ {
					println(j, ":", lockNames[gp.m.locksHeld[j].rank], " ", gp.m.locksHeld[j].rank, " ", unsafe.Pointer(gp.m.locksHeld[j].l))
				}
				throw("hi")
			}
		}
		// Debug code for checking explicitly for some lock dependencies, etc.
		if false && i > 1 && gp.m.locksHeld[i-1].rank == _LnotifyList {
			println(gp.m.procid, " ======")
			for j := 0; j < i; j++ {
				println(j, ":", lockNames[gp.m.locksHeld[j].rank], " ", gp.m.locksHeld[j].rank, " ", unsafe.Pointer(gp.m.locksHeld[j].l))
			}
			//throw("Lock dependency")
		}
		if i > 1 && gp.m.locksHeld[i-1].rank < gp.m.locksHeld[i-2].rank /* && !(gp.m.locksHeld[i-2].rank == _LtraceTab && gp.m.locksHeld[i-1].rank == _Lstackpool) */ {
			println(gp.m.procid, " ======")
			for j := 0; j < i; j++ {
				println(j, ":", lockNames[gp.m.locksHeld[j].rank], " ", gp.m.locksHeld[j].rank, " ", unsafe.Pointer(gp.m.locksHeld[j].l))
			}
			throw("lock ordering problem")
		}
		//println("acquireLabeled", name, rank, uint64(uintptr(unsafe.Pointer(l))))
		// Actually lock the lock.
		lock2(l)
	})
}

// The following functions are the entry-points to record lock
// operations.
//
// All of these are nosplit and switch to the system stack immediately
// to avoid stack growths. Since a stack growth could itself have lock
// operations, this prevents re-entrant calls.

//go:nosplit
func lockLogAcquire(l *mutex) {
	if !RANKCHECKS || l == &debuglock {
		return
	}
	//pc := getcallerpc()
	//sp := getcallersp()
	//gp := getg()
	systemstack(func() {
		println("acquire", uint64(uintptr(unsafe.Pointer(l))))
		//n := gentraceback(pc, sp, 0, gp, skip, &pcs[0], len(pcs), nil, nil, _TraceJumpStack)
	})
}

//go:nosplit
func lockLogRelease(l *mutex) {
	if !RANKCHECKS || l == &debuglock {
		return
	}
	gp := getg()
	systemstack(func() {
		if gp.m.lockIndex < 1 {
			// XXX Why does this happen?  Does systemstack() sometimes change m?
			//println(gp.m.procid, "no locks held", l)
			return
		}
		found := false
		for i := gp.m.lockIndex - 1; i >= 0; i-- {
			if gp.m.locksHeld[i].l == uintptr(unsafe.Pointer(l)) {
				found = true
				for j := i; j < gp.m.lockIndex-1; j++ {
					gp.m.locksHeld[j] = gp.m.locksHeld[j+1]
				}
				gp.m.lockIndex--
			}
		}
		if !found {
			println(gp.m.procid, "unmatching lock", l)
		}
		//println("release", uint64(uintptr(unsafe.Pointer(l))))
	})
}

//go:nosplit
func lockLogMayAcquire(l *mutex, name int) {
	if !RANKCHECKS {
		return
	}
	//pc := getcallerpc()
	//sp := getcallersp()
	//gp := getg()
	gp := getg()

	systemstack(func() {
		i := gp.m.lockIndex
		if i > 0 && name < gp.m.locksHeld[i-1].rank {
			println(gp.m.procid, " ======")
			for j := 0; j < i; j++ {
				println(j, ":", lockNames[gp.m.locksHeld[j].rank], " ", gp.m.locksHeld[j].rank, " ", unsafe.Pointer(gp.m.locksHeld[j].l))
			}
			println(i, ":", lockNames[name], " ", unsafe.Pointer(l))
			throw("lock ordering problem, maybe")
		}
	})
}

// lockLogMoreStack records that a conditional morestack call may
// acquire the heap lock. This should be used with the compiler's
// -d=maymorestack=runtime.lockLogMoreStack flag.
//
//go:linkname lockLogMoreStack
//go:nosplit
func lockLogMoreStack() {
	// Only log if we're on a user stack (so it could possibly
	// grow) and if we already hold a lock (otherwise this can't
	// contribute to the lock graph anyway).
	gp := getg()
	if gp == nil || gp.m == nil || gp.m.locks == 0 || gp != gp.m.curg {
		return
	}

	// It's safe to call lockLogMayAcquire directly because it
	// doesn't grow the stack.
	lockLogMayAcquire(&mheap_.lock, _Lmheap)
}
