// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Go allocation tracer.

package runtime

const (
	atEvBad uint8 = iota
	atEvSpanAcquire
	atEvAlloc
	atEvAllocArray
	atEvSweep
	atEvFree
	atEvMarkTerm
	atEvSync
	atEvBatchStart
	atEvBatchEnd
)

var noPatc = allocTraceContext{}
