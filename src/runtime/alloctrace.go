// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Go allocation tracer.

package runtime

import (
	"runtime/internal/atomic"
	"runtime/internal/sys"
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
	atEvSweepTerm
	atEvMarkTerm
	atEvSync
	atEvBatchStart
	atEvBatchEnd
	atEvStackAlloc
	atEvStackFree
	atEvAllocPC
	atEvAllocArrayPC
)

var (
	allocTraceEnabled = false

	globalATStateLock mutex
	globalATState     = allocTraceContext{
		pid: 0, // P = -1, or no P at all.
	}
)

const AllocTraceBatchSize = 32 << 10

type allocTraceContext struct {
	buf       *allocTraceBuf
	allocBase [numSpanClasses]uintptr
	freeBase  uintptr
	pid       uint32
	lastSync  uint64
}

//go:systemstack
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
	} else {
		atomic.Xadd64(&allocTracePool.emptySize, -int64(unsafe.Sizeof(allocTraceBuf{})))
	}
	c.writeBatchStart()
	if now == 0 {
		now = uint64(cputicks())
	}
	c.writeSync(now)
}

//go:nosplit
func (c *allocTraceContext) init(id int32) {
	c.pid = uint32(id + 1)
}

//go:nosplit
func (c *allocTraceContext) sync(now uint64) {
	if c.buf == nil || !c.buf.hasSpace(1+8) {
		systemstack(func() {
			c.swapAllocTraceBuf(false, now) // this will sync the new buf.
		})
		return
	}
	c.writeSync(now)
}

//go:nosplit
func (c *allocTraceContext) reserve(bytes uintptr, now uint64) {
	if now != 0 && now >= c.lastSync+(1<<28) {
		// Don't let the sync record exceed 4 bytes (~100 ms diff).
		c.sync(now)
	}
	if c.buf == nil || !c.buf.hasSpace(bytes) {
		systemstack(func() {
			c.swapAllocTraceBuf(false, now)
		})
	}
}

//go:nosplit
func (c *allocTraceContext) spanAcquire(base uintptr, class uint8) {
	c.allocBase[class] = base
	c.reserve(1+1+8, 0)
	c.buf.write8(atEvSpanAcquire)
	c.buf.write8(class)
	c.buf.writep(base)
}

//go:nosplit
func (c *allocTraceContext) allocLarge(addr, size uintptr, noscan, array bool) {
	now := uint64(cputicks())
	if c.lastSync > now {
		systemstack(func() {
			throw("time went backwards in alloc")
		})
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

//go:nosplit
func (c *allocTraceContext) allocSmall(addr, dataSize, size, allocpc uintptr, class uint8, array bool) {
	if class < 2 {
		throw("allocSmall called for large span")
	}
	if c.allocBase[class] == 0 {
		systemstack(func() {
			print("runtime: class = ", class, ", pid = ", int32(c.pid)-1, "\n")
			throw("no span acquired")
		})
	}
	now := uint64(cputicks())
	if c.lastSync > now {
		throw("time went backwards in alloc")
	}

	r := uintptr(1 + 1 + 3 + 2 + 4)
	if allocpc != 0 {
		r += 9
	}
	c.reserve(r, now)
	ev := atEvAlloc
	if array {
		ev = atEvAllocArray
		if allocpc != 0 {
			ev = atEvAllocArrayPC
		}
	} else if allocpc != 0 {
		ev = atEvAllocPC
	}
	c.buf.write8(ev)
	c.buf.write8(class)
	c.buf.writep(addr - c.allocBase[class]) // 55296 max value.
	c.buf.writep(size - dataSize)           // 4095 max value.
	if allocpc != 0 {
		c.buf.writep(allocpc)
	}
	c.buf.write64(now - c.lastSync)
}

//go:systemstack
func (c *allocTraceContext) allocStack(base uintptr, order uint8) {
	now := uint64(cputicks())
	if c.lastSync > now {
		throw("time went backwards in stackAlloc")
	}
	c.reserve(1+1+sys.PtrSize+4, now)
	c.buf.write8(atEvStackAlloc)
	c.buf.write8(order)
	c.buf.writep(base)
	c.buf.write64(now - c.lastSync)
}

//go:systemstack
func (c *allocTraceContext) freeStack(base uintptr) {
	now := uint64(cputicks())
	if c.lastSync > now {
		throw("time went backwards in stackFree")
	}
	c.reserve(1+sys.PtrSize+4, now)
	c.buf.write8(atEvStackFree)
	c.buf.writep(base)
	c.buf.write64(now - c.lastSync)
}

//go:nosplit
func (c *allocTraceContext) spanRelease(base uintptr, class uint8) {
	if c.allocBase[class] != base {
		systemstack(func() {
			print("runtime: class = ", class, ", base = ", hex(base), ", pid = ", int32(c.pid)-1, "\n")
			print("runtime: allocBase[class] = ", c.allocBase[class], "\n")
			throw("released unacquired (?) span")
		})
	}
	c.allocBase[class] = 0
	c.reserve(1+1, 0)
	c.buf.write8(atEvSpanRelease)
	c.buf.write8(class)
}

//go:nosplit
func (c *allocTraceContext) sweepTerm() {
	now := uint64(cputicks())
	if c.lastSync > now {
		throw("time went backwards in sweep term")
	}
	c.reserve(1+4, now)
	c.buf.write8(atEvSweepTerm)
	c.buf.write64(now - c.lastSync)
}

//go:nosplit
func (c *allocTraceContext) markTerm() {
	now := uint64(cputicks())
	if c.lastSync > now {
		throw("time went backwards in mark term")
	}
	c.reserve(1+4, now)
	c.buf.write8(atEvMarkTerm)
	c.buf.write64(now - c.lastSync)
}

//go:nosplit
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

//go:nosplit
func (c *allocTraceContext) free(addr uintptr) {
	c.reserve(1+4, 0)
	c.buf.write8(atEvFree)
	c.buf.writep(addr - c.freeBase)
}

//go:nosplit
func (c *allocTraceContext) sweepEnd() {
	c.freeBase = 0
}

//go:nosplit
func (c *allocTraceContext) writeSync(now uint64) {
	c.lastSync = now
	c.buf.write8(atEvSync)
	c.buf.write64(now)
}

//go:nosplit
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

// Must be nosplit because it's called from nosplit
// contexts.
//go:nosplit
func (a *allocTraceBuf) hasSpace(b uintptr) bool {
	return a.len+b < uintptr(len(a.data)-1)
}

// Must be nosplit because it's called from nosplit
// contexts.
//go:nosplit
func (a *allocTraceBuf) writep(p uintptr) {
	a.write64(uint64(p))
}

// Must be nosplit because it's called from nosplit
// contexts.
//go:nosplit
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

// Must be nosplit because it's called from nosplit
// contexts.
//go:nosplit
func (a *allocTraceBuf) write8(b uint8) {
	a.data[a.len] = b
	a.len++
}

var allocTracePool struct {
	// Lock-free stack of allocTraceBufs.
	emptySize uint64
	empty     lfstack
	ready     lfstack

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
		buf := allocTracePool.reading
		if atomic.Load64(&allocTracePool.emptySize) > 4<<30 {
			sysFree(unsafe.Pointer(buf), unsafe.Sizeof(allocTraceBuf{}), &memstats.other_sys)
		} else {
			buf.len = 0
			atomic.Xadd64(&allocTracePool.emptySize, int64(unsafe.Sizeof(allocTraceBuf{})))
			allocTracePool.empty.push(&buf.lfnode)
		}
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

// stopAllocTrace stops accumulating allocation traces and flushes
// any in-progress allocation traces.
func stopAllocTrace() {
	stopTheWorldGC("stop allocation tracing")
	systemstack(func() {
		lock(&globalATStateLock)
		allocTraceEnabled = false
		for _, p := range allp[:cap(allp)] {
			p.mcache.atState.swapAllocTraceBuf(true, 0)
		}
		globalATState.swapAllocTraceBuf(true, 0)
		unlock(&globalATStateLock)
	})
	startTheWorldGC()
}

func allocTraceInit() {
	const prefix = "GOALLOCTRACE="
	var val int

	switch GOOS {
	case "aix", "darwin", "dragonfly", "freebsd", "netbsd", "openbsd", "illumos", "solaris", "linux":
		// Similar to goenv_unix but extracts the environment value for
		// GOALLOCTRACE directly.
		// TODO(moehrmann): remove when general goenvs() can be called before cpuinit()
		n := int32(0)
		for argv_index(argv, argc+1+n) != nil {
			n++
		}

		for i := int32(0); i < n; i++ {
			p := argv_index(argv, argc+1+i)
			s := *(*string)(unsafe.Pointer(&stringStruct{unsafe.Pointer(p), findnull(p)}))

			if hasPrefix(s, prefix) {
				v, ok := atoi(s[len(prefix):])
				if ok {
					val = v
				}
				break
			}
		}
	}
	if val > 0 {
		allocTraceEnabled = true
	}
}
