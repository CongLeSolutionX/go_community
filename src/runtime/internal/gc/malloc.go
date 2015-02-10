// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gc

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
	_sem "runtime/internal/sem"
)

// For now this must be bracketed with a stoptheworld and a starttheworld to ensure
// all go routines see the new barrier.
func gcinstallmarkwb() {
	_sched.Gcphase = _sched.GCmark
}

// force = 0 - start concurrent GC
// force = 1 - do STW GC regardless of current heap usage
// force = 2 - go STW GC and eager sweep
func Gogc(force int32) {
	// The gc is turned off (via enablegc) until the bootstrap has completed.
	// Also, malloc gets called in the guts of a number of libraries that might be
	// holding locks. To avoid deadlocks during stoptheworld, don't bother
	// trying to run gc while holding a lock. The next mallocgc without a lock
	// will do the gc instead.

	mp := _sched.Acquirem()
	if gp := _core.Getg(); gp == mp.G0 || mp.Locks > 1 || !_lock.Memstats.Enablegc || _lock.Panicking != 0 || Gcpercent < 0 {
		_sched.Releasem(mp)
		return
	}
	_sched.Releasem(mp)
	mp = nil

	if force == 0 {
		_lock.Lock(&Bggc.lock)
		if !Bggc.started {
			Bggc.Working = 1
			Bggc.started = true
			go backgroundgc()
		} else if Bggc.Working == 0 {
			Bggc.Working = 1
			_sched.Ready(Bggc.g)
		}
		_lock.Unlock(&Bggc.lock)
	} else {
		gcwork(force)
	}
}

func gcwork(force int32) {

	_sem.Semacquire(&Worldsema, false)

	// Pick up the remaining unswept/not being swept spans concurrently
	for gosweepone() != ^uintptr(0) {
		sweep.nbgsweep++
	}

	// Ok, we're doing it!  Stop everybody else

	mp := _sched.Acquirem()
	mp.Preemptoff = "gcing"
	_sched.Releasem(mp)
	Gctimer.Count++
	if force == 0 {
		Gctimer.Cycle.sweepterm = _lock.Nanotime()
	}

	if _sched.Trace.Enabled {
		_sched.TraceGoSched()
		traceGCStart()
	}

	// Pick up the remaining unswept/not being swept spans before we STW
	for gosweepone() != ^uintptr(0) {
		sweep.nbgsweep++
	}
	_lock.Systemstack(Stoptheworld)
	_lock.Systemstack(finishsweep_m) // finish sweep before we start concurrent scan.
	if force == 0 {                  // Do as much work concurrently as possible
		_sched.Gcphase = _sched.GCscan
		_lock.Systemstack(Starttheworld)
		Gctimer.Cycle.scan = _lock.Nanotime()
		// Do a concurrent heap scan before we stop the world.
		_lock.Systemstack(gcscan_m)
		Gctimer.Cycle.installmarkwb = _lock.Nanotime()
		_lock.Systemstack(Stoptheworld)
		_lock.Systemstack(gcinstallmarkwb)
		_lock.Systemstack(Starttheworld)
		Gctimer.Cycle.mark = _lock.Nanotime()
		_lock.Systemstack(gcmark_m)
		Gctimer.Cycle.markterm = _lock.Nanotime()
		_lock.Systemstack(Stoptheworld)
		_lock.Systemstack(gcinstalloffwb_m)
	} else {
		// For non-concurrent GC (force != 0) g stack have not been scanned so
		// set gcscanvalid such that mark termination scans all stacks.
		// No races here since we are in a STW phase.
		for _, gp := range _lock.Allgs {
			gp.Gcworkdone = false  // set to true in gcphasework
			gp.Gcscanvalid = false // stack has not been scanned
		}
	}

	startTime := _lock.Nanotime()
	if mp != _sched.Acquirem() {
		_lock.Throw("gogc: rescheduled")
	}

	clearpools()

	// Run gc on the g0 stack.  We do this so that the g stack
	// we're currently running on will no longer change.  Cuts
	// the root set down a bit (g0 stacks are not scanned, and
	// we don't need to scan gc's internal state).  We also
	// need to switch to g0 so we can shrink the stack.
	n := 1
	if _lock.Debug.Gctrace > 1 {
		n = 2
	}
	eagersweep := force >= 2
	for i := 0; i < n; i++ {
		if i > 0 {
			// refresh start time if doing a second GC
			startTime = _lock.Nanotime()
		}
		// switch to g0, call gc, then switch back
		_lock.Systemstack(func() {
			gc_m(startTime, eagersweep)
		})
	}

	_lock.Systemstack(func() {
		gccheckmark_m(startTime, eagersweep)
	})

	if _sched.Trace.Enabled {
		traceGCDone()
		_sched.TraceGoStart()
	}

	// all done
	mp.Preemptoff = ""

	if force == 0 {
		Gctimer.Cycle.sweep = _lock.Nanotime()
	}

	_sem.Semrelease(&Worldsema)

	if force == 0 {
		if Gctimer.Verbose > 1 {
			GCprinttimes()
		} else if Gctimer.Verbose > 0 {
			calctimes() // ignore result
		}
	}

	_lock.Systemstack(Starttheworld)

	_sched.Releasem(mp)
	mp = nil

	// now that gc is done, kick off finalizer thread if needed
	if !_sched.XConcurrentSweep {
		// give the queued finalizers, if any, a chance to run
		Gosched()
	}
}

// gctimes records the time in nanoseconds of each phase of the concurrent GC.
type gctimes struct {
	sweepterm     int64 // stw
	scan          int64
	installmarkwb int64 // stw
	mark          int64
	markterm      int64 // stw
	sweep         int64
}

// gcchronograph holds timer information related to GC phases
// max records the maximum time spent in each GC phase since GCstarttimes.
// total records the total time spent in each GC phase since GCstarttimes.
// cycle records the absolute time (as returned by nanoseconds()) that each GC phase last started at.
type Gcchronograph struct {
	Count    int64
	Verbose  int64
	Maxpause int64
	Max      gctimes
	Total    gctimes
	Cycle    gctimes
}

var Gctimer Gcchronograph

// calctimes converts gctimer.cycle into the elapsed times, updates gctimer.total
// and updates gctimer.max with the max pause time.
func calctimes() gctimes {
	var times gctimes

	var max = func(a, b int64) int64 {
		if a > b {
			return a
		}
		return b
	}

	times.sweepterm = Gctimer.Cycle.scan - Gctimer.Cycle.sweepterm
	Gctimer.Total.sweepterm += times.sweepterm
	Gctimer.Max.sweepterm = max(Gctimer.Max.sweepterm, times.sweepterm)
	Gctimer.Maxpause = max(Gctimer.Maxpause, Gctimer.Max.sweepterm)

	times.scan = Gctimer.Cycle.installmarkwb - Gctimer.Cycle.scan
	Gctimer.Total.scan += times.scan
	Gctimer.Max.scan = max(Gctimer.Max.scan, times.scan)

	times.installmarkwb = Gctimer.Cycle.mark - Gctimer.Cycle.installmarkwb
	Gctimer.Total.installmarkwb += times.installmarkwb
	Gctimer.Max.installmarkwb = max(Gctimer.Max.installmarkwb, times.installmarkwb)
	Gctimer.Maxpause = max(Gctimer.Maxpause, Gctimer.Max.installmarkwb)

	times.mark = Gctimer.Cycle.markterm - Gctimer.Cycle.mark
	Gctimer.Total.mark += times.mark
	Gctimer.Max.mark = max(Gctimer.Max.mark, times.mark)

	times.markterm = Gctimer.Cycle.sweep - Gctimer.Cycle.markterm
	Gctimer.Total.markterm += times.markterm
	Gctimer.Max.markterm = max(Gctimer.Max.markterm, times.markterm)
	Gctimer.Maxpause = max(Gctimer.Maxpause, Gctimer.Max.markterm)

	return times
}

// GCprinttimes prints latency information in nanoseconds about various
// phases in the GC. The information for each phase includes the maximum pause
// and total time since the most recent call to GCstarttimes as well as
// the information from the most recent Concurent GC cycle. Calls from the
// application to runtime.GC() are ignored.
func GCprinttimes() {
	if Gctimer.Verbose == 0 {
		println("GC timers not enabled")
		return
	}

	// Explicitly put times on the heap so printPhase can use it.
	times := new(gctimes)
	*times = calctimes()
	cycletime := Gctimer.Cycle.sweep - Gctimer.Cycle.sweepterm
	pause := times.sweepterm + times.installmarkwb + times.markterm
	gomaxprocs := GOMAXPROCS(-1)

	_sched.Printlock()
	print("GC: #", Gctimer.Count, " ", cycletime, "ns @", Gctimer.Cycle.sweepterm, " pause=", pause, " maxpause=", Gctimer.Maxpause, " goroutines=", Allglen, " gomaxprocs=", gomaxprocs, "\n")
	printPhase := func(label string, get func(*gctimes) int64, procs int) {
		print("GC:     ", label, " ", get(times), "ns\tmax=", get(&Gctimer.Max), "\ttotal=", get(&Gctimer.Total), "\tprocs=", procs, "\n")
	}
	printPhase("sweep term:", func(t *gctimes) int64 { return t.sweepterm }, gomaxprocs)
	printPhase("scan:      ", func(t *gctimes) int64 { return t.scan }, 1)
	printPhase("install wb:", func(t *gctimes) int64 { return t.installmarkwb }, gomaxprocs)
	printPhase("mark:      ", func(t *gctimes) int64 { return t.mark }, 1)
	printPhase("mark term: ", func(t *gctimes) int64 { return t.markterm }, gomaxprocs)
	_sched.Printunlock()
}
