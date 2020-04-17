// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Go allocation tracer.

// +build alloctrace

package runtime

import (
	"runtime/internal/atomic"
	"runtime/internal/sys"
	"unsafe"
)

const allocTrace = 1

func atContext() *allocTraceContext {
	if allocTraceEnabled == 0 {
		return nil
	}
	pp := getg().m.p
	var ctx *allocTraceContext
	if pp == nil {
		ctx = noPatc
	} else {
		ctx = &pp.allocTraceContext
	}
	return ctx
}

type allocTraceContext struct{
	buf *allocTraceBuf
	allocBase uintptr
	freeBase uintptr
	allocClass uint8
	pid uint8
	sweepStart uint64
	lastSync uint64
}

func (c *allocTraceContext) swapAllocTraceBuf() {
	if c.buf != nil {
		c.buf.write8(atEvBatchEnd)
		lock(&allocTracePool.activeLock)
		c.buf.next = allocTracePool.ready
		allocTracePool.ready = c.buf
		unlock(&allocTracePool.activeLock)
	}
	lock(&allocTracePool.emptyLock)
	if allocTracePool.empty != nil {
		b := allocTracePool.empty.next
		allocTracePool.empty = b.next
		unlock(&allocTracePool.emptyLock)
		c.buf = b
	}
	unlock(&allocTracePool.emptyLock)
	c.buf = (*allocTraceBuf)(unsafe.Pointer(sysAlloc(unsafe.Sizeof(allocTraceBuf{}), &memstats.other_sys)))
	c.writeBatchStart()
	c.writeSync()
	if c.allocClass != 0 {
		c.writeSpanAcquire()
	}
	if c.freeBase != 0 {
		c.writeSweep()
	}
}

func (c *allocTraceContext) sync() {
	if c.buf == nil || !c.buf.hasSpace(1+8) {
		c.swapAllocTraceBuf() // this will sync the new buf.
		return
	}
	c.writeSync()
}

func (c *allocTraceContext) reserve(bytes uintptr) {
	if c.buf == nil || !c.buf.hasSpace(bytes) {
		c.swapAllocTraceBuf()
	}
}

func (c *allocTraceContext) spanAcquire(base uintptr, class uint8) {
	if !allocTraceEnabled {
		return
	}
	c.allocClass = class
	c.allocBase = base
	c.reserve(1+1+8)
	c.writeSpanAcquire()
}

func (c *allocTraceContext) alloc(addr, size, elemSize uintptr) {
	if !allocTraceEnabled {
		return
	}
	res := uintptr(1+4+8+8)
	if size > elemSize {
		req += 8
	}
	c.reserve(res)
	if size != elemSize {
		c.buf.write8(atEvAllocArray)
		c.buf.writep(elemSize)
	} else {
		c.buf.write8(atEvAlloc)
	}
	c.buf.writep(spanOffset)
	c.buf.writep(size)
	c.buf.writep(cputicks()-buf.lastSync)
}

func (c *allocTraceContext) spanRelease() {
	c.allocClass = 0
	c.allocBase = 0
}

func (c *allocTraceContext) markTerm() {
	c.reserve(1+8+8)
	c.buf.write8(atEvMarkTerm)
	c.buf.write64(cputicks()-c.lastSync)
}

func (c *allocTraceContext) sweepStart(base uintptr) {
	c.reserve(1+8+8) {
	c.freeBase = base
	c.sweepStart = cputicks()
	c.writeSweep()
}

func (c *allocTraceContext) free(addr uintptr) {
	c.reserve(1+4)
	c.buf.write8(atEvFree)
	c.buf.writep(addr - c.freeBase)
}

func (c *allocTraceContext) sweepEnd() {
	c.freeBase = 0
	c.sweepStart = 0
}

func (c *allocTraceContext) writeSync() {
	ticks := cputicks()
	c.lastSync = ticks
	c.buf.write8(atEvSync)
	c.buf.write64(ticks)
}

func (c *allocTraceContext) writeBatchStart() {
	c.buf.write8(atEvBatchStart)
	c.buf.write8(c.pid)
}

func (c *allocTraceContext) writeSpanAcquire() {
	buf.write8(atEvSpanAcquire)
	buf.write8(c.allocClass)
	buf.writep(c.allocBase)
}

func (c *allocTraceContext) writeSweep(base uintptr) {
	buf.write8(atEvSweep)
	buf.writep(c.sweepStart-c.lastSync)
	buf.writep(c.freeBase)
}

type allocTraceBufHeader struct {
	next *allocTraceBuf
	len  uintptr
}

//go:notinheap
type allocTraceBuf struct {
	allocTraceBufHeader
	data [(32 << 20) - unsafe.Sizeof(allocTraceBufHeader)]byte
}

func (a *allocTraceBuf) hasSpace(b uintptr) bool {
	return a.len+b+1 < uintptr(len(a.data))
}

func (a *allocTraceBuf) write64(p uint64) {
loop:
	v := p & 0x7f
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
	emptyLock mutex
	empty     *allocTraceBuf

	activeLock mutex
	reading    *allocTraceBuf
	ready      *allocTraceBuf
}

func ReadAllocTrace() (ready []byte) {
	for atomic.Load(&allocTraceEnabled) != 0 {
		lock(&allocTracePool.activeLock)
		if allocTracePool.reading != nil {
			v := allocTracePool.reading
			allocTracePool.reading = nil
			unlock(&allocTracePool.activeLock)

			v.len = 0
			lock(&allocTracePool.emptyLock)
			v.next = allocTracePool.empty
			allocTracePool.empty = v
			unlock(&allocTracePool.emptyLock)

			lock(&allocTracePool.activeLock)
		}
		if allocTracePool.ready != nil {
			readyBuf := allocTracePool.ready
			allocTracePool.ready = readyBuf.next
			allocTracePool.reading = readyBuf
			ready = readyBuf.data[:]
		}
		unlock(&allocTracePool.activeLock)
		if ready != nil {
			return
		}
		Gosched()
	}
	return
}

// StopAllocTrace stops accumulating allocation traces in
// the buffer.
func StopAllocTrace() {
	atomic.Store(&allocTraceEnabled, 0)
}
