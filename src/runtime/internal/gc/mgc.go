// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// TODO(rsc): The code having to do with the heap bitmap needs very serious cleanup.
// It has gotten completely out of control.

// Garbage collector (GC).
//
// The GC runs concurrently with mutator threads, is type accurate (aka precise), allows multiple
// GC thread to run in parallel. It is a concurrent mark and sweep that uses a write barrier. It is
// non-generational and non-compacting. Allocation is done using size segregated per P allocation
// areas to minimize fragmentation while eliminating locks in the common case.
//
// The algorithm decomposes into several steps.
// This is a high level description of the algorithm being used. For an overview of GC a good
// place to start is Richard Jones' gchandbook.org.
//
// The algorithm's intellectual heritage includes Dijkstra's on-the-fly algorithm, see
// Edsger W. Dijkstra, Leslie Lamport, A. J. Martin, C. S. Scholten, and E. F. M. Steffens. 1978.
// On-the-fly garbage collection: an exercise in cooperation. Commun. ACM 21, 11 (November 1978),
// 966-975.
// For journal quality proofs that these steps are complete, correct, and terminate see
// Hudson, R., and Moss, J.E.B. Copying Garbage Collection without stopping the world.
// Concurrency and Computation: Practice and Experience 15(3-5), 2003.
//
//  0. Set phase = GCscan from GCoff.
//  1. Wait for all P's to acknowledge phase change.
//         At this point all goroutines have passed through a GC safepoint and
//         know we are in the GCscan phase.
//  2. GC scans all goroutine stacks, mark and enqueues all encountered pointers
//       (marking avoids most duplicate enqueuing but races may produce benign duplication).
//       Preempted goroutines are scanned before P schedules next goroutine.
//  3. Set phase = GCmark.
//  4. Wait for all P's to acknowledge phase change.
//  5. Now write barrier marks and enqueues black, grey, or white to white pointers.
//       Malloc still allocates white (non-marked) objects.
//  6. Meanwhile GC transitively walks the heap marking reachable objects.
//  7. When GC finishes marking heap, it preempts P's one-by-one and
//       retakes partial wbufs (filled by write barrier or during a stack scan of the goroutine
//       currently scheduled on the P).
//  8. Once the GC has exhausted all available marking work it sets phase = marktermination.
//  9. Wait for all P's to acknowledge phase change.
// 10. Malloc now allocates black objects, so number of unmarked reachable objects
//        monotonically decreases.
// 11. GC preempts P's one-by-one taking partial wbufs and marks all unmarked yet
//        reachable objects.
// 12. When GC completes a full cycle over P's and discovers no new grey
//         objects, (which means all reachable objects are marked) set phase = GCoff.
// 13. Wait for all P's to acknowledge phase change.
// 14. Now malloc allocates white (but sweeps spans before use).
//         Write barrier becomes nop.
// 15. GC does background sweeping, see description below.
// 16. When sufficient allocation has taken place replay the sequence starting at 0 above,
//         see discussion of GC rate below.

// Changing phases.
// Phases are changed by setting the gcphase to the next phase and possibly calling ackgcphase.
// All phase action must be benign in the presence of a change.
// Starting with GCoff
// GCoff to GCscan
//     GSscan scans stacks and globals greying them and never marks an object black.
//     Once all the P's are aware of the new phase they will scan gs on preemption.
//     This means that the scanning of preempted gs can't start until all the Ps
//     have acknowledged.
//     When a stack is scanned, this phase also installs stack barriers to
//     track how much of the stack has been active.
//     This transition enables write barriers because stack barriers
//     assume that writes to higher frames will be tracked by write
//     barriers. Technically this only needs write barriers for writes
//     to stack slots, but we enable write barriers in general.
// GCscan to GCmark
//     In GCmark, work buffers are drained until there are no more
//     pointers to scan.
//     No scanning of objects (making them black) can happen until all
//     Ps have enabled the write barrier, but that already happened in
//     the transition to GCscan.
// GCmark to GCmarktermination
//     The only change here is that we start allocating black so the Ps must acknowledge
//     the change before we begin the termination algorithm
// GCmarktermination to GSsweep
//     Object currently on the freelist must be marked black for this to work.
//     Are things on the free lists black or white? How does the sweep phase work?

// Concurrent sweep.
//
// The sweep phase proceeds concurrently with normal program execution.
// The heap is swept span-by-span both lazily (when a goroutine needs another span)
// and concurrently in a background goroutine (this helps programs that are not CPU bound).
// At the end of STW mark termination all spans are marked as "needs sweeping".
//
// The background sweeper goroutine simply sweeps spans one-by-one.
//
// To avoid requesting more OS memory while there are unswept spans, when a
// goroutine needs another span, it first attempts to reclaim that much memory
// by sweeping. When a goroutine needs to allocate a new small-object span, it
// sweeps small-object spans for the same object size until it frees at least
// one object. When a goroutine needs to allocate large-object span from heap,
// it sweeps spans until it frees at least that many pages into heap. There is
// one case where this may not suffice: if a goroutine sweeps and frees two
// nonadjacent one-page spans to the heap, it will allocate a new two-page
// span, but there can still be other one-page unswept spans which could be
// combined into a two-page span.
//
// It's critical to ensure that no operations proceed on unswept spans (that would corrupt
// mark bits in GC bitmap). During GC all mcaches are flushed into the central cache,
// so they are empty. When a goroutine grabs a new span into mcache, it sweeps it.
// When a goroutine explicitly frees an object or sets a finalizer, it ensures that
// the span is swept (either by sweeping it, or by waiting for the concurrent sweep to finish).
// The finalizer goroutine is kicked off only when all spans are swept.
// When the next GC starts, it sweeps all not-yet-swept spans (if any).

// GC rate.
// Next GC is after we've allocated an extra amount of memory proportional to
// the amount already in use. The proportion is controlled by GOGC environment variable
// (100 by default). If GOGC=100 and we're using 4M, we'll GC again when we get to 8M
// (this mark is tracked in next_gc variable). This keeps the GC cost in linear
// proportion to the allocation cost. Adjusting GOGC just changes the linear constant
// (and also the amount of extra memory used).

package gc

import (
	_base "runtime/internal/base"
	"unsafe"
)

func Gcinit() {
	if unsafe.Sizeof(_base.Workbuf{}) != _base.WorkbufSize {
		_base.Throw("size of Workbuf is suboptimal")
	}

	_base.Work.Markfor = parforalloc(_base.MaxGcproc)
	_ = setGCPercent(readgogc())
	for datap := &_base.Firstmoduledata; datap != nil; datap = datap.Next {
		datap.Gcdatamask = progToPointerMask((*byte)(unsafe.Pointer(datap.Gcdata)), datap.Edata-datap.Data)
		datap.Gcbssmask = progToPointerMask((*byte)(unsafe.Pointer(datap.Gcbss)), datap.Ebss-datap.Bss)
	}
	_base.Memstats.Next_gc = _base.Heapminimum
}

func readgogc() int32 {
	p := Gogetenv("GOGC")
	if p == "" {
		return 100
	}
	if p == "off" {
		return -1
	}
	return int32(Atoi(p))
}

func setGCPercent(in int32) (out int32) {
	_base.Lock(&_base.Mheap_.Lock)
	out = _base.Gcpercent
	if in < 0 {
		in = -1
	}
	_base.Gcpercent = in
	_base.Heapminimum = _base.DefaultHeapMinimum * uint64(_base.Gcpercent) / 100
	_base.Unlock(&_base.Mheap_.Lock)
	return out
}

//go:nosplit
func setGCPhase(x uint32) {
	_base.Atomicstore(&_base.Gcphase, x)
	_base.WriteBarrierEnabled = _base.Gcphase == _base.GCmark || _base.Gcphase == _base.GCmarktermination || _base.Gcphase == _base.GCscan
}

// gcBgCreditSlack is the amount of scan work credit background
// scanning can accumulate locally before updating
// gcController.bgScanCredit. Lower values give mutator assists more
// accurate accounting of background scanning. Higher values reduce
// memory contention.
const gcBgCreditSlack = 2000

const (
	GcBackgroundMode = iota // concurrent GC
	GcForceMode             // stop-the-world GC now
	GcForceBlockMode        // stop-the-world GC now and wait for sweep
)

func Gc(mode int) {
	// Timing/utilization tracking
	var stwprocs, maxprocs int32
	var tSweepTerm, tScan, tInstallWB, tMark, tMarkTerm int64

	// debug.gctrace variables
	var heap0, heap1, heap2, heapGoal uint64

	// memstats statistics
	var now, pauseStart, pauseNS int64

	// Ok, we're doing it!  Stop everybody else
	Semacquire(&Worldsema, false)

	// Pick up the remaining unswept/not being swept spans concurrently
	//
	// This shouldn't happen if we're being invoked in background
	// mode since proportional sweep should have just finished
	// sweeping everything, but rounding errors, etc, may leave a
	// few spans unswept. In forced mode, this is necessary since
	// GC can be forced at any point in the sweeping cycle.
	for Gosweepone() != ^uintptr(0) {
		Sweep.Nbgsweep++
	}

	if _base.Trace.Enabled {
		traceGCStart()
	}

	if mode == GcBackgroundMode {
		gcBgMarkStartWorkers()
	}
	now = _base.Nanotime()
	stwprocs, maxprocs = gcprocs(), _base.Gomaxprocs
	tSweepTerm = now
	heap0 = _base.Memstats.Heap_live

	pauseStart = now
	_base.Systemstack(StopTheWorldWithSema)
	_base.Systemstack(finishsweep_m) // finish sweep before we start concurrent scan.
	// clearpools before we start the GC. If we wait they memory will not be
	// reclaimed until the next GC cycle.
	clearpools()

	gcResetMarkState()

	if mode == GcBackgroundMode { // Do as much work concurrently as possible
		_base.GcController.StartCycle()
		heapGoal = _base.GcController.HeapGoal

		_base.Systemstack(func() {
			// Enter scan phase. This enables write
			// barriers to track changes to stack frames
			// above the stack barrier.
			//
			// TODO: This has evolved to the point where
			// we carefully ensure invariants we no longer
			// depend on. Either:
			//
			// 1) Enable full write barriers for the scan,
			// but eliminate the ragged barrier below
			// (since the start the world ensures all Ps
			// have observed the write barrier enable) and
			// consider draining during the scan.
			//
			// 2) Only enable write barriers for writes to
			// the stack at this point, and then enable
			// write barriers for heap writes when we
			// enter the mark phase. This means we cannot
			// drain in the scan phase and must perform a
			// ragged barrier to ensure all Ps have
			// enabled heap write barriers before we drain
			// or enable assists.
			//
			// 3) Don't install stack barriers over frame
			// boundaries where there are up-pointers.
			setGCPhase(_base.GCscan)

			gcBgMarkPrepare() // Must happen before assist enable.

			// At this point all Ps have enabled the write
			// barrier, thus maintaining the no white to
			// black invariant. Enable mutator assists to
			// put back-pressure on fast allocating
			// mutators.
			_base.Atomicstore(&_base.GcBlackenEnabled, 1)

			// Concurrent scan.
			StartTheWorldWithSema()
			now = _base.Nanotime()
			pauseNS += now - pauseStart
			tScan = now
			_base.GcController.AssistStartTime = now
			gcscan_m()

			// Enter mark phase.
			tInstallWB = _base.Nanotime()
			setGCPhase(_base.GCmark)
			// Ensure all Ps have observed the phase
			// change and have write barriers enabled
			// before any blackening occurs.
			forEachP(func(*_base.P) {})
		})
		// Concurrent mark.
		tMark = _base.Nanotime()

		// Enable background mark workers and wait for
		// background mark completion.
		_base.GcController.BgMarkStartTime = _base.Nanotime()
		_base.Work.BgMark1.Clear()
		_base.Work.BgMark1.Wait()

		// The global work list is empty, but there can still be work
		// sitting in the per-P work caches and there can be more
		// objects reachable from global roots since they don't have write
		// barriers. Rescan some roots and flush work caches.
		_base.Systemstack(func() {
			// rescan global data and bss.
			markroot(nil, _base.RootData)
			markroot(nil, _base.RootBss)

			// Disallow caching workbufs.
			_base.GcBlackenPromptly = true

			// Flush all currently cached workbufs. This
			// also forces any remaining background
			// workers out of their loop.
			forEachP(func(_p_ *_base.P) {
				_p_.Gcw.Dispose()
			})
		})

		// Wait for this more aggressive background mark to complete.
		_base.Work.BgMark2.Clear()
		_base.Work.BgMark2.Wait()

		// Begin mark termination.
		now = _base.Nanotime()
		tMarkTerm = now
		pauseStart = now
		_base.Systemstack(StopTheWorldWithSema)
		// The gcphase is _GCmark, it will transition to _GCmarktermination
		// below. The important thing is that the wb remains active until
		// all marking is complete. This includes writes made by the GC.

		// Flush the gcWork caches. This must be done before
		// endCycle since endCycle depends on statistics kept
		// in these caches.
		gcFlushGCWork()

		_base.GcController.EndCycle()
	} else {
		// For non-concurrent GC (mode != gcBackgroundMode)
		// The g stacks have not been scanned so clear g state
		// such that mark termination scans all stacks.
		gcResetGState()

		t := _base.Nanotime()
		tScan, tInstallWB, tMark, tMarkTerm = t, t, t, t
		heapGoal = heap0
	}

	// World is stopped.
	// Start marktermination which includes enabling the write barrier.
	_base.Atomicstore(&_base.GcBlackenEnabled, 0)
	_base.GcBlackenPromptly = false
	setGCPhase(_base.GCmarktermination)

	heap1 = _base.Memstats.Heap_live
	startTime := _base.Nanotime()

	mp := _base.Acquirem()
	mp.Preemptoff = "gcing"
	_g_ := _base.Getg()
	_g_.M.Traceback = 2
	gp := _g_.M.Curg
	_base.Casgstatus(gp, _base.Grunning, _base.Gwaiting)
	gp.Waitreason = "garbage collection"

	// Run gc on the g0 stack.  We do this so that the g stack
	// we're currently running on will no longer change.  Cuts
	// the root set down a bit (g0 stacks are not scanned, and
	// we don't need to scan gc's internal state).  We also
	// need to switch to g0 so we can shrink the stack.
	_base.Systemstack(func() {
		gcMark(startTime)
		// Must return immediately.
		// The outer function's stack may have moved
		// during gcMark (it shrinks stacks, including the
		// outer function's stack), so we must not refer
		// to any of its variables. Return back to the
		// non-system stack to pick up the new addresses
		// before continuing.
	})

	_base.Systemstack(func() {
		heap2 = _base.Work.BytesMarked
		if _base.Debug.Gccheckmark > 0 {
			// Run a full stop-the-world mark using checkmark bits,
			// to check that we didn't forget to mark anything during
			// the concurrent mark process.
			gcResetGState() // Rescan stacks
			gcResetMarkState()
			initCheckmarks()
			gcMark(startTime)
			clearCheckmarks()
		}

		// marking is complete so we can turn the write barrier off
		setGCPhase(_base.GCoff)
		gcSweep(mode)

		if _base.Debug.Gctrace > 1 {
			startTime = _base.Nanotime()
			// The g stacks have been scanned so
			// they have gcscanvalid==true and gcworkdone==true.
			// Reset these so that all stacks will be rescanned.
			gcResetGState()
			gcResetMarkState()
			finishsweep_m()

			// Still in STW but gcphase is _GCoff, reset to _GCmarktermination
			// At this point all objects will be found during the gcMark which
			// does a complete STW mark and object scan.
			setGCPhase(_base.GCmarktermination)
			gcMark(startTime)
			setGCPhase(_base.GCoff) // marking is done, turn off wb.
			gcSweep(mode)
		}
	})

	_g_.M.Traceback = 0
	_base.Casgstatus(gp, _base.Gwaiting, _base.Grunning)

	if _base.Trace.Enabled {
		traceGCDone()
	}

	// all done
	mp.Preemptoff = ""

	if _base.Gcphase != _base.GCoff {
		_base.Throw("gc done but gcphase != _GCoff")
	}

	// Update timing memstats
	now, unixNow := _base.Nanotime(), Unixnanotime()
	pauseNS += now - pauseStart
	_base.Atomicstore64(&_base.Memstats.Last_gc, uint64(unixNow)) // must be Unix time to make sense to user
	_base.Memstats.Pause_ns[_base.Memstats.Numgc%uint32(len(_base.Memstats.Pause_ns))] = uint64(pauseNS)
	_base.Memstats.Pause_end[_base.Memstats.Numgc%uint32(len(_base.Memstats.Pause_end))] = uint64(unixNow)
	_base.Memstats.Pause_total_ns += uint64(pauseNS)

	// Update work.totaltime.
	sweepTermCpu := int64(stwprocs) * (tScan - tSweepTerm)
	scanCpu := tInstallWB - tScan
	installWBCpu := int64(0)
	// We report idle marking time below, but omit it from the
	// overall utilization here since it's "free".
	markCpu := _base.GcController.AssistTime + _base.GcController.DedicatedMarkTime + _base.GcController.FractionalMarkTime
	markTermCpu := int64(stwprocs) * (now - tMarkTerm)
	cycleCpu := sweepTermCpu + scanCpu + installWBCpu + markCpu + markTermCpu
	_base.Work.Totaltime += cycleCpu

	// Compute overall GC CPU utilization.
	totalCpu := _base.Sched.Totaltime + (now-_base.Sched.Procresizetime)*int64(_base.Gomaxprocs)
	_base.Memstats.Gc_cpu_fraction = float64(_base.Work.Totaltime) / float64(totalCpu)

	_base.Memstats.Numgc++

	_base.Systemstack(StartTheWorldWithSema)
	Semrelease(&Worldsema)

	_base.Releasem(mp)
	mp = nil

	if _base.Debug.Gctrace > 0 {
		tEnd := now
		util := int(_base.Memstats.Gc_cpu_fraction * 100)

		var sbuf [24]byte
		_base.Printlock()
		print("gc ", _base.Memstats.Numgc,
			" @", string(itoaDiv(sbuf[:], uint64(tSweepTerm-RuntimeInitTime)/1e6, 3)), "s ",
			util, "%: ")
		prev := tSweepTerm
		for i, ns := range []int64{tScan, tInstallWB, tMark, tMarkTerm, tEnd} {
			if i != 0 {
				print("+")
			}
			print(string(fmtNSAsMS(sbuf[:], uint64(ns-prev))))
			prev = ns
		}
		print(" ms clock, ")
		for i, ns := range []int64{sweepTermCpu, scanCpu, installWBCpu, _base.GcController.AssistTime, _base.GcController.DedicatedMarkTime + _base.GcController.FractionalMarkTime, _base.GcController.IdleMarkTime, markTermCpu} {
			if i == 4 || i == 5 {
				// Separate mark time components with /.
				print("/")
			} else if i != 0 {
				print("+")
			}
			print(string(fmtNSAsMS(sbuf[:], uint64(ns))))
		}
		print(" ms cpu, ",
			heap0>>20, "->", heap1>>20, "->", heap2>>20, " MB, ",
			heapGoal>>20, " MB goal, ",
			maxprocs, " P")
		if mode != GcBackgroundMode {
			print(" (forced)")
		}
		print("\n")
		_base.Printunlock()
	}
	Sweep.Nbgsweep = 0
	Sweep.npausesweep = 0

	// now that gc is done, kick off finalizer thread if needed
	if !_base.XConcurrentSweep {
		// give the queued finalizers, if any, a chance to run
		Gosched()
	}
}

// gcBgMarkStartWorkers prepares background mark worker goroutines.
// These goroutines will not run until the mark phase, but they must
// be started while the work is not stopped and from a regular G
// stack. The caller must hold worldsema.
func gcBgMarkStartWorkers() {
	// Background marking is performed by per-P G's. Ensure that
	// each P has a background GC G.
	for _, p := range &_base.Allp {
		if p == nil || p.Status == _base.Pdead {
			break
		}
		if p.GcBgMarkWorker == nil {
			go GcBgMarkWorker(p)
			_base.Notetsleepg(&_base.Work.BgMarkReady, -1)
			_base.Noteclear(&_base.Work.BgMarkReady)
		}
	}
}

// gcBgMarkPrepare sets up state for background marking.
// Mutator assists must not yet be enabled.
func gcBgMarkPrepare() {
	// Background marking will stop when the work queues are empty
	// and there are no more workers (note that, since this is
	// concurrent, this may be a transient state, but mark
	// termination will clean it up). Between background workers
	// and assists, we don't really know how many workers there
	// will be, so we pretend to have an arbitrarily large number
	// of workers, almost all of which are "waiting". While a
	// worker is working it decrements nwait. If nproc == nwait,
	// there are no workers.
	_base.Work.Nproc = ^uint32(0)
	_base.Work.Nwait = ^uint32(0)

	// Reset background mark completion points.
	_base.Work.BgMark1.Done = 1
	_base.Work.BgMark2.Done = 1
}

func GcBgMarkWorker(p *_base.P) {
	// Register this G as the background mark worker for p.
	if p.GcBgMarkWorker != nil {
		_base.Throw("P already has a background mark worker")
	}
	gp := _base.Getg()

	mp := _base.Acquirem()
	p.GcBgMarkWorker = gp
	// After this point, the background mark worker is scheduled
	// cooperatively by gcController.findRunnable. Hence, it must
	// never be preempted, as this would put it into _Grunnable
	// and put it on a run queue. Instead, when the preempt flag
	// is set, this puts itself into _Gwaiting to be woken up by
	// gcController.findRunnable at the appropriate time.
	_base.Notewakeup(&_base.Work.BgMarkReady)
	for {
		// Go to sleep until woken by gcContoller.findRunnable.
		// We can't releasem yet since even the call to gopark
		// may be preempted.
		_base.Gopark(func(g *_base.G, mp unsafe.Pointer) bool {
			_base.Releasem((*_base.M)(mp))
			return true
		}, unsafe.Pointer(mp), "mark worker (idle)", _base.TraceEvGoBlock, 0)

		// Loop until the P dies and disassociates this
		// worker. (The P may later be reused, in which case
		// it will get a new worker.)
		if p.GcBgMarkWorker != gp {
			break
		}

		// Disable preemption so we can use the gcw. If the
		// scheduler wants to preempt us, we'll stop draining,
		// dispose the gcw, and then preempt.
		mp = _base.Acquirem()

		if _base.GcBlackenEnabled == 0 {
			_base.Throw("gcBgMarkWorker: blackening not enabled")
		}

		startTime := _base.Nanotime()

		decnwait := _base.Xadd(&_base.Work.Nwait, -1)
		if decnwait == _base.Work.Nproc {
			println("runtime: work.nwait=", decnwait, "work.nproc=", _base.Work.Nproc)
			_base.Throw("work.nwait was > work.nproc")
		}

		done := false
		switch p.GcMarkWorkerMode {
		default:
			_base.Throw("gcBgMarkWorker: unexpected gcMarkWorkerMode")
		case _base.GcMarkWorkerDedicatedMode:
			_base.GcDrain(&p.Gcw, gcBgCreditSlack)
			// gcDrain did the xadd(&work.nwait +1) to
			// match the decrement above. It only returns
			// at a mark completion point.
			done = true
			if !p.Gcw.Empty() {
				_base.Throw("gcDrain returned with buffer")
			}
		case _base.GcMarkWorkerFractionalMode, _base.GcMarkWorkerIdleMode:
			gcDrainUntilPreempt(&p.Gcw, gcBgCreditSlack)

			// If we are nearing the end of mark, dispose
			// of the cache promptly. We must do this
			// before signaling that we're no longer
			// working so that other workers can't observe
			// no workers and no work while we have this
			// cached, and before we compute done.
			if _base.GcBlackenPromptly {
				p.Gcw.Dispose()
			}

			// Was this the last worker and did we run out
			// of work?
			incnwait := _base.Xadd(&_base.Work.Nwait, +1)
			if incnwait > _base.Work.Nproc {
				println("runtime: p.gcMarkWorkerMode=", p.GcMarkWorkerMode,
					"work.nwait=", incnwait, "work.nproc=", _base.Work.Nproc)
				_base.Throw("work.nwait > work.nproc")
			}
			done = incnwait == _base.Work.Nproc && _base.Work.Full == 0 && _base.Work.Partial == 0
		}

		// If this worker reached a background mark completion
		// point, signal the main GC goroutine.
		if done {
			if _base.GcBlackenPromptly {
				if _base.Work.BgMark1.Done == 0 {
					_base.Throw("completing mark 2, but bgMark1.done == 0")
				}
				_base.Work.BgMark2.Complete()
			} else {
				_base.Work.BgMark1.Complete()
			}
		}

		duration := _base.Nanotime() - startTime
		switch p.GcMarkWorkerMode {
		case _base.GcMarkWorkerDedicatedMode:
			_base.Xaddint64(&_base.GcController.DedicatedMarkTime, duration)
			_base.Xaddint64(&_base.GcController.DedicatedMarkWorkersNeeded, 1)
		case _base.GcMarkWorkerFractionalMode:
			_base.Xaddint64(&_base.GcController.FractionalMarkTime, duration)
			_base.Xaddint64(&_base.GcController.FractionalMarkWorkersNeeded, 1)
		case _base.GcMarkWorkerIdleMode:
			_base.Xaddint64(&_base.GcController.IdleMarkTime, duration)
		}
	}
}

// gcFlushGCWork disposes the gcWork caches of all Ps. The world must
// be stopped.
//go:nowritebarrier
func gcFlushGCWork() {
	// Gather all cached GC work. All other Ps are stopped, so
	// it's safe to manipulate their GC work caches.
	for i := 0; i < int(_base.Gomaxprocs); i++ {
		_base.Allp[i].Gcw.Dispose()
	}
}

// gcMark runs the mark (or, for concurrent GC, mark termination)
// STW is in effect at this point.
//TODO go:nowritebarrier
func gcMark(start_time int64) {
	if _base.Debug.Allocfreetrace > 0 {
		tracegc()
	}

	if _base.Gcphase != _base.GCmarktermination {
		_base.Throw("in gcMark expecting to see gcphase as _GCmarktermination")
	}
	_base.Work.Tstart = start_time

	gcCopySpans() // TODO(rlh): should this be hoisted and done only once? Right now it is done for normal marking and also for checkmarking.

	// Make sure the per-P gcWork caches are empty. During mark
	// termination, these caches can still be used temporarily,
	// but must be disposed to the global lists immediately.
	gcFlushGCWork()

	_base.Work.Nwait = 0
	_base.Work.Ndone = 0
	_base.Work.Nproc = uint32(gcprocs())

	if _base.Trace.Enabled {
		_base.TraceGCScanStart()
	}

	parforsetup(_base.Work.Markfor, _base.Work.Nproc, uint32(_base.RootCount+_base.Allglen), false, markroot)
	if _base.Work.Nproc > 1 {
		_base.Noteclear(&_base.Work.Alldone)
		helpgc(int32(_base.Work.Nproc))
	}

	_base.Gchelperstart()
	_base.Parfordo(_base.Work.Markfor)

	var gcw _base.GcWork
	_base.GcDrain(&gcw, -1)
	gcw.Dispose()

	if _base.Work.Full != 0 {
		_base.Throw("work.full != 0")
	}
	if _base.Work.Partial != 0 {
		_base.Throw("work.partial != 0")
	}

	if _base.Work.Nproc > 1 {
		_base.Notesleep(&_base.Work.Alldone)
	}

	for i := 0; i < int(_base.Gomaxprocs); i++ {
		if _base.Allp[i].Gcw.Wbuf != 0 {
			_base.Throw("P has cached GC work at end of mark termination")
		}
	}

	if _base.Trace.Enabled {
		_base.TraceGCScanDone()
	}

	// TODO(austin): This doesn't have to be done during STW, as
	// long as we block the next GC cycle until this is done. Move
	// it after we start the world, but before dropping worldsema.
	// (See issue #11465.)
	freeStackSpans()

	Cachestats()

	// Compute the reachable heap size at the beginning of the
	// cycle. This is approximately the marked heap size at the
	// end (which we know) minus the amount of marked heap that
	// was allocated after marking began (which we don't know, but
	// is approximately the amount of heap that was allocated
	// since marking began).
	allocatedDuringCycle := _base.Memstats.Heap_live - _base.Work.InitialHeapLive
	if _base.Work.BytesMarked >= allocatedDuringCycle {
		_base.Memstats.Heap_reachable = _base.Work.BytesMarked - allocatedDuringCycle
	} else {
		// This can happen if most of the allocation during
		// the cycle never became reachable from the heap.
		// Just set the reachable heap approximation to 0 and
		// let the heapminimum kick in below.
		_base.Memstats.Heap_reachable = 0
	}

	// Trigger the next GC cycle when the allocated heap has grown
	// by triggerRatio over the reachable heap size. Assume that
	// we're in steady state, so the reachable heap size is the
	// same now as it was at the beginning of the GC cycle.
	_base.Memstats.Next_gc = uint64(float64(_base.Memstats.Heap_reachable) * (1 + _base.GcController.TriggerRatio))
	if _base.Memstats.Next_gc < _base.Heapminimum {
		_base.Memstats.Next_gc = _base.Heapminimum
	}
	if int64(_base.Memstats.Next_gc) < 0 {
		print("next_gc=", _base.Memstats.Next_gc, " bytesMarked=", _base.Work.BytesMarked, " heap_live=", _base.Memstats.Heap_live, " initialHeapLive=", _base.Work.InitialHeapLive, "\n")
		_base.Throw("next_gc underflow")
	}

	// Update other GC heap size stats.
	_base.Memstats.Heap_live = _base.Work.BytesMarked
	_base.Memstats.Heap_marked = _base.Work.BytesMarked
	_base.Memstats.Heap_scan = uint64(_base.GcController.ScanWork)

	minNextGC := _base.Memstats.Heap_live + _base.SweepMinHeapDistance*uint64(_base.Gcpercent)/100
	if _base.Memstats.Next_gc < minNextGC {
		// The allocated heap is already past the trigger.
		// This can happen if the triggerRatio is very low and
		// the reachable heap estimate is less than the live
		// heap size.
		//
		// Concurrent sweep happens in the heap growth from
		// heap_live to next_gc, so bump next_gc up to ensure
		// that concurrent sweep has some heap growth in which
		// to perform sweeping before we start the next GC
		// cycle.
		_base.Memstats.Next_gc = minNextGC
	}

	if _base.Trace.Enabled {
		TraceHeapAlloc()
		traceNextGC()
	}
}

func gcSweep(mode int) {
	if _base.Gcphase != _base.GCoff {
		_base.Throw("gcSweep being done but phase is not GCoff")
	}
	gcCopySpans()

	_base.Lock(&_base.Mheap_.Lock)
	_base.Mheap_.Sweepgen += 2
	_base.Mheap_.Sweepdone = 0
	Sweep.spanidx = 0
	_base.Unlock(&_base.Mheap_.Lock)

	if !_base.ConcurrentSweep || mode == GcForceBlockMode {
		// Special case synchronous sweep.
		// Record that no proportional sweeping has to happen.
		_base.Lock(&_base.Mheap_.Lock)
		_base.Mheap_.SweepPagesPerByte = 0
		_base.Mheap_.PagesSwept = 0
		_base.Unlock(&_base.Mheap_.Lock)
		// Sweep all spans eagerly.
		for Sweepone() != ^uintptr(0) {
			Sweep.npausesweep++
		}
		// Do an additional mProf_GC, because all 'free' events are now real as well.
		mProf_GC()
		mProf_GC()
		return
	}

	// Account how much sweeping needs to be done before the next
	// GC cycle and set up proportional sweep statistics.
	var pagesToSweep uintptr
	for _, s := range _base.Work.Spans {
		if s.State == _base.XMSpanInUse {
			pagesToSweep += s.Npages
		}
	}
	heapDistance := int64(_base.Memstats.Next_gc) - int64(_base.Memstats.Heap_live)
	// Add a little margin so rounding errors and concurrent
	// sweep are less likely to leave pages unswept when GC starts.
	heapDistance -= 1024 * 1024
	if heapDistance < _base.PageSize {
		// Avoid setting the sweep ratio extremely high
		heapDistance = _base.PageSize
	}
	_base.Lock(&_base.Mheap_.Lock)
	_base.Mheap_.SweepPagesPerByte = float64(pagesToSweep) / float64(heapDistance)
	_base.Mheap_.PagesSwept = 0
	_base.Mheap_.SpanBytesAlloc = 0
	_base.Unlock(&_base.Mheap_.Lock)

	// Background sweep.
	_base.Lock(&Sweep.Lock)
	if Sweep.Parked {
		Sweep.Parked = false
		_base.Ready(Sweep.G, 0)
	}
	_base.Unlock(&Sweep.Lock)
	mProf_GC()
}

func gcCopySpans() {
	// Cache runtime.mheap_.allspans in work.spans to avoid conflicts with
	// resizing/freeing allspans.
	// New spans can be created while GC progresses, but they are not garbage for
	// this round:
	//  - new stack spans can be created even while the world is stopped.
	//  - new malloc spans can be created during the concurrent sweep
	// Even if this is stop-the-world, a concurrent exitsyscall can allocate a stack from heap.
	_base.Lock(&_base.Mheap_.Lock)
	// Free the old cached mark array if necessary.
	if _base.Work.Spans != nil && &_base.Work.Spans[0] != &H_allspans[0] {
		_base.SysFree(unsafe.Pointer(&_base.Work.Spans[0]), uintptr(len(_base.Work.Spans))*unsafe.Sizeof(_base.Work.Spans[0]), &_base.Memstats.Other_sys)
	}
	// Cache the current array for sweeping.
	_base.Mheap_.Gcspans = _base.Mheap_.Allspans
	_base.Work.Spans = H_allspans
	_base.Unlock(&_base.Mheap_.Lock)
}

// gcResetGState resets the GC state of all G's and returns the length
// of allgs.
func gcResetGState() (numgs int) {
	// This may be called during a concurrent phase, so make sure
	// allgs doesn't change.
	_base.Lock(&_base.Allglock)
	for _, gp := range _base.Allgs {
		gp.Gcscandone = false  // set to true in gcphasework
		gp.Gcscanvalid = false // stack has not been scanned
		gp.Gcalloc = 0
		gp.Gcscanwork = 0
	}
	numgs = len(_base.Allgs)
	_base.Unlock(&_base.Allglock)
	return
}

// gcResetMarkState resets state prior to marking (concurrent or STW).
//
// TODO(austin): Merge with gcResetGState. See issue #11427.
func gcResetMarkState() {
	_base.Work.BytesMarked = 0
	_base.Work.InitialHeapLive = _base.Memstats.Heap_live
}

// Hooks for other packages

var Poolcleanup func()

func clearpools() {
	// clear sync.Pools
	if Poolcleanup != nil {
		Poolcleanup()
	}

	// Clear central sudog cache.
	// Leave per-P caches alone, they have strictly bounded size.
	// Disconnect cached list before dropping it on the floor,
	// so that a dangling ref to one entry does not pin all of them.
	_base.Lock(&_base.Sched.Sudoglock)
	var sg, sgnext *_base.Sudog
	for sg = _base.Sched.Sudogcache; sg != nil; sg = sgnext {
		sgnext = sg.Next
		sg.Next = nil
	}
	_base.Sched.Sudogcache = nil
	_base.Unlock(&_base.Sched.Sudoglock)

	// Clear central defer pools.
	// Leave per-P pools alone, they have strictly bounded size.
	_base.Lock(&_base.Sched.Deferlock)
	for i := range _base.Sched.Deferpool {
		// disconnect cached list before dropping it on the floor,
		// so that a dangling ref to one entry does not pin all of them.
		var d, dlink *_base.Defer
		for d = _base.Sched.Deferpool[i]; d != nil; d = dlink {
			dlink = d.Link
			d.Link = nil
		}
		_base.Sched.Deferpool[i] = nil
	}
	_base.Unlock(&_base.Sched.Deferlock)

	for _, p := range &_base.Allp {
		if p == nil {
			break
		}
		// clear tinyalloc pool
		if c := p.Mcache; c != nil {
			c.Tiny = nil
			c.Tinyoffset = 0
		}
	}
}

// itoaDiv formats val/(10**dec) into buf.
func itoaDiv(buf []byte, val uint64, dec int) []byte {
	i := len(buf) - 1
	idec := i - dec
	for val >= 10 || i >= idec {
		buf[i] = byte(val%10 + '0')
		i--
		if i == idec {
			buf[i] = '.'
			i--
		}
		val /= 10
	}
	buf[i] = byte(val + '0')
	return buf[i:]
}

// fmtNSAsMS nicely formats ns nanoseconds as milliseconds.
func fmtNSAsMS(buf []byte, ns uint64) []byte {
	if ns >= 10e6 {
		// Format as whole milliseconds.
		return itoaDiv(buf, ns/1e6, 0)
	}
	// Format two digits of precision, with at most three decimal places.
	x := ns / 1e3
	if x == 0 {
		buf[0] = '0'
		return buf[:1]
	}
	dec := 3
	for x >= 100 {
		x /= 10
		dec--
	}
	return itoaDiv(buf, x, dec)
}
