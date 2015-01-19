// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sched

// An MSpan is a run of pages.
const (
	XMSpanInUse = iota // allocated for garbage collected heap
	MSpanStack         // allocated for use by stack allocator
	MSpanFree
	MSpanListHead
	MSpanDead
)

const (
	// flags to malloc
	XFlagNoScan = 1 << 0 // GC doesn't have to scan object
	XFlagNoZero = 1 << 1 // don't zero memory
)
