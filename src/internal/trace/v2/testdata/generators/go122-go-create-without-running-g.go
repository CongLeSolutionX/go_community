// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Regression test for an issue found in development.
//
// The core of the issue is that if generation counters
// aren't considered as part of sequence numbers, then
// it's possible to accidentally advance without a
// GoStatus event.
//
// The situation is one in which it just so happens that
// an event on the frontier for a following generation
// has a sequence number exactly one higher than the last
// sequence number for e.g. a goroutine in the previous
// generation. The parser should wait to find a GoStatus
// event before advancing into the next generation at all.
// It turns out this situation is pretty rare; the GoStatus
// event almost always shows up first in practice. But it
// can and did happen.

package main

import (
	"internal/trace/v2"
	"internal/trace/v2/event/go122"
	testgen "internal/trace/v2/internal/testgen/go122"
)

func main() {
	testgen.Main(gen)
}

func gen(t *testgen.Trace) {
	g1 := t.Generation(1)

	// A goroutine gets created on a running P, then starts running.
	b0 := g1.Batch(trace.ThreadID(0), 0)
	b0.Event("ProcStatus", trace.ProcID(0), go122.ProcRunning)
	b0.Event("GoCreate", trace.GoID(5), testgen.NoStack, testgen.NoStack)
	b0.Event("GoStart", trace.GoID(5), testgen.Seq(1))
	b0.Event("GoStop", "whatever", testgen.NoStack)
}
