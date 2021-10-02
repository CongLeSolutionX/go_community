// Copyright 2021 The Go Authors. All rights reserved.
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
var gcCPULimiter gcCPULimiterState

type gcCPULimiterState struct {
	limiting atomic.Bool
	bucket   struct {
		// Invariants:
		// - fill > 0
		// - capacity > 0
		// - fill <= capacity
		fill, capacity uint64
	}
	overflow uint64

	lastAssistTime      int64
	lastAssistTimeCheck int64
}

// enabled returns true if the CPU limiter is currently
// enabled, meaning the the Go GC should take action to
// limit CPU utilization.
func (l *gcCPULimiterState) enabled() bool {
	return l.limiting.Load()
}

func (l *gcCPULimiterState) update(gcEnabled bool, assistTime int64, now int64, gomaxprocs int32) {
	windowTotalTime := (now - l.lastAssistTimeCheck) * int64(gomaxprocs)
	windowGCTime := assistTime - l.lastAssistTime
	if gcEnabled {
		windowGCTime += int64(float64(windowTotalTime) * gcBackgroundUtilization)
	}
	l.accumulate(windowTotalTime-windowGCTime, windowGCTime)
	l.lastAssistTime = assistTime
	l.lastAssistTimeCheck = now
}

// accumulate adds updates the bucket and signals whether the limiter is enabled.
func (l *gcCPULimiterState) accumulate(mutatorTime, gcTime int64) {
	enabled := l.bucket.fill == l.bucket.capacity

	// Let's be careful about three things here:
	// 1. The addition and subtraction, for the invariants.
	// 2. Overflow.
	// 3. Excessive mutation of l.limiting, which is accessed
	//    by all assists, potentially more than once.
	change := gcTime - mutatorTime

	// Handle limiting case.
	if change > 0 && l.bucket.capacity-l.bucket.fill < uint64(change) {
		l.overflow += uint64(change) - (l.bucket.capacity - l.bucket.fill)
		l.bucket.fill = l.bucket.capacity
		if !enabled {
			l.limiting.Store(true)
		}
		return
	}

	// Handle non-limiting cases.
	if change < 0 && l.bucket.fill < uint64(-change) {
		// Bucket emptied.
		l.bucket.fill = 0
	} else {
		// All other cases.
		l.bucket.fill -= uint64(-change)
	}
	if enabled {
		l.limiting.Store(false)
	}
}

// capacityPerProc is the limiter's bucket capacity for each P in GOMAXPROCS.
const capacityPerProc = 1e9 // 1 second in nanoseconds

// resetCapacity updates the capacity based on GOMAXPROCS.
func (l *gcCPULimiterState) resetCapacity(nprocs int32) {
	enabled := l.bucket.fill == l.bucket.capacity

	l.bucket.capacity = uint64(nprocs) * capacityPerProc
	if l.bucket.fill > l.bucket.capacity {
		l.bucket.fill = l.bucket.capacity
		if !enabled {
			l.limiting.Store(true)
		}
	}
}
