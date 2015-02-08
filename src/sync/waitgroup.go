// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sync

import (
	"sync/atomic"
	"unsafe"
)

// A WaitGroup waits for a collection of goroutines to finish.
// The main goroutine calls Add to set the number of
// goroutines to wait for.  Then each of the goroutines
// runs and calls Done when finished.  At the same time,
// Wait can be used to block until all goroutines have finished.
type WaitGroup struct {
	state uint64 // high 32 bits is counter, low 32 bits is waiter count
	sema  uint32
}

// Add adds delta, which may be negative, to the WaitGroup counter.
// If the counter becomes zero, all goroutines blocked on Wait are released.
// If the counter goes negative, Add panics.
//
// Note that calls with a positive delta that occur when the counter is zero
// must happen before a Wait. Calls with a negative delta, or calls with a
// positive delta that start when the counter is greater than zero, may happen
// at any time.
// Typically this means the calls to Add should execute before the statement
// creating the goroutine or other event to be waited for.
// If a WaitGroup is reused to wait for several independent sets of events,
// new Add calls must happen after all previous Wait calls has returned.
// See the WaitGroup example.
func (wg *WaitGroup) Add(delta int) {
	if raceenabled {
		_ = wg.state // trigger nil deref early
		if delta < 0 {
			// Synchronize decrements with Wait.
			raceReleaseMerge(unsafe.Pointer(wg))
		}
		raceDisable()
		defer raceEnable()
	}
	state := atomic.AddUint64(&wg.state, uint64(delta)<<32)
	v := int32(state >> 32)
	w := uint32(state)
	if raceenabled {
		if delta > 0 && v == int32(delta) {
			// The first increment must be synchronized with Wait.
			// Need to model this as a read, because there can be
			// several concurrent wg.counter transitions from 0.
			raceRead(unsafe.Pointer(&wg.sema))
		}
	}
	if v < 0 {
		panic("sync: negative WaitGroup counter")
	}
	if v > 0 || w == 0 {
		return
	}
	// This goroutine set counter to 0 when waiters > 0.
	// Now there can't be concurrent mutations of state:
	// - Add's must not happen concurrently with Wait,
	// - Wait does not increment waiters if it sees counter == 0.
	// Still do a cheap sanity check to detect WaitGroup misuse.
	if wg.state != state {
		panic("sync: WaitGroup misuse: Add races with Wait")
	}
	wg.state = 0
	for ; w != 0; w-- {
		runtime_Semrelease(&wg.sema)
	}
}

// Done decrements the WaitGroup counter.
func (wg *WaitGroup) Done() {
	wg.Add(-1)
}

// Wait blocks until the WaitGroup counter is zero.
func (wg *WaitGroup) Wait() {
	if raceenabled {
		_ = wg.state // trigger nil deref early
		raceDisable()
	}
	for {
		state := atomic.LoadUint64(&wg.state)
		v := int32(state >> 32)
		w := uint32(state)
		if v == 0 {
			// Counter is 0, no need to wait.
			if raceenabled {
				raceEnable()
				raceAcquire(unsafe.Pointer(wg))
			}
			return
		}
		// Increment waiters count.
		if atomic.CompareAndSwapUint64(&wg.state, state, state+1) {
			if raceenabled && w == 0 {
				// Wait must be synchronized with the first Add.
				// Need to model this is as a write to race with the read in Add.
				// As a consequence, can do the write only for the first waiter,
				// otherwise concurrent Waits will race with each other.
				raceWrite(unsafe.Pointer(&wg.sema))
			}
			runtime_Semacquire(&wg.sema)
			if raceenabled {
				raceEnable()
				raceAcquire(unsafe.Pointer(wg))
			}
			return
		}
	}
}
