// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Utilities for printing time-stamped traces of steps in an
// algorithm.

package runtime

import "runtime/internal/atomic"

// An elapsed tracks the elapsed time between a sequence of events.
type elapsed struct {
	i     uint32
	times [32]struct {
		ts    int64
		p     int32
		label string
	}
}

const elapsedEnabled = true

func (e *elapsed) reset() {
	atomic.Store(&e.i, 0)
}

// tic adds an event with label "label" to e's log with the time since
// the previous event. This is safe to use concurrently.
//
//go:nosplit
//go:yeswritebarrierrec
func (e *elapsed) tic(label string) {
	if !elapsedEnabled || e == nil {
		return
	}

	i := int(atomic.Xadd(&e.i, 1)) - 1
	if i >= len(e.times) {
		return
	}
	e.times[i].ts = nanotime()
	e.times[i].p = 0
	if g := getg(); g != nil {
		if g.m != nil && g.m.p != 0 {
			e.times[i].p = g.m.p.ptr().id
		}
	}
	e.times[i].label = label
}

func (e *elapsed) nanos() int64 {
	if e == nil {
		return 0
	}
	i := int(e.i)
	if i == 0 {
		return 0
	}
	if i > len(e.times) {
		i = len(e.times)
	}
	return e.times[i-1].ts - e.times[0].ts
}

// print prints e's log and resets e.
func (e *elapsed) print() {
	if !elapsedEnabled || e == nil {
		return
	}

	i := int(e.i)
	if i > len(e.times) {
		i = len(e.times)
	}
	if i == 0 {
		return
	}
	printlock()
	last := e.times[0].ts
	for _, time := range e.times[:i] {
		print("+", time.ts-last, "ns (p ", time.p, ") ", time.label, "\n")
		last = time.ts
	}
	printunlock()
	e.reset()
}
