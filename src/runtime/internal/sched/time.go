// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Time-related runtime and pieces of package time.

package sched

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	"unsafe"
)

// Package time knows the layout of this structure.
// If this struct changes, adjust ../time/sleep.go:/runtimeTimer.
// For GOOS=nacl, package syscall knows the layout of this structure.
// If this struct changes, adjust ../syscall/net_nacl.go:/runtimeTimer.
type Timer struct {
	I int // heap index

	// Timer wakes up at when, and then at when+period, ... (period > 0 only)
	// each time calling f(now, arg) in the timer goroutine, so f must be
	// a well-behaved function and not block.
	When   int64
	period int64
	F      func(interface{}, uintptr)
	Arg    interface{}
	Seq    uintptr
}

var Timers struct {
	Lock         _core.Mutex
	gp           *_core.G
	created      bool
	sleeping     bool
	rescheduling bool
	waitnote     _core.Note
	T            []*Timer
}

// nacl fake time support - time in nanoseconds since 1970
var faketime int64

func Addtimer(t *Timer) {
	_lock.Lock(&Timers.Lock)
	AddtimerLocked(t)
	_lock.Unlock(&Timers.Lock)
}

// Add a timer to the heap and start or kick the timer proc.
// If the new timer is earlier than any of the others.
// Timers are locked.
func AddtimerLocked(t *Timer) {
	// when must never be negative; otherwise timerproc will overflow
	// during its delta calculation and never expire other runtime·timers.
	if t.When < 0 {
		t.When = 1<<63 - 1
	}
	t.I = len(Timers.T)
	Timers.T = append(Timers.T, t)
	SiftupTimer(t.I)
	if t.I == 0 {
		// siftup moved to top: new earliest deadline.
		if Timers.sleeping {
			Timers.sleeping = false
			Notewakeup(&Timers.waitnote)
		}
		if Timers.rescheduling {
			Timers.rescheduling = false
			Goready(Timers.gp)
		}
	}
	if !Timers.created {
		Timers.created = true
		go timerproc()
	}
}

// Timerproc runs the time-driven events.
// It sleeps until the next event in the timers heap.
// If addtimer inserts a new earlier event, addtimer1 wakes timerproc early.
func timerproc() {
	Timers.gp = _core.Getg()
	Timers.gp.Issystem = true
	for {
		_lock.Lock(&Timers.Lock)
		Timers.sleeping = false
		now := _lock.Nanotime()
		delta := int64(-1)
		for {
			if len(Timers.T) == 0 {
				delta = -1
				break
			}
			t := Timers.T[0]
			delta = t.When - now
			if delta > 0 {
				break
			}
			if t.period > 0 {
				// leave in heap but adjust next time to fire
				t.When += t.period * (1 + -delta/t.period)
				SiftdownTimer(0)
			} else {
				// remove from heap
				last := len(Timers.T) - 1
				if last > 0 {
					Timers.T[0] = Timers.T[last]
					Timers.T[0].I = 0
				}
				Timers.T[last] = nil
				Timers.T = Timers.T[:last]
				if last > 0 {
					SiftdownTimer(0)
				}
				t.I = -1 // mark as removed
			}
			f := t.F
			arg := t.Arg
			seq := t.Seq
			_lock.Unlock(&Timers.Lock)
			if Raceenabled {
				Raceacquire(unsafe.Pointer(t))
			}
			f(arg, seq)
			_lock.Lock(&Timers.Lock)
		}
		if delta < 0 || faketime > 0 {
			// No timers left - put goroutine to sleep.
			Timers.rescheduling = true
			Goparkunlock(&Timers.Lock, "timer goroutine (idle)")
			continue
		}
		// At least one timer pending.  Sleep until then.
		Timers.sleeping = true
		Noteclear(&Timers.waitnote)
		_lock.Unlock(&Timers.Lock)
		Notetsleepg(&Timers.waitnote, delta)
	}
}

func timejump() *_core.G {
	if faketime == 0 {
		return nil
	}

	_lock.Lock(&Timers.Lock)
	if !Timers.created || len(Timers.T) == 0 {
		_lock.Unlock(&Timers.Lock)
		return nil
	}

	var gp *_core.G
	if faketime < Timers.T[0].When {
		faketime = Timers.T[0].When
		if Timers.rescheduling {
			Timers.rescheduling = false
			gp = Timers.gp
		}
	}
	_lock.Unlock(&Timers.Lock)
	return gp
}

// Heap maintenance algorithms.

func SiftupTimer(i int) {
	t := Timers.T
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
	t := Timers.T
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
