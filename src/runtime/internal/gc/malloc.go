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

	_sem.Semacquire(&Worldsema, false)

	if force == 0 && _lock.Memstats.Heap_alloc < _lock.Memstats.Next_gc {
		// typically threads which lost the race to grab
		// worldsema exit here when gc is done.
		_sem.Semrelease(&Worldsema)
		return
	}

	// Pick up the remaining unswept/not being swept spans concurrently
	for gosweepone() != ^uintptr(0) {
		sweep.nbgsweep++
	}

	// Ok, we're doing it!  Stop everybody else

	startTime := _lock.Nanotime()
	mp = _sched.Acquirem()
	mp.Gcing = 1
	_sched.Releasem(mp)
	Gctimer.count++
	if force == 0 {
		Gctimer.cycle.sweepterm = _lock.Nanotime()
	}
	_lock.Systemstack(Stoptheworld)
	_lock.Systemstack(finishsweep_m) // finish sweep before we start concurrent scan.
	if force == 0 {                  // Do as much work concurrently as possible
		_lock.Systemstack(Starttheworld)
		Gctimer.cycle.scan = _lock.Nanotime()
		// Do a concurrent heap scan before we stop the world.
		_lock.Systemstack(gcscan_m)
		Gctimer.cycle.installmarkwb = _lock.Nanotime()
		_lock.Systemstack(Stoptheworld)
		gcinstallmarkwb()
		_lock.Systemstack(Starttheworld)
		Gctimer.cycle.mark = _lock.Nanotime()
		_lock.Systemstack(gcmark_m)
		Gctimer.cycle.markterm = _lock.Nanotime()
		_lock.Systemstack(Stoptheworld)
		_lock.Systemstack(gcinstalloffwb_m)
	}

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

	// all done
	mp.Gcing = 0

	if force == 0 {
		Gctimer.cycle.sweep = _lock.Nanotime()
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
	if !_sched.ConcurrentSweep {
		// give the queued finalizers, if any, a chance to run
		Gosched()
	}
}

// gctimes records the time in nanoseconds of each phase of the concurrent GC.
type gctimes struct {
	sweepterm     int64 // stw
	scan          int64 // stw
	installmarkwb int64
	mark          int64
	markterm      int64 // stw
	sweep         int64
}

// gcchronograph holds timer information related to GC phases
// max records the maximum time spent in each GC phase since GCstarttimes.
// total records the total time spent in each GC phase since GCstarttimes.
// cycle records the absolute time (as returned by nanoseconds()) that each GC phase last started at.
type Gcchronograph struct {
	count    int64
	Verbose  int64
	maxpause int64
	max      gctimes
	total    gctimes
	cycle    gctimes
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

	times.sweepterm = Gctimer.cycle.scan - Gctimer.cycle.sweepterm
	Gctimer.total.sweepterm += times.sweepterm
	Gctimer.max.sweepterm = max(Gctimer.max.sweepterm, times.sweepterm)
	Gctimer.maxpause = max(Gctimer.maxpause, Gctimer.max.sweepterm)

	times.scan = Gctimer.cycle.installmarkwb - Gctimer.cycle.scan
	Gctimer.total.scan += times.scan
	Gctimer.max.scan = max(Gctimer.max.scan, times.scan)

	times.installmarkwb = Gctimer.cycle.mark - Gctimer.cycle.installmarkwb
	Gctimer.total.installmarkwb += times.installmarkwb
	Gctimer.max.installmarkwb = max(Gctimer.max.installmarkwb, times.installmarkwb)
	Gctimer.maxpause = max(Gctimer.maxpause, Gctimer.max.installmarkwb)

	times.mark = Gctimer.cycle.markterm - Gctimer.cycle.mark
	Gctimer.total.mark += times.mark
	Gctimer.max.mark = max(Gctimer.max.mark, times.mark)

	times.markterm = Gctimer.cycle.sweep - Gctimer.cycle.markterm
	Gctimer.total.markterm += times.markterm
	Gctimer.max.markterm = max(Gctimer.max.markterm, times.markterm)
	Gctimer.maxpause = max(Gctimer.maxpause, Gctimer.max.markterm)

	return times
}

// GCprinttimes prints latency information in nanoseconds about various
// phases in the GC. The information for each phase includes the maximum pause
// and total time since the most recent call to GCstarttimes as well as
// the information from the most recent Concurent GC cycle. Calls from the
// application to runtime.GC() are ignored.
func GCprinttimes() {
	times := calctimes()
	println("GC:", Gctimer.count, "maxpause=", Gctimer.maxpause, "Go routines=", Allglen)
	println("          sweep termination: max=", Gctimer.max.sweepterm, "total=", Gctimer.total.sweepterm, "cycle=", times.sweepterm, "absolute time=", Gctimer.cycle.sweepterm)
	println("          scan:              max=", Gctimer.max.scan, "total=", Gctimer.total.scan, "cycle=", times.scan, "absolute time=", Gctimer.cycle.scan)
	println("          installmarkwb:     max=", Gctimer.max.installmarkwb, "total=", Gctimer.total.installmarkwb, "cycle=", times.installmarkwb, "absolute time=", Gctimer.cycle.installmarkwb)
	println("          mark:              max=", Gctimer.max.mark, "total=", Gctimer.total.mark, "cycle=", times.mark, "absolute time=", Gctimer.cycle.mark)
	println("          markterm:          max=", Gctimer.max.markterm, "total=", Gctimer.total.markterm, "cycle=", times.markterm, "absolute time=", Gctimer.cycle.markterm)
	cycletime := Gctimer.cycle.sweep - Gctimer.cycle.sweepterm
	println("          Total cycle time =", cycletime)
	totalstw := times.sweepterm + times.installmarkwb + times.markterm
	println("          Cycle STW time     =", totalstw)
}

// GC runs a garbage collection.
func GC() {
	Gogc(2)
}
