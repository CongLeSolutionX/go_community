// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Go allocation tracer.

// +build alloctrace

package runtime

import (
	"runtime/internal/atomic"
	"unsafe"
)

const allocTrace = 1

type allocTraceContext struct {
	buf             *allocTraceBuf
	allocBase       [numSpanClasses]uintptr
	freeBase        uintptr
	pid             uint8
	sweepStartTicks uint64
	lastSync        uint64
}

func (c *allocTraceContext) swapAllocTraceBuf(clear bool) {
	if c.buf != nil {
		c.buf.write8(atEvBatchEnd)
		allocTracePool.ready.push(&c.buf.lfnode)
		c.buf = nil
		if atomic.Cas(&allocTracePool.readerSleeping, 1, 0) {
			notewakeup(&allocTracePool.readerSleep)
		}
	}
	if clear {
		return
	}
	c.buf = (*allocTraceBuf)(allocTracePool.empty.pop())
	if c.buf == nil {
		c.buf = (*allocTraceBuf)(unsafe.Pointer(sysAlloc(unsafe.Sizeof(allocTraceBuf{}), &memstats.other_sys)))
	}
	c.writeBatchStart()
	c.writeSync()
}

func (c *allocTraceContext) init(id int32) {
	c.pid = uint8(id)
}

func (c *allocTraceContext) sync() {
	if allocTraceEnabled == 0 {
		return
	}
	if c.buf == nil || !c.buf.hasSpace(1+8) {
		c.swapAllocTraceBuf(false) // this will sync the new buf.
		return
	}
	c.writeSync()
}

func (c *allocTraceContext) reserve(bytes uintptr) {
	if c.buf == nil || !c.buf.hasSpace(bytes) {
		c.swapAllocTraceBuf(false)
	}
}

func (c *allocTraceContext) spanAcquire(base uintptr, class uint8) {
	if allocTraceEnabled == 0 {
		return
	}
	c.allocBase[class] = base
	c.reserve(1 + 1 + 8)
	c.buf.write8(atEvSpanAcquire)
	c.buf.write8(class)
	c.buf.writep(base)
}

func (c *allocTraceContext) allocArray(addr, size, elemSize uintptr, class uint8) {
	if allocTraceEnabled == 0 {
		return
	}
	if class >= 2 && c.allocBase[class] == 0 {
		print("runtime: class = ", class, ", pid = ", int32(c.pid)-1, "\n")
		throw("no span acquired")
	}
	c.reserve(1 + 1 + 8 + 8 + 8 + 8)
	c.buf.write8(atEvAllocArray)
	c.buf.writep(elemSize)
	c.buf.write8(class)
	c.buf.writep(addr - c.allocBase[class])
	c.buf.writep(size)
	t := uint64(cputicks())
	if c.lastSync > t {
		throw("time went backwards in alloc")
	}
	c.buf.write64(t - c.lastSync)
}

func (c *allocTraceContext) alloc(addr, size uintptr, class uint8) {
	if allocTraceEnabled == 0 {
		return
	}
	if class >= 2 && c.allocBase[class] == 0 {
		print("runtime: class = ", class, ", pid = ", int32(c.pid)-1, "\n")
		throw("no span acquired")
	}
	c.reserve(1 + 1 + 8 + 8 + 8)
	c.buf.write8(atEvAlloc)
	c.buf.write8(class)
	c.buf.writep(addr - c.allocBase[class])
	c.buf.writep(size)
	t := uint64(cputicks())
	if c.lastSync > t {
		throw("time went backwards in alloc")
	}
	c.buf.write64(t - c.lastSync)
}

func (c *allocTraceContext) spanRelease(base uintptr, class uint8) {
	if allocTraceEnabled == 0 {
		return
	}
	if c.allocBase[class] != base {
		print("runtime: class = ", class, ", base = ", hex(base), ", pid = ", int32(c.pid)-1, "\n")
		print("runtime: allocBase[class] = ", c.allocBase[class], "\n")
		throw("released unacquired (?) span")
	}
	c.allocBase[class] = 0
	c.reserve(1 + 1)
	c.buf.write8(atEvSpanRelease)
	c.buf.write8(class)
}

func (c *allocTraceContext) markTerm() {
	if allocTraceEnabled == 0 {
		return
	}
	c.reserve(1 + 8)
	c.buf.write8(atEvMarkTerm)
	t := uint64(cputicks())
	if c.lastSync > t {
		throw("time went backwards in mark term")
	}
	c.buf.write64(t - c.lastSync)
}

func (c *allocTraceContext) sweepStart(base uintptr) {
	if allocTraceEnabled == 0 {
		return
	}
	c.reserve(1 + 8 + 8)
	c.freeBase = base
	c.sweepStartTicks = uint64(cputicks())
	c.buf.write8(atEvSweep)
	if c.lastSync > c.sweepStartTicks {
		throw("time went backwards in sweep start")
	}
	c.buf.write64(c.sweepStartTicks - c.lastSync)
	c.buf.writep(base)
}

func (c *allocTraceContext) free(addr uintptr) {
	if allocTraceEnabled == 0 {
		return
	}
	c.reserve(1 + 4)
	c.buf.write8(atEvFree)
	c.buf.writep(addr - c.freeBase)
}

func (c *allocTraceContext) sweepEnd() {
	if allocTraceEnabled == 0 {
		return
	}
	c.freeBase = 0
	c.sweepStartTicks = 0
}

func (c *allocTraceContext) writeSync() {
	ticks := uint64(cputicks())
	c.lastSync = ticks
	c.buf.write8(atEvSync)
	c.buf.write64(ticks)
}

func (c *allocTraceContext) writeBatchStart() {
	c.buf.write8(atEvBatchStart)
	c.buf.write8(c.pid)
}

type allocTraceBufHeader struct {
	lfnode
	len uintptr
}

//go:notinheap
type allocTraceBuf struct {
	allocTraceBufHeader
	data [AllocTraceBatchSize - unsafe.Sizeof(allocTraceBufHeader{})]byte
}

func (a *allocTraceBuf) hasSpace(b uintptr) bool {
	return a.len+b+1 < uintptr(len(a.data))
}

func (a *allocTraceBuf) writep(p uintptr) {
	a.write64(uint64(p))
}

func (a *allocTraceBuf) write64(p uint64) {
loop:
	v := uint8(p & 0x7f)
	p >>= 7
	if p == 0 {
		a.write8(v)
		return
	}
	a.write8((1 << 7) | v)
	goto loop
}

func (a *allocTraceBuf) write8(b uint8) {
	a.data[a.len] = b
	a.len++
}

var allocTraceEnabled uint32 = 1

var allocTracePool struct {
	// Lock-free stack of allocTraceBufs.
	empty lfstack
	ready lfstack

	// Current allocTraceBuf being read.
	// Accessed without atomics, but only
	// ever has one reader/writer at a time.
	reading *allocTraceBuf

	// State for the reader to sleep.
	readerSleeping uint32
	readerSleep    note
}

// ReadAllocTrace returns a byte slice containing a batch
// of allocation trace events in a binary encoding.
//
// The returned byte slice must be read before calling
// this function again, and the returned slice may not be
// held onto.
//
// If this function returns nil, then there are no more
// events to read.
//
// This function is not safe to call concurrently.
//
// In order to produce a correct encoding, the returned
// byte slice must be padded up to AllocTraceBatchSize bytes
// in any output.
func ReadAllocTrace() (bytes []byte) {
	// Recycle the buffer we just read.
	if allocTracePool.reading != nil {
		allocTracePool.reading.len = 0
		allocTracePool.empty.push(&allocTracePool.reading.lfnode)
		allocTracePool.reading = nil
	}

	// Try to get a new one.
	ready := (*allocTraceBuf)(allocTracePool.ready.pop())
	if ready == nil && allocTraceEnabled != 0 {
		noteclear(&allocTracePool.readerSleep)
		atomic.Store(&allocTracePool.readerSleeping, 1)
		notetsleepg(&allocTracePool.readerSleep, -1)
		ready = (*allocTraceBuf)(allocTracePool.ready.pop())
	}
	if ready != nil {
		allocTracePool.reading = ready
		bytes = ready.data[:]
	}
	return
}

// StopAllocTrace stops accumulating allocation traces and flushes
// any in-progress allocation traces.
func StopAllocTrace() {
	stopTheWorldGC("stop allocation tracing")
	allocTraceEnabled = 0
	for _, p := range allp[:cap(allp)] {
		p.mcache.atState.swapAllocTraceBuf(true)
	}
	startTheWorldGC()
}

/*
func VerifyAllocTrace() {
	stopTheWorldGC("allocation tracing verification")
	for _, p := range allp[:cap(allp)] {
		if p.mcache.atState.buf != nil {
			throw("reacquired trace bufs")
		}
	}
	ready := (*allocTraceBuf)(allocTracePool.ready.pop())
	if ready != nil {
		throw("lingering trace bufs to be read")
	}
	startTheWorldGC()
}
*/
