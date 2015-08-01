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

package base

const (
	DebugGC         = 0
	ConcurrentSweep = true
	FinBlockSize    = 4 * 1024
	RootData        = 0
	RootBss         = 1
	RootFinalizers  = 2
	RootSpans       = 3
	RootFlushCaches = 4
	RootCount       = 5

	// firstStackBarrierOffset is the approximate byte offset at
	// which to place the first stack barrier from the current SP.
	// This is a lower bound on how much stack will have to be
	// re-scanned during mark termination. Subsequent barriers are
	// placed at firstStackBarrierOffset * 2^n offsets.
	//
	// For debugging, this can be set to 0, which will install a
	// stack barrier at every frame. If you do this, you may also
	// have to raise _StackMin, since the stack barrier
	// bookkeeping will use a large amount of each stack.
	FirstStackBarrierOffset = 1024
	DebugStackBarrier       = false
)

// heapminimum is the minimum heap size at which to trigger GC.
// For small heaps, this overrides the usual GOGC*live set rule.
//
// When there is a very small live set but a lot of allocation, simply
// collecting when the heap reaches GOGC*live results in many GC
// cycles and high total per-GC overhead. This minimum amortizes this
// per-GC overhead while keeping the heap reasonably small.
//
// During initialization this is set to 4MB*GOGC/100. In the case of
// GOGC==0, this will set heapminimum to 0, resulting in constant
// collection even when the heap size is small, which is useful for
// debugging.
var Heapminimum uint64 = DefaultHeapMinimum

// defaultHeapMinimum is the value of heapminimum for GOGC==100.
const DefaultHeapMinimum = 4 << 20

// Initialized from $GOGC.  GOGC=off means no GC.
var Gcpercent int32

// Garbage collector phase.
// Indicates to write barrier and sychronization task to preform.
var Gcphase uint32
var WriteBarrierEnabled bool // compiler emits references to this in write barriers

// gcBlackenEnabled is 1 if mutator assists and background mark
// workers are allowed to blacken objects. This must only be set when
// gcphase == _GCmark.
var GcBlackenEnabled uint32

// gcBlackenPromptly indicates that optimizations that may
// hide work from the global work queue should be disabled.
//
// If gcBlackenPromptly is true, per-P gcWork caches should
// be flushed immediately and new objects should be allocated black.
//
// There is a tension between allocating objects white and
// allocating them black. If white and the objects die before being
// marked they can be collected during this GC cycle. On the other
// hand allocating them black will reduce _GCmarktermination latency
// since more work is done in the mark phase. This tension is resolved
// by allocating white until the mark phase is approaching its end and
// then allocating black for the remainder of the mark phase.
var GcBlackenPromptly bool

const (
	GCoff             = iota // GC not running; sweeping in background, write barrier disabled
	GCstw                    // unused state
	GCscan                   // GC collecting roots into workbufs, write barrier ENABLED
	GCmark                   // GC marking from workbufs, write barrier ENABLED
	GCmarktermination        // GC mark termination: allocate black, P's help GC, write barrier ENABLED
)

// gcMarkWorkerMode represents the mode that a concurrent mark worker
// should operate in.
//
// Concurrent marking happens through four different mechanisms. One
// is mutator assists, which happen in response to allocations and are
// not scheduled. The other three are variations in the per-P mark
// workers and are distinguished by gcMarkWorkerMode.
type gcMarkWorkerMode int

const (
	// gcMarkWorkerDedicatedMode indicates that the P of a mark
	// worker is dedicated to running that mark worker. The mark
	// worker should run without preemption until concurrent mark
	// is done.
	GcMarkWorkerDedicatedMode gcMarkWorkerMode = iota

	// gcMarkWorkerFractionalMode indicates that a P is currently
	// running the "fractional" mark worker. The fractional worker
	// is necessary when GOMAXPROCS*gcGoalUtilization is not an
	// integer. The fractional worker should run until it is
	// preempted and will be scheduled to pick up the fractional
	// part of GOMAXPROCS*gcGoalUtilization.
	GcMarkWorkerFractionalMode

	// gcMarkWorkerIdleMode indicates that a P is running the mark
	// worker because it has nothing else to do. The idle worker
	// should run until it is preempted and account its time
	// against gcController.idleMarkTime.
	GcMarkWorkerIdleMode
)

// gcController implements the GC pacing controller that determines
// when to trigger concurrent garbage collection and how much marking
// work to do in mutator assists and background marking.
//
// It uses a feedback control algorithm to adjust the memstats.next_gc
// trigger based on the heap growth and GC CPU utilization each cycle.
// This algorithm optimizes for heap growth to match GOGC and for CPU
// utilization between assist and background marking to be 25% of
// GOMAXPROCS. The high-level design of this algorithm is documented
// at https://golang.org/s/go15gcpacing.
var GcController = GcControllerState{
	// Initial trigger ratio guess.
	TriggerRatio: 7 / 8.0,
}

type GcControllerState struct {
	// scanWork is the total scan work performed this cycle. This
	// is updated atomically during the cycle. Updates may be
	// batched arbitrarily, since the value is only read at the
	// end of the cycle.
	//
	// Currently this is the bytes of heap scanned. For most uses,
	// this is an opaque unit of work, but for estimation the
	// definition is important.
	ScanWork int64

	// bgScanCredit is the scan work credit accumulated by the
	// concurrent background scan. This credit is accumulated by
	// the background scan and stolen by mutator assists. This is
	// updated atomically. Updates occur in bounded batches, since
	// it is both written and read throughout the cycle.
	BgScanCredit int64

	// assistTime is the nanoseconds spent in mutator assists
	// during this cycle. This is updated atomically. Updates
	// occur in bounded batches, since it is both written and read
	// throughout the cycle.
	AssistTime int64

	// dedicatedMarkTime is the nanoseconds spent in dedicated
	// mark workers during this cycle. This is updated atomically
	// at the end of the concurrent mark phase.
	DedicatedMarkTime int64

	// fractionalMarkTime is the nanoseconds spent in the
	// fractional mark worker during this cycle. This is updated
	// atomically throughout the cycle and will be up-to-date if
	// the fractional mark worker is not currently running.
	FractionalMarkTime int64

	// idleMarkTime is the nanoseconds spent in idle marking
	// during this cycle. This is updated atomically throughout
	// the cycle.
	IdleMarkTime int64

	// bgMarkStartTime is the absolute start time in nanoseconds
	// that the background mark phase started.
	BgMarkStartTime int64

	// heapGoal is the goal memstats.heap_live for when this cycle
	// ends. This is computed at the beginning of each cycle.
	HeapGoal uint64

	// dedicatedMarkWorkersNeeded is the number of dedicated mark
	// workers that need to be started. This is computed at the
	// beginning of each cycle and decremented atomically as
	// dedicated mark workers get started.
	DedicatedMarkWorkersNeeded int64

	// assistRatio is the ratio of allocated bytes to scan work
	// that should be performed by mutator assists. This is
	// computed at the beginning of each cycle.
	AssistRatio float64

	// fractionalUtilizationGoal is the fraction of wall clock
	// time that should be spent in the fractional mark worker.
	// For example, if the overall mark utilization goal is 25%
	// and GOMAXPROCS is 6, one P will be a dedicated mark worker
	// and this will be set to 0.5 so that 50% of the time some P
	// is in a fractional mark worker. This is computed at the
	// beginning of each cycle.
	fractionalUtilizationGoal float64

	// triggerRatio is the heap growth ratio at which the garbage
	// collection cycle should start. E.g., if this is 0.6, then
	// GC should start when the live heap has reached 1.6 times
	// the heap size marked by the previous cycle. This is updated
	// at the end of of each cycle.
	TriggerRatio float64

	// reviseTimer is a timer that triggers periodic revision of
	// control variables during the cycle.
	reviseTimer Timer

	_ [CacheLineSize]byte

	// fractionalMarkWorkersNeeded is the number of fractional
	// mark workers that need to be started. This is either 0 or
	// 1. This is potentially updated atomically at every
	// scheduling point (hence it gets its own cache line).
	FractionalMarkWorkersNeeded int64

	_ [CacheLineSize]byte
}

// startCycle resets the GC controller's state and computes estimates
// for a new GC cycle. The caller must hold worldsema.
func (c *GcControllerState) StartCycle() {
	c.ScanWork = 0
	c.BgScanCredit = 0
	c.AssistTime = 0
	c.DedicatedMarkTime = 0
	c.FractionalMarkTime = 0
	c.IdleMarkTime = 0

	// If this is the first GC cycle or we're operating on a very
	// small heap, fake heap_marked so it looks like next_gc is
	// the appropriate growth from heap_marked, even though the
	// real heap_marked may not have a meaningful value (on the
	// first cycle) or may be much smaller (resulting in a large
	// error response).
	if Memstats.Next_gc <= Heapminimum {
		Memstats.Heap_marked = uint64(float64(Memstats.Next_gc) / (1 + c.TriggerRatio))
		Memstats.Heap_reachable = Memstats.Heap_marked
	}

	// Compute the heap goal for this cycle
	c.HeapGoal = Memstats.Heap_reachable + Memstats.Heap_reachable*uint64(Gcpercent)/100

	// Compute the total mark utilization goal and divide it among
	// dedicated and fractional workers.
	totalUtilizationGoal := float64(Gomaxprocs) * gcGoalUtilization
	c.DedicatedMarkWorkersNeeded = int64(totalUtilizationGoal)
	c.fractionalUtilizationGoal = totalUtilizationGoal - float64(c.DedicatedMarkWorkersNeeded)
	if c.fractionalUtilizationGoal > 0 {
		c.FractionalMarkWorkersNeeded = 1
	} else {
		c.FractionalMarkWorkersNeeded = 0
	}

	// Clear per-P state
	for _, p := range &Allp {
		if p == nil {
			break
		}
		p.GcAssistTime = 0
	}

	// Compute initial values for controls that are updated
	// throughout the cycle.
	c.revise()

	// Set up a timer to revise periodically
	c.reviseTimer.F = func(interface{}, uintptr) {
		GcController.revise()
	}
	c.reviseTimer.period = 10 * 1000 * 1000
	c.reviseTimer.When = Nanotime() + c.reviseTimer.period
	Addtimer(&c.reviseTimer)
}

// revise updates the assist ratio during the GC cycle to account for
// improved estimates. This should be called periodically during
// concurrent mark.
func (c *GcControllerState) revise() {
	// Compute the expected scan work. This is a strict upper
	// bound on the possible scan work in the current heap.
	//
	// You might consider dividing this by 2 (or by
	// (100+GOGC)/100) to counter this over-estimation, but
	// benchmarks show that this has almost no effect on mean
	// mutator utilization, heap size, or assist time and it
	// introduces the danger of under-estimating and letting the
	// mutator outpace the garbage collector.
	scanWorkExpected := Memstats.Heap_scan

	// Compute the mutator assist ratio so by the time the mutator
	// allocates the remaining heap bytes up to next_gc, it will
	// have done (or stolen) the estimated amount of scan work.
	heapDistance := int64(c.HeapGoal) - int64(Work.InitialHeapLive)
	if heapDistance <= 1024*1024 {
		// heapDistance can be negative if GC start is delayed
		// or if the allocation that pushed heap_live over
		// next_gc is large or if the trigger is really close
		// to GOGC. We don't want to set the assist negative
		// (or divide by zero, or set it really high), so
		// enforce a minimum on the distance.
		heapDistance = 1024 * 1024
	}
	c.AssistRatio = float64(scanWorkExpected) / float64(heapDistance)
}

// endCycle updates the GC controller state at the end of the
// concurrent part of the GC cycle.
func (c *GcControllerState) EndCycle() {
	h_t := c.TriggerRatio // For debugging

	// Proportional response gain for the trigger controller. Must
	// be in [0, 1]. Lower values smooth out transient effects but
	// take longer to respond to phase changes. Higher values
	// react to phase changes quickly, but are more affected by
	// transient changes. Values near 1 may be unstable.
	const triggerGain = 0.5

	// Stop the revise timer
	Deltimer(&c.reviseTimer)

	// Compute next cycle trigger ratio. First, this computes the
	// "error" for this cycle; that is, how far off the trigger
	// was from what it should have been, accounting for both heap
	// growth and GC CPU utilization. We computing the actual heap
	// growth during this cycle and scale that by how far off from
	// the goal CPU utilization we were (to estimate the heap
	// growth if we had the desired CPU utilization). The
	// difference between this estimate and the GOGC-based goal
	// heap growth is the error.
	goalGrowthRatio := float64(Gcpercent) / 100
	actualGrowthRatio := float64(Memstats.Heap_live)/float64(Memstats.Heap_marked) - 1
	duration := Nanotime() - c.BgMarkStartTime

	// Assume background mark hit its utilization goal.
	utilization := gcGoalUtilization
	// Add assist utilization; avoid divide by zero.
	if duration > 0 {
		utilization += float64(c.AssistTime) / float64(duration*int64(Gomaxprocs))
	}

	triggerError := goalGrowthRatio - c.TriggerRatio - utilization/gcGoalUtilization*(actualGrowthRatio-c.TriggerRatio)

	// Finally, we adjust the trigger for next time by this error,
	// damped by the proportional gain.
	c.TriggerRatio += triggerGain * triggerError
	if c.TriggerRatio < 0 {
		// This can happen if the mutator is allocating very
		// quickly or the GC is scanning very slowly.
		c.TriggerRatio = 0
	} else if c.TriggerRatio > goalGrowthRatio*0.95 {
		// Ensure there's always a little margin so that the
		// mutator assist ratio isn't infinity.
		c.TriggerRatio = goalGrowthRatio * 0.95
	}

	if Debug.Gcpacertrace > 0 {
		// Print controller state in terms of the design
		// document.
		H_m_prev := Memstats.Heap_marked
		H_T := Memstats.Next_gc
		h_a := actualGrowthRatio
		H_a := Memstats.Heap_live
		h_g := goalGrowthRatio
		H_g := int64(float64(H_m_prev) * (1 + h_g))
		u_a := utilization
		u_g := gcGoalUtilization
		W_a := c.ScanWork
		print("pacer: H_m_prev=", H_m_prev,
			" h_t=", h_t, " H_T=", H_T,
			" h_a=", h_a, " H_a=", H_a,
			" h_g=", h_g, " H_g=", H_g,
			" u_a=", u_a, " u_g=", u_g,
			" W_a=", W_a,
			" goalΔ=", goalGrowthRatio-h_t,
			" actualΔ=", h_a-h_t,
			" u_a/u_g=", u_a/u_g,
			"\n")
	}
}

// findRunnableGCWorker returns the background mark worker for _p_ if it
// should be run. This must only be called when gcBlackenEnabled != 0.
func (c *GcControllerState) findRunnableGCWorker(_p_ *P) *G {
	if GcBlackenEnabled == 0 {
		Throw("gcControllerState.findRunnable: blackening not enabled")
	}
	if _p_.GcBgMarkWorker == nil {
		Throw("gcControllerState.findRunnable: no background mark worker")
	}
	if Work.BgMark1.Done != 0 && Work.BgMark2.Done != 0 {
		// Background mark is done. Don't schedule background
		// mark worker any more. (This is not just an
		// optimization. Without this we can spin scheduling
		// the background worker and having it return
		// immediately with no work to do.)
		return nil
	}

	decIfPositive := func(ptr *int64) bool {
		if *ptr > 0 {
			if Xaddint64(ptr, -1) >= 0 {
				return true
			}
			// We lost a race
			Xaddint64(ptr, +1)
		}
		return false
	}

	if decIfPositive(&c.DedicatedMarkWorkersNeeded) {
		// This P is now dedicated to marking until the end of
		// the concurrent mark phase.
		_p_.GcMarkWorkerMode = GcMarkWorkerDedicatedMode
		// TODO(austin): This P isn't going to run anything
		// else for a while, so kick everything out of its run
		// queue.
	} else {
		if _p_.Gcw.Wbuf == 0 && Work.Full == 0 && Work.Partial == 0 {
			// No work to be done right now. This can
			// happen at the end of the mark phase when
			// there are still assists tapering off. Don't
			// bother running background mark because
			// it'll just return immediately.
			return nil
		}
		if !decIfPositive(&c.FractionalMarkWorkersNeeded) {
			// No more workers are need right now.
			return nil
		}

		// This P has picked the token for the fractional worker.
		// Is the GC currently under or at the utilization goal?
		// If so, do more work.
		//
		// We used to check whether doing one time slice of work
		// would remain under the utilization goal, but that has the
		// effect of delaying work until the mutator has run for
		// enough time slices to pay for the work. During those time
		// slices, write barriers are enabled, so the mutator is running slower.
		// Now instead we do the work whenever we're under or at the
		// utilization work and pay for it by letting the mutator run later.
		// This doesn't change the overall utilization averages, but it
		// front loads the GC work so that the GC finishes earlier and
		// write barriers can be turned off sooner, effectively giving
		// the mutator a faster machine.
		//
		// The old, slower behavior can be restored by setting
		//	gcForcePreemptNS = forcePreemptNS.
		const gcForcePreemptNS = 0

		// TODO(austin): We could fast path this and basically
		// eliminate contention on c.fractionalMarkWorkersNeeded by
		// precomputing the minimum time at which it's worth
		// next scheduling the fractional worker. Then Ps
		// don't have to fight in the window where we've
		// passed that deadline and no one has started the
		// worker yet.
		//
		// TODO(austin): Shorter preemption interval for mark
		// worker to improve fairness and give this
		// finer-grained control over schedule?
		now := Nanotime() - GcController.BgMarkStartTime
		then := now + gcForcePreemptNS
		timeUsed := c.FractionalMarkTime + gcForcePreemptNS
		if then > 0 && float64(timeUsed)/float64(then) > c.fractionalUtilizationGoal {
			// Nope, we'd overshoot the utilization goal
			Xaddint64(&c.FractionalMarkWorkersNeeded, +1)
			return nil
		}
		_p_.GcMarkWorkerMode = GcMarkWorkerFractionalMode
	}

	// Run the background mark worker
	gp := _p_.GcBgMarkWorker
	Casgstatus(gp, Gwaiting, Grunnable)
	if Trace.Enabled {
		TraceGoUnpark(gp, 0)
	}
	return gp
}

// gcGoalUtilization is the goal CPU utilization for background
// marking as a fraction of GOMAXPROCS.
const gcGoalUtilization = 0.25

// bgMarkSignal synchronizes the GC coordinator and background mark workers.
type BgMarkSignal struct {
	// Workers race to cas to 1. Winner signals coordinator.
	Done uint32
	// Coordinator to wake up.
	lock Mutex
	g    *G
	wake bool
}

func (s *BgMarkSignal) Wait() {
	Lock(&s.lock)
	if s.wake {
		// Wakeup already happened
		Unlock(&s.lock)
	} else {
		s.g = Getg()
		Goparkunlock(&s.lock, "mark wait (idle)", TraceEvGoBlock, 1)
	}
	s.wake = false
	s.g = nil
}

// complete signals the completion of this phase of marking. This can
// be called multiple times during a cycle; only the first call has
// any effect.
func (s *BgMarkSignal) Complete() {
	if Cas(&s.Done, 0, 1) {
		// This is the first worker to reach this completion point.
		// Signal the main GC goroutine.
		Lock(&s.lock)
		if s.g == nil {
			// It hasn't parked yet.
			s.wake = true
		} else {
			Ready(s.g, 0)
		}
		Unlock(&s.lock)
	}
}

func (s *BgMarkSignal) Clear() {
	s.Done = 0
}

var Work struct {
	Full    uint64               // lock-free list of full blocks workbuf
	empty   uint64               // lock-free list of empty blocks workbuf
	Partial uint64               // lock-free list of partially filled blocks workbuf
	pad0    [CacheLineSize]uint8 // prevents false-sharing between full/empty and nproc/nwait
	Nproc   uint32
	Tstart  int64
	Nwait   uint32
	Ndone   uint32
	Alldone Note
	Markfor *Parfor

	BgMarkReady Note   // signal background mark worker has started
	bgMarkDone  uint32 // cas to 1 when at a background mark completion point
	// Background mark completion signaling

	// Coordination for the 2 parts of the mark phase.
	BgMark1 BgMarkSignal
	BgMark2 BgMarkSignal

	// Copy of mheap.allspans for marker or sweeper.
	Spans []*Mspan

	// totaltime is the CPU nanoseconds spent in GC since the
	// program started if debug.gctrace > 0.
	Totaltime int64

	// bytesMarked is the number of bytes marked this cycle. This
	// includes bytes blackened in scanned objects, noscan objects
	// that go straight to black, and permagrey objects scanned by
	// markroot during the concurrent scan phase. This is updated
	// atomically during the cycle. Updates may be batched
	// arbitrarily, since the value is only read at the end of the
	// cycle.
	//
	// Because of benign races during marking, this number may not
	// be the exact number of marked bytes, but it should be very
	// close.
	BytesMarked uint64

	// initialHeapLive is the value of memstats.heap_live at the
	// beginning of this GC cycle.
	InitialHeapLive uint64
}

// gcMarkWorkAvailable returns true if executing a mark worker
// on p is potentially useful.
func gcMarkWorkAvailable(p *P) bool {
	if !p.Gcw.empty() {
		return true
	}
	if Atomicload64(&Work.Full) != 0 || Atomicload64(&Work.Partial) != 0 {
		return true // global work available
	}
	return false
}

// Timing

//go:nowritebarrier
func gchelper() {
	_g_ := Getg()
	_g_.M.Traceback = 2
	Gchelperstart()

	if Trace.Enabled {
		TraceGCScanStart()
	}

	// parallel mark for over GC roots
	Parfordo(Work.Markfor)
	if Gcphase != GCscan {
		var gcw GcWork
		GcDrain(&gcw, -1) // blocks in getfull
		gcw.Dispose()
	}

	if Trace.Enabled {
		TraceGCScanDone()
	}

	nproc := Work.Nproc // work.nproc can change right after we increment work.ndone
	if Xadd(&Work.Ndone, +1) == nproc-1 {
		Notewakeup(&Work.Alldone)
	}
	_g_.M.Traceback = 0
}

func Gchelperstart() {
	_g_ := Getg()

	if _g_.M.Helpgc < 0 || _g_.M.Helpgc >= MaxGcproc {
		Throw("gchelperstart: bad m->helpgc")
	}
	if _g_ != _g_.M.G0 {
		Throw("gchelper not running on g0 stack")
	}
}
