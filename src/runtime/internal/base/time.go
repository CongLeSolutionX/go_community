// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Time-related runtime and pieces of package time.

package base

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
	Period int64
	F      func(interface{}, uintptr)
	Arg    interface{}
	Seq    uintptr
}

var Timers struct {
	Lock         Mutex
	Gp           *G
	Created      bool
	Sleeping     bool
	Rescheduling bool
	Waitnote     Note
	T            []*Timer
}

// nacl fake time support - time in nanoseconds since 1970
var Faketime int64

func timejump() *G {
	if Faketime == 0 {
		return nil
	}

	Lock(&Timers.Lock)
	if !Timers.Created || len(Timers.T) == 0 {
		Unlock(&Timers.Lock)
		return nil
	}

	var gp *G
	if Faketime < Timers.T[0].When {
		Faketime = Timers.T[0].When
		if Timers.Rescheduling {
			Timers.Rescheduling = false
			gp = Timers.Gp
		}
	}
	Unlock(&Timers.Lock)
	return gp
}
