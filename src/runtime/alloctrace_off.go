// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Go allocation tracer.

// +build !alloctrace

package runtime

const allocTrace = 0

func atContext() *allocTraceContext {
	return nil
}

type allocTraceContext struct {
}

func (c *allocTraceContext) sync() {
}

func (c *allocTraceContext) spanAcquire(base uintptr, class uint8) {
}

func (c *allocTraceContext) alloc(addr, size, elemSize uintptr) {
}

func (c *allocTraceContext) spanRelease() {
}

func (c *allocTraceContext) markTerm() {
}

func (c *allocTraceContext) sweepStart(base uintptr) {
}

func (c *allocTraceContext) free(addr uintptr) {
}

func (c *allocTraceContext) sweepEnd() {
}

func ReadAllocTrace() (ready []byte) {
	return
}

func StopAllocTrace() {
}
