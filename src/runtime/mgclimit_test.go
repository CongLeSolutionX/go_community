// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime_test

import (
	. "runtime"
	"testing"
	"time"
)

func TestGCCPULimiter(t *testing.T) {
	const procs = 14

	// Create mock time.
	ticks := int64(0)
	advance := func(d time.Duration) int64 {
		t.Helper()
		ticks += int64(d)
		return ticks
	}

	l := NewGCCPULimiter(ticks, procs)
	if l.Capacity() != procs*CapacityPerProc {
		t.Fatalf("unexpected capacity: %d", l.Capacity())
	}
	if l.Fill() != 0 {
		t.Fatalf("expected empty bucket to start")
	}

	// Test filling the bucket with just mutator time.

	l.Update(0, advance(10*time.Millisecond))
	l.Update(0, advance(1*time.Second))
	l.Update(0, advance(1*time.Hour))
	if l.Fill() != 0 {
		t.Fatalf("expected empty bucket from only accumulating mutator time, got fill of %d cpu-ns", l.Fill())
	}

	// Test transitioning the bucket to enable the GC.

	l.StartGCTransition(true, 0, advance(109*time.Millisecond))
	l.FinishGCTransition(advance(2*time.Millisecond + 1*time.Microsecond))

	if expect := uint64((2*time.Millisecond + 1*time.Microsecond) * procs); l.Fill() != expect {
		t.Fatalf("expected fill of %d, got %d cpu-ns", expect, l.Fill())
	}

	// Test passing time without assists during a GC. Specifically, just enough to drain the bucket to
	// exactly procs nanoseconds (easier to get to because of rounding).
	//
	// The window we need to drain the bucket is 1/(1-2*gcBackgroundUtilization) times the current fill:
	//
	//   fill + (window * procs * gcBackgroundUtilization - window * procs * (1-gcBackgroundUtilization)) = n
	//   fill = n - (window * procs * gcBackgroundUtilization - window * procs * (1-gcBackgroundUtilization))
	//   fill = n + window * procs * ((1-gcBackgroundUtilization) - gcBackgroundUtilization)
	//   fill = n + window * procs * (1-2*gcBackgroundUtilization)
	//   window = (fill - n) / (procs * (1-2*gcBackgroundUtilization)))
	//
	// And here we want n=procs:
	factor := (1 / (1 - 2*GCBackgroundUtilization))
	fill := (2*time.Millisecond + 1*time.Microsecond) * procs
	l.Update(0, advance(time.Duration(factor*float64(fill-procs)/procs)))
	if l.Fill() != procs {
		t.Fatalf("expected fill %d cpu-ns from draining after a GC started, got fill of %d cpu-ns", procs, l.Fill())
	}

	// Drain to zero for the rest of the test.
	l.Update(0, advance(2*procs*CapacityPerProc))
	if l.Fill() != 0 {
		t.Fatalf("expected empty bucket from draining, got fill of %d cpu-ns", l.Fill())
	}

	assistCPUTime := int64(0)
	doAssist := func(d time.Duration, frac float64) int64 {
		t.Helper()
		assistCPUTime += int64(frac * float64(d) * procs)
		return assistCPUTime
	}

	// Test filling up the bucket with 50% total GC work (so, not moving the bucket at all).
	l.Update(doAssist(10*time.Millisecond, 0.5-GCBackgroundUtilization), advance(10*time.Millisecond))
	if l.Fill() != 0 {
		t.Fatalf("expected empty bucket from 50%% GC work, got fill of %d cpu-ns", l.Fill())
	}

	// Test adding to the bucket overall with 100% GC work.
	l.Update(doAssist(time.Millisecond, 1.0-GCBackgroundUtilization), advance(time.Millisecond))
	if expect := uint64(procs * time.Millisecond); l.Fill() != expect {
		t.Errorf("expected %d fill from 100%% GC CPU, got fill of %d cpu-ns", expect, l.Fill())
	}
	if l.Limiting() {
		t.Errorf("limiter is enabled after filling bucket but shouldn't be")
	}
	if t.Failed() {
		t.FailNow()
	}

	// Test filling the bucket exactly full.
	l.Update(doAssist(CapacityPerProc-time.Millisecond, 1.0-GCBackgroundUtilization), advance(CapacityPerProc-time.Millisecond))
	if l.Fill() != l.Capacity() {
		t.Errorf("expected bucket filled to capacity %d, got %d", l.Capacity(), l.Fill())
	}
	if !l.Limiting() {
		t.Errorf("limiter is not enabled after filling bucket but should be")
	}
	if l.Overflow() != 0 {
		t.Errorf("bucket filled exactly should not have overflow, found %d", l.Overflow())
	}
	if t.Failed() {
		t.FailNow()
	}

	// Test adding with a delta of exactly zero. That is, GC work is exactly 50% of all resources.
	// Specifically, the limiter should still be on, and no overflow should accumulate.
	l.Update(doAssist(1*time.Second, 0.5-GCBackgroundUtilization), advance(1*time.Second))
	if l.Fill() != l.Capacity() {
		t.Errorf("expected bucket filled to capacity %d, got %d", l.Capacity(), l.Fill())
	}
	if !l.Limiting() {
		t.Errorf("limiter is not enabled after filling bucket but should be")
	}
	if l.Overflow() != 0 {
		t.Errorf("bucket filled exactly should not have overflow, found %d", l.Overflow())
	}
	if t.Failed() {
		t.FailNow()
	}

	// Drain the bucket by half.
	l.Update(doAssist(CapacityPerProc, 0), advance(CapacityPerProc))
	if expect := l.Capacity() / 2; l.Fill() != expect {
		t.Errorf("failed to drain to %d, got fill %d", expect, l.Fill())
	}
	if l.Limiting() {
		t.Errorf("limiter is enabled after draining bucket but shouldn't be")
	}
	if t.Failed() {
		t.FailNow()
	}

	// Test overfilling the bucket.
	l.Update(doAssist(CapacityPerProc, 1.0-GCBackgroundUtilization), advance(CapacityPerProc))
	if l.Fill() != l.Capacity() {
		t.Errorf("failed to fill to capacity %d, got fill %d", l.Capacity(), l.Fill())
	}
	if !l.Limiting() {
		t.Errorf("limiter is not enabled after overfill but should be")
	}
	if expect := uint64(CapacityPerProc * procs / 2); l.Overflow() != expect {
		t.Errorf("bucket overfilled should have overflow %d, found %d", expect, l.Overflow())
	}
	if t.Failed() {
		t.FailNow()
	}

	// Test ending the cycle with some assists left over.

	l.StartGCTransition(false, doAssist(1*time.Millisecond, 1.0-GCBackgroundUtilization), advance(1*time.Millisecond))
	if l.Fill() != l.Capacity() {
		t.Errorf("failed to maintain fill to capacity %d, got fill %d", l.Capacity(), l.Fill())
	}
	if !l.Limiting() {
		t.Errorf("limiter is not enabled after overfill but should be")
	}
	if expect := uint64((CapacityPerProc/2 + time.Millisecond) * procs); l.Overflow() != expect {
		t.Errorf("bucket overfilled should have overflow %d, found %d", expect, l.Overflow())
	}
	if t.Failed() {
		t.FailNow()
	}

	// Ensure any attempts to update are skipped.
	l.Update(0, advance(2*time.Millisecond))
	if l.Fill() != l.Capacity() {
		t.Errorf("failed to maintain fill to capacity %d, got fill %d", l.Capacity(), l.Fill())
	}
	if !l.Limiting() {
		t.Errorf("limiter is not enabled after overfill but should be")
	}
	if expect := uint64((CapacityPerProc/2 + time.Millisecond) * procs); l.Overflow() != expect {
		t.Errorf("bucket overfilled should have overflow %d, found %d", expect, l.Overflow())
	}
	if t.Failed() {
		t.FailNow()
	}

	// Make sure the STW adds to the bucket.
	l.FinishGCTransition(advance(3 * time.Millisecond))
	if l.Fill() != l.Capacity() {
		t.Errorf("failed to maintain fill to capacity %d, got fill %d", l.Capacity(), l.Fill())
	}
	if !l.Limiting() {
		t.Errorf("limiter is not enabled after overfill but should be")
	}
	if expect := uint64((CapacityPerProc/2 + 6*time.Millisecond) * procs); l.Overflow() != expect {
		t.Errorf("bucket overfilled should have overflow %d, found %d", expect, l.Overflow())
	}
	if t.Failed() {
		t.FailNow()
	}

	// Resize procs up and make sure limiting stops.
	expectFill := l.Capacity()
	l.ResetCapacity(advance(0), procs+10)
	if l.Fill() != expectFill {
		t.Errorf("failed to maintain fill at old capacity %d, got fill %d", expectFill, l.Fill())
	}
	if l.Limiting() {
		t.Errorf("limiter is enabled after resetting capacity higher")
	}
	if expect := uint64((CapacityPerProc/2 + 6*time.Millisecond) * procs); l.Overflow() != expect {
		t.Errorf("bucket overflow %d should have remained constant, found %d", expect, l.Overflow())
	}
	if t.Failed() {
		t.FailNow()
	}

	// Resize procs down and make sure limiting begins again.
	// Also make sure resizing doesn't affect overflow. This isn't
	// a case where we want to report overflow, because we're not
	// actively doing work to achieve it. It's that we have fewer
	// CPU resources now.
	l.ResetCapacity(advance(0), procs-10)
	if l.Fill() != l.Capacity() {
		t.Errorf("failed lower fill to new capacity %d, got fill %d", l.Capacity(), l.Fill())
	}
	if !l.Limiting() {
		t.Errorf("limiter is disabled after resetting capacity lower")
	}
	if expect := uint64((CapacityPerProc/2 + 6*time.Millisecond) * procs); l.Overflow() != expect {
		t.Errorf("bucket overflow %d should have remained constant, found %d", expect, l.Overflow())
	}
	if t.Failed() {
		t.FailNow()
	}
}
