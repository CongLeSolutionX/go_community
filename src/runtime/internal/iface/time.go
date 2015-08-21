// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Time-related runtime and pieces of package time.

package iface

import (
	_base "runtime/internal/base"
	_gc "runtime/internal/gc"
	"unsafe"
)

// Package time APIs.
// Godoc uses the comments in package time, not these.

// time.now is implemented in assembly.

// timeSleep puts the current goroutine to sleep for at least ns nanoseconds.
//go:linkname timeSleep time.Sleep
func timeSleep(ns int64) {
	if ns <= 0 {
		return
	}

	t := new(_base.Timer)
	t.When = _base.Nanotime() + ns
	t.F = goroutineReady
	t.Arg = _base.Getg()
	_base.Lock(&_base.Timers.Lock)
	AddtimerLocked(t)
	_base.Goparkunlock(&_base.Timers.Lock, "sleep", _base.TraceEvGoSleep, 2)
}

// Go runtime.

// Ready the goroutine arg.
func goroutineReady(arg interface{}, seq uintptr) {
	_gc.Goready(arg.(*_base.G), 0)
}

// Add a timer to the heap and start or kick the timer proc.
// If the new timer is earlier than any of the others.
// Timers are locked.
func AddtimerLocked(t *_base.Timer) {
	// when must never be negative; otherwise timerproc will overflow
	// during its delta calculation and never expire other runtime·timers.
	if t.When < 0 {
		t.When = 1<<63 - 1
	}
	t.I = len(_base.Timers.T)
	_base.Timers.T = append(_base.Timers.T, t)
	SiftupTimer(t.I)
	if t.I == 0 {
		// siftup moved to top: new earliest deadline.
		if _base.Timers.Sleeping {
			_base.Timers.Sleeping = false
			_base.Notewakeup(&_base.Timers.Waitnote)
		}
		if _base.Timers.Rescheduling {
			_base.Timers.Rescheduling = false
			_gc.Goready(_base.Timers.Gp, 0)
		}
	}
	if !_base.Timers.Created {
		_base.Timers.Created = true
		go Timerproc()
	}
}

// Timerproc runs the time-driven events.
// It sleeps until the next event in the timers heap.
// If addtimer inserts a new earlier event, addtimer1 wakes timerproc early.
func Timerproc() {
	_base.Timers.Gp = _base.Getg()
	for {
		_base.Lock(&_base.Timers.Lock)
		_base.Timers.Sleeping = false
		now := _base.Nanotime()
		delta := int64(-1)
		for {
			if len(_base.Timers.T) == 0 {
				delta = -1
				break
			}
			t := _base.Timers.T[0]
			delta = t.When - now
			if delta > 0 {
				break
			}
			if t.Period > 0 {
				// leave in heap but adjust next time to fire
				t.When += t.Period * (1 + -delta/t.Period)
				SiftdownTimer(0)
			} else {
				// remove from heap
				last := len(_base.Timers.T) - 1
				if last > 0 {
					_base.Timers.T[0] = _base.Timers.T[last]
					_base.Timers.T[0].I = 0
				}
				_base.Timers.T[last] = nil
				_base.Timers.T = _base.Timers.T[:last]
				if last > 0 {
					SiftdownTimer(0)
				}
				t.I = -1 // mark as removed
			}
			f := t.F
			arg := t.Arg
			seq := t.Seq
			_base.Unlock(&_base.Timers.Lock)
			if _base.Raceenabled {
				Raceacquire(unsafe.Pointer(t))
			}
			f(arg, seq)
			_base.Lock(&_base.Timers.Lock)
		}
		if delta < 0 || _base.Faketime > 0 {
			// No timers left - put goroutine to sleep.
			_base.Timers.Rescheduling = true
			_base.Goparkunlock(&_base.Timers.Lock, "timer goroutine (idle)", _base.TraceEvGoBlock, 1)
			continue
		}
		// At least one timer pending.  Sleep until then.
		_base.Timers.Sleeping = true
		_base.Noteclear(&_base.Timers.Waitnote)
		_base.Unlock(&_base.Timers.Lock)
		_base.Notetsleepg(&_base.Timers.Waitnote, delta)
	}
}

// Heap maintenance algorithms.

func SiftupTimer(i int) {
	t := _base.Timers.T
	when := t[i].When
	tmp := t[i]
	for i > 0 {
		p := (i - 1) / 4 // parent
		if when >= t[p].When {
			break
		}
		t[i] = t[p]
		t[i].I = i
		t[p] = tmp
		t[p].I = p
		i = p
	}
}

func SiftdownTimer(i int) {
	t := _base.Timers.T
	n := len(t)
	when := t[i].When
	tmp := t[i]
	for {
		c := i*4 + 1 // left child
		c3 := c + 2  // mid child
		if c >= n {
			break
		}
		w := t[c].When
		if c+1 < n && t[c+1].When < w {
			w = t[c+1].When
			c++
		}
		if c3 < n {
			w3 := t[c3].When
			if c3+1 < n && t[c3+1].When < w3 {
				w3 = t[c3+1].When
				c3++
			}
			if w3 < w {
				w = w3
				c = c3
			}
		}
		if w >= when {
			break
		}
		t[i] = t[c]
		t[i].I = i
		t[c] = tmp
		t[c].I = c
		i = c
	}
}
