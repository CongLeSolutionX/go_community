// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file implements static lock ranking of the locks in the runtime. If a lock
// is not given a rank, then it is assumed to be a leaf lock, which means no other
// lock can be acquired while it is held. Therefore, leaf locks do not need to be
// given an explicit rank. We list all of the architecture-independent leaf locks
// for documentation purposes, but don't list any of the architecture-dependent
// locks (which are all leaf locks).  debugLock is ignored for ranking, since it is used
// when printing out lock ranking errors.
//
// lockInit(l *mutex, rank int) is used to set the rank of lock before it is used.
// If there is no clear place to initialize a lock, then the rank of a lock can be
// specified during the lock call itself via lockRankAcquire(l *mutex, rank int).
//
// Besides the static lock ranking (which is a total ordering of the locks), we
// also represent and enforce the actual partial order among the locks in the
// arc[] array below. That is, if it is possible that lock B can be acquired when
// lock A is the previous acquired lock that is still held, then there should be
// an entry for A in arcs[B][]. We will currently fail not only if the total order
// (the lock ranking) is violated, but also if there is a missing entry in the
// partial order.

package runtime

const LEAFRANK = 1000

// Constants representing the lock rank of the architecture-independent locks in
// the runtime.
const (
	_Ldummy = iota

	// Locks held above sched
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

	_Ltimers // Multiple timers locked simultaneously in destroy()
	_Litab
	_LreflectOffs
	_Lhchan // Multiple hchans acquired in lock order in syncadjustsudogs()
	_Lfin
	_LnotifyList
	_LtraceBuf
	_LtraceStrings
	_LmspanSpecial
	_Lprof
	_LgcBitsArenas
	_Lroot
	_Ltrace
	_LtraceTab
	_LnetpollInit
	_Lstackpool
	_LstackLarge
	_Ldefer
	_Lsudog
	_LwbufSpans

	// Memory-related leaf locks
	_Lmheap
	_LmheapSpecial
	_Lmcentral

	// Other leaf locks
	_Lspine
	_LgFree

	// Independent pair of locks
	_LrwmutexW
	_LrwmutexR

	// Leaf locks with no dependencies, so these constants are not actually used anywhere.
	// There are other architecture-dependent leaf locks as well.
	_LglobalAlloc
	_LnewmHandoff
	_LdebugPtrmask
	_LfaketimeState
	_Lticks
	_LraceFini
	_LpollCache
	_Ldebug
)

// The names associated with each of the above ranks
var lockNames []string = []string{
	"",

	"scavenge",
	"forcegc",
	"sweepWaiters",
	"assistQueue",
	"cpuprof",
	"sweep",

	"sched",
	"deadlock",
	"panic",
	"allg",
	"allp",
	"pollDesc",

	"timers",
	"itab",
	"reflectOffs",

	"hchan",
	"fin",
	"notifyList",
	"traceBuf",
	"traceStrings",
	"mspanSpecial",
	"prof",
	"gcBitsArenas",
	"root",
	"trace",
	"traceTab",

	"netpollInit",
	"stackpool",
	"stackLarge",
	"defer",
	"sudog",
	"wbufSpans",

	"mheap",
	"mheapSpecial",
	"mcentral",

	"spine",
	"gFree",

	"rwmutexW",
	"rwmutexR",

	"globalAlloc.mutex",
	"newmHandoff.lock",
	"debugPtrmask.lock",
	"faketimeState.lock",
	"ticks.lock",
	"raceFiniLock",
	"pollCache.lock",
	"debugLock",
}

// Increase if needed if for a particular lock, there are more than 25 locks that
// can be held immediately "before" it in the lock list.
const maxArcsPerLock = 25

// A partial order among the various lock types, listing the immediate ordering
// that has actually been observed in the runtime. Each entry (which corresponds
// to a particular lock rank) specifies the list of locks that can be already be held
// immediately "above" it.
//
// So, for example, the _Lsched entry shows that all the locks preceding it in
// rank can actually be held. The fin lock shows that only the sched, timers, or
// hchan lock can be held immediately above it when it is acquired.
var arcs [][maxArcsPerLock]int = [][maxArcsPerLock]int{
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
	{},                             // _Litab
	{_Litab},                       // _LreflectOffs
	{_Lscavenge, _Lsweep, _Lhchan}, // _Lhchan
	{_Lsched, _Ltimers, _Lhchan},   // _Lfin
	{},                             // _LnotifyList
	{},                             // _LtraceBuf
	{_LtraceBuf},                   // _LtraceStrings
	{_Lscavenge, _Lcpuprof, _Lsched, _Lallg, _Lallp, _Ltimers, _Litab, _LreflectOffs, _Lhchan, _LnotifyList, _LtraceBuf, _LtraceStrings},                         // _LmspanSpecial
	{_Lscavenge, _LassistQueue, _Lcpuprof, _Lsweep, _Lsched, _Lallg, _Lallp, _Ltimers, _Litab, _LreflectOffs, _LnotifyList, _LtraceBuf, _LtraceStrings, _Lhchan}, // _Lprof
	{_Lscavenge, _LassistQueue, _Lcpuprof, _Lsched, _Lallg, _Ltimers, _Litab, _LreflectOffs, _LnotifyList, _LtraceBuf, _LtraceStrings, _Lhchan},                  // _LgcBitsArenas
	{}, // _Lroot
	{_Lscavenge, _LassistQueue, _Lsched, _Lhchan, _LtraceBuf, _LtraceStrings, _Lroot, _Lsweep},                                                                            // _Ltrace
	{_Lscavenge, _Lforcegc, _LsweepWaiters, _LassistQueue, _Lsched, _Lallg, _Ltimers, _Lfin, _LnotifyList, _LtraceBuf, _LtraceStrings, _Lhchan, _Lroot, _Ltrace, _Lsweep}, // _LtraceTab
	{_Ltimers}, // _LnetpollInit
	{_Lscavenge, _LsweepWaiters, _LassistQueue, _Lcpuprof, _Lsweep, _Lsched, _LpollDesc, _Ltimers, _Litab, _LreflectOffs, _LnotifyList, _LtraceBuf, _LtraceStrings, _Lhchan, _Lprof, _LgcBitsArenas, _Lroot, _Ltrace, _LnetpollInit}, // _Lstackpool
	{_Lsched, _Litab, _Lhchan, _Lprof, _Lroot}, // _LstackLarge
	{},                      // _Ldefer
	{_LnotifyList, _Lhchan}, // _Lsudog
	{_Lscavenge, _LsweepWaiters, _Lsched, _Lallg, _LpollDesc, _Ltimers, _Litab, _LreflectOffs, _LnotifyList, _Lhchan, _LmspanSpecial, _Lroot, _Ldefer, _Lsudog},                                                                                                            // _LwbufSpans
	{_Lscavenge, _LsweepWaiters, _LassistQueue, _Lcpuprof, _Lsweep, _Lsched, _Lallg, _Lallp, _Ltimers, _Litab, _LreflectOffs, _LnotifyList, _LtraceBuf, _LtraceStrings, _Lhchan, _LmspanSpecial, _Lprof, _Lroot, _Lstackpool, _LstackLarge, _Ldefer, _Lsudog, _LwbufSpans}, // _Lmheap
	{_Lscavenge, _Lcpuprof, _Lsched, _Lallg, _Lallp, _Ltimers, _Litab, _LreflectOffs, _LnotifyList, _LtraceBuf, _LtraceStrings, _Lhchan},                                                                                                                                   // _LmheapSpecial
	{_Lscavenge, _Lforcegc, _LassistQueue, _Lcpuprof, _Lsched, _Lallg, _Lallp, _Ltimers, _Litab, _LreflectOffs, _LnotifyList, _LtraceBuf, _LtraceStrings, _Lhchan},                                                                                                         // _Lmcentral
	{_Lscavenge, _Lcpuprof, _Lsched, _Lallg, _Ltimers, _LreflectOffs, _LnotifyList, _LtraceStrings, _Lhchan},                                                                                                                                                               // _Lspine
	{_Lsched}, // _LgFree

	{},           // _LrwmutexW
	{_LrwmutexW}, // _LrwmutexR

	{}, // _LglobalAlloc
	{}, // _LnewmHandoff
	{}, // _LdebugPtrmask
	{}, // _LfaketimeState
	{}, // _Lticks
	{}, // _LraceFini
	{}, // _LpollCache
	{}, // _LdebugLock
}
