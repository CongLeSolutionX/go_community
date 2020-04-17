// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Go allocation tracer.

package runtime

import (
	"runtime/internal/atomic"
	"unsafe"
)

const (
	atEvBad uint8 = iota
	atEvSpanAcquire
	atEvAlloc
	atEvAllocArray
	atEvAllocLarge
	atEvAllocLargeNoscan
	atEvAllocLargeArray
	atEvAllocLargeArrayNoscan
	atEvSpanRelease
	atEvSweep
	atEvFree
	atEvMarkTerm
	atEvSync
	atEvBatchStart
	atEvBatchEnd
)

var allocTraceEnabled = allocTraceInit

const AllocTraceBatchSize = 32 << 10

type allocTraceContext struct {
	buf       *allocTraceBuf
	allocBase [numSpanClasses]uintptr
	freeBase  uintptr
	pid       int32
	lastSync  uint64
}

func (c *allocTraceContext) swapAllocTraceBuf(clear bool, now uint64) {
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
	if now == 0 {
		now = uint64(cputicks())
	}
	c.writeSync(now)
}

func (c *allocTraceContext) init(id int32) {
	c.pid = id
}

func (c *allocTraceContext) sync(now uint64) {
	if c.buf == nil || !c.buf.hasSpace(1+8) {
		c.swapAllocTraceBuf(false, now) // this will sync the new buf.
		return
	}
	c.writeSync(now)
}

func (c *allocTraceContext) reserve(bytes uintptr, now uint64) {
	if now != 0 && now >= c.lastSync+(1<<28) {
		// Don't let the sync record exceed 4 bytes (~100 ms diff).
		c.sync(now)
	}
	if c.buf == nil || !c.buf.hasSpace(bytes) {
		c.swapAllocTraceBuf(false, now)
	}
}

func (c *allocTraceContext) spanAcquire(base uintptr, class uint8) {
	c.allocBase[class] = base
	c.reserve(1+1+8, 0)
	c.buf.write8(atEvSpanAcquire)
	c.buf.write8(class)
	c.buf.writep(base)
}

func (c *allocTraceContext) allocLarge(addr, size uintptr, noscan, array bool) {
	now := uint64(cputicks())
	if c.lastSync > now {
		throw("time went backwards in alloc")
	}
	c.reserve(1+8+8+4, now)
	ev := atEvAllocLarge
	if array {
		ev = atEvAllocLargeArray
	} else if noscan {
		ev = atEvAllocLargeNoscan
	} else if array && noscan {
		ev = atEvAllocLargeArrayNoscan
	}
	c.buf.write8(ev)
	c.buf.writep(addr)
	c.buf.writep(size)
	c.buf.write64(now - c.lastSync)
}

func (c *allocTraceContext) allocSmall(addr, dataSize, size uintptr, class uint8, array bool) {
	if class < 2 {
		println(dataSize)
		throw("allocSmall called for large span")
	}
	if c.allocBase[class] == 0 {
		print("runtime: class = ", class, ", pid = ", int32(c.pid)-1, "\n")
		throw("no span acquired")
	}
	now := uint64(cputicks())
	if c.lastSync > now {
		throw("time went backwards in alloc")
	}
	c.reserve(1+1+3+2+4, now)
	ev := atEvAlloc
	if array {
		ev = atEvAllocArray
	}
	c.buf.write8(ev)
	c.buf.write8(class)
	c.buf.writep(addr - c.allocBase[class])
	c.buf.writep(size - dataSize)
	c.buf.write64(now - c.lastSync)
}

func (c *allocTraceContext) spanRelease(base uintptr, class uint8) {
	if c.allocBase[class] != base {
		print("runtime: class = ", class, ", base = ", hex(base), ", pid = ", int32(c.pid)-1, "\n")
		print("runtime: allocBase[class] = ", c.allocBase[class], "\n")
		throw("released unacquired (?) span")
	}
	c.allocBase[class] = 0
	c.reserve(1+1, 0)
	c.buf.write8(atEvSpanRelease)
	c.buf.write8(class)
}

func (c *allocTraceContext) markTerm() {
	now := uint64(cputicks())
	if c.lastSync > now {
		throw("time went backwards in mark term")
	}
	c.reserve(1+4, now)
	c.buf.write8(atEvMarkTerm)
	c.buf.write64(now - c.lastSync)
}

func (c *allocTraceContext) sweepStart(base uintptr) {
	now := uint64(cputicks())
	if c.lastSync > now {
		throw("time went backwards in sweep start")
	}
	c.reserve(1+4+8, now)
	c.freeBase = base
	c.buf.write8(atEvSweep)
	c.buf.write64(now - c.lastSync)
	c.buf.writep(base)
}

func (c *allocTraceContext) free(addr uintptr) {
	c.reserve(1+4, 0)
	c.buf.write8(atEvFree)
	c.buf.writep(addr - c.freeBase)
}

func (c *allocTraceContext) sweepEnd() {
	c.freeBase = 0
}

func (c *allocTraceContext) writeSync(now uint64) {
	c.lastSync = now
	c.buf.write8(atEvSync)
	c.buf.write64(now)
}

func (c *allocTraceContext) writeBatchStart() {
	c.buf.write8(atEvBatchStart)
	c.buf.write64(uint64(c.pid))
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
	return a.len+b < uintptr(len(a.data)-1)
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

// readAllocTrace returns a byte slice containing a batch
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
func readAllocTrace() (bytes []byte) {
	// Recycle the buffer we just read.
	if allocTracePool.reading != nil {
		allocTracePool.reading.len = 0
		allocTracePool.empty.push(&allocTracePool.reading.lfnode)
		allocTracePool.reading = nil
	}

	// Try to get a new one.
	ready := (*allocTraceBuf)(allocTracePool.ready.pop())
	if ready == nil && allocTraceEnabled {
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

// startAllocTrace begins accumulating allocation traces.
func startAllocTrace() {
	stopTheWorldGC("start allocation tracing")
	allocTraceEnabled = true
	startTheWorldGC()
}

// stopAllocTrace stops accumulating allocation traces and flushes
// any in-progress allocation traces.
func stopAllocTrace() {
	stopTheWorldGC("stop allocation tracing")
	allocTraceEnabled = false
	for _, p := range allp[:cap(allp)] {
		p.mcache.atState.swapAllocTraceBuf(true, 0)
	}
	startTheWorldGC()
}
