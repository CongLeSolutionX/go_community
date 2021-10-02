// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import "runtime/internal/atomic"

// gcCPULimiter is a mechanism to limit GC CPU utilization in situations
// where it might become excessive and inhibit application progress (e.g.
// a death spiral).
//
// The core of the limiter is a leaky bucket mechanism that fills with GC
// CPU time and drains with mutator time. Because the bucket fills and
// drains with time directly (i.e. without any weighting), this effectively
// sets a very conservative limit of 50%. This limit could be enforced directly,
// however, but the purpose of the bucket is to accomodate spikes in GC CPU
// utilization without hurting throughput.
//
// Note that the bucket in the leaky bucket mechanism can never go negative,
// so the GC never gets credit for a lot of CPU time spent without the GC
// running. This is intentional, as an application that stays idle for, say,
// an entire day, could build up enough credit to fail to prevent a death
// spiral the following day. The bucket's capacity is the GC's only leeway.
//
// The capacity thus also sets the window the limiter considers. For example,
// if the capacity of the bucket is 1 cpu-second, then the limiter will not
// kick in until at least 1 full cpu-second in the last 2 cpu-second window
// is spent on GC CPU time.
var gcCPULimiter gcCPULimiterState

type gcCPULimiterState struct {
	lock mutex

	enabled atomic.Bool
	bucket  struct {
		// Invariants:
		// - fill >= 0
		// - capacity >= 0
		// - fill <= capacity
		fill, capacity uint64
	}
	// TODO(mknyszek): Export this as a runtime/metric to provide an estimate of
	// how much GC work is being dropped on the floor.
	overflow uint64

	// gcEnabled is an internal copy of gcBlackenEnabled that determines
	// whether the limiter tracks total assist time.
	//
	// gcBlackenEnabled isn't used directly so as to keep this structure
	// unit-testable.
	gcEnabled bool

	// transitioning is true when the GC is in a STW and transitioning between
	// the mark and sweep phases.
	transitioning bool

	// lastTotalAssistTime is the last value of a monotonically increasing
	// count of GC assist time, like gcController.assistTime.
	lastTotalAssistTime int64

	// lastUpdate is the nanotime timestamp of the last time update was called.
	lastUpdate int64

	// nprocs is an internal copy of gomaxprocs, used to determine total available
	// CPU time.
	//
	// gomaxprocs isn't used directly so as to keep this structure unit-testable.
	nprocs int32
}

// limiting returns true if the CPU limiter is currently enabled, meaning the Go GC
// should take action to limit CPU utilization.
//
// It is safe to call concurrently with other operations.
func (l *gcCPULimiterState) limiting() bool {
	return l.enabled.Load()
}

// startGCTransition notifies the limiter of a GC transition. totalAssistTime
// is the same as described for update. now must be the start of the STW pause
// for the GC transition.
//
// This call takes ownership of the limiter and disables all other means of
// updating the limiter. Release ownership by calling finishGCTransition.
//
// It is safe to call concurrently with other operations.
func (l *gcCPULimiterState) startGCTransition(enableGC bool, totalAssistTime, now int64) {
	lock(&l.lock)
	if l.gcEnabled == enableGC {
		throw("transitioning GC to the same state as before?")
	}
	// Flush whatever was left between the last update and now.
	l.updateLocked(totalAssistTime, now)
	if enableGC && totalAssistTime != 0 {
		throw("assist time must be zero on entry to a GC cycle")
	}
	l.gcEnabled = enableGC
	l.transitioning = true
	unlock(&l.lock)
}

// finishGCTransition notifies the limiter that the GC transition is complete
// and releases ownership of it. It also accumulates STW time in the bucket.
// now must be the timestamp from the end of the STW pause.
func (l *gcCPULimiterState) finishGCTransition(now int64) {
	lock(&l.lock)
	if !l.transitioning {
		throw("finishGCTransition called without starting one?")
	}
	// Count the full nprocs set of CPU time because the world is stopped
	// between startGCTransition and finishGCTransition. Even though the GC
	// isn't running on all CPUs, it is preventing user code from doing so,
	// so it might as well be.
	l.accumulate(0, (now-l.lastUpdate)*int64(l.nprocs))
	l.lastUpdate = now
	l.transitioning = false
	unlock(&l.lock)
}

// update updates the bucket given runtime-specific information. totalAssistTime must
// be a value that increases monotonically throughout the GC cycle, and is reset
// at the start of a new mark phase. now is the current monotonic time in nanoseconds.
//
// This is safe to call concurrently with other operations.
func (l *gcCPULimiterState) update(totalAssistTime int64, now int64) {
	lock(&l.lock)
	if !l.transitioning {
		l.updateLocked(totalAssistTime, now)
	}
	unlock(&l.lock)
}

// updatedLocked is the implementation of update. l.lock must be held.
func (l *gcCPULimiterState) updateLocked(totalAssistTime int64, now int64) {
	assertLockHeld(&l.lock)

	windowTotalTime := (now - l.lastUpdate) * int64(l.nprocs)
	l.lastUpdate = now
	if !l.gcEnabled {
		l.accumulate(windowTotalTime, 0)
		return
	}
	windowGCTime := totalAssistTime - l.lastTotalAssistTime
	windowGCTime += int64(float64(windowTotalTime) * gcBackgroundUtilization)
	l.accumulate(windowTotalTime-windowGCTime, windowGCTime)
	l.lastTotalAssistTime = totalAssistTime
}

// accumulate adds time to the bucket and signals whether the limiter is enabled.
//
// This is an internal function that deals just with the bucket. Prefer update.
// l.lock must be held.
func (l *gcCPULimiterState) accumulate(mutatorTime, gcTime int64) {
	assertLockHeld(&l.lock)

	headroom := l.bucket.capacity - l.bucket.fill
	enabled := headroom == 0

	// Let's be careful about three things here:
	// 1. The addition and subtraction, for the invariants.
	// 2. Overflow.
	// 3. Excessive mutation of l.enabled, which is accessed
	//    by all assists, potentially more than once.
	change := gcTime - mutatorTime

	// Handle limiting case.
	if change > 0 && headroom <= uint64(change) {
		l.overflow += uint64(change) - headroom
		l.bucket.fill = l.bucket.capacity
		if !enabled {
			l.enabled.Store(true)
		}
		return
	}

	// Handle non-limiting cases.
	if change < 0 && l.bucket.fill <= uint64(-change) {
		// Bucket emptied.
		l.bucket.fill = 0
	} else {
		// All other cases.
		l.bucket.fill -= uint64(-change)
	}
	if change != 0 && enabled {
		l.enabled.Store(false)
	}
}

// capacityPerProc is the limiter's bucket capacity for each P in GOMAXPROCS.
const capacityPerProc = 1e9 // 1 second in nanoseconds

// resetCapacity updates the capacity based on GOMAXPROCS. Must not be called
// while the GC is enabled.
//
// It is safe to call concurrently with other operations.
func (l *gcCPULimiterState) resetCapacity(now int64, nprocs int32) {
	lock(&l.lock)
	// Flush the rest of the time for this period.
	l.updateLocked(0, now)
	l.nprocs = nprocs

	l.bucket.capacity = uint64(nprocs) * capacityPerProc
	if l.bucket.fill > l.bucket.capacity {
		l.bucket.fill = l.bucket.capacity
		l.enabled.Store(true)
	} else if l.bucket.fill < l.bucket.capacity {
		l.enabled.Store(false)
	}
	unlock(&l.lock)
}
