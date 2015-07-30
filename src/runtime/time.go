// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Time-related runtime and pieces of package time.

package runtime

import (
	_base "runtime/internal/base"
	_race "runtime/internal/race"
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
	_base.AddtimerLocked(t)
	_base.Goparkunlock(&_base.Timers.Lock, "sleep", _base.TraceEvGoSleep, 2)
}

// startTimer adds t to the timer heap.
//go:linkname startTimer time.startTimer
func startTimer(t *_base.Timer) {
	if _base.Raceenabled {
		_race.Racerelease(unsafe.Pointer(t))
	}
	_base.Addtimer(t)
}

// stopTimer removes t from the timer heap if it is there.
// It returns true if t was removed, false if t wasn't even there.
//go:linkname stopTimer time.stopTimer
func stopTimer(t *_base.Timer) bool {
	return _base.Deltimer(t)
}

// Go runtime.

// Ready the goroutine arg.
func goroutineReady(arg interface{}, seq uintptr) {
	_base.Goready(arg.(*_base.G), 0)
}

// Entry points for net, time to call nanotime.

//go:linkname net_runtimeNano net.runtimeNano
func net_runtimeNano() int64 {
	return _base.Nanotime()
}

//go:linkname time_runtimeNano time.runtimeNano
func time_runtimeNano() int64 {
	return _base.Nanotime()
}
