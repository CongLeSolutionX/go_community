// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Export guts for testing.

package runtime

import (
	"math/bits"
	"runtime/internal/atomic"
	"runtime/internal/sys"
	"unsafe"
)

var Fadd64 = fadd64
var Fsub64 = fsub64
var Fmul64 = fmul64
var Fdiv64 = fdiv64
var F64to32 = f64to32
var F32to64 = f32to64
var Fcmp64 = fcmp64
var Fintto64 = fintto64
var F64toint = f64toint

var Entersyscall = entersyscall
var Exitsyscall = exitsyscall
var LockedOSThread = lockedOSThread
var Xadduintptr = atomic.Xadduintptr

var FuncPC = funcPC

var Fastlog2 = fastlog2

var Atoi = atoi
var Atoi32 = atoi32

var Nanotime = nanotime

var PhysHugePageSize = physHugePageSize

type LFNode struct {
	Next    uint64
	Pushcnt uintptr
}

func LFStackPush(head *uint64, node *LFNode) {
	(*lfstack)(head).push((*lfnode)(unsafe.Pointer(node)))
}

func LFStackPop(head *uint64) *LFNode {
	return (*LFNode)(unsafe.Pointer((*lfstack)(head).pop()))
}

func GCMask(x interface{}) (ret []byte) {
	systemstack(func() {
		ret = getgcmask(x)
	})
	return
}

func RunSchedLocalQueueTest() {
	_p_ := new(p)
	gs := make([]g, len(_p_.runq))
	for i := 0; i < len(_p_.runq); i++ {
		if g, _ := runqget(_p_); g != nil {
			throw("runq is not empty initially")
		}
		for j := 0; j < i; j++ {
			runqput(_p_, &gs[i], false)
		}
		for j := 0; j < i; j++ {
			if g, _ := runqget(_p_); g != &gs[i] {
				print("bad element at iter ", i, "/", j, "\n")
				throw("bad element")
			}
		}
		if g, _ := runqget(_p_); g != nil {
			throw("runq is not empty afterwards")
		}
	}
}

func RunSchedLocalQueueStealTest() {
	p1 := new(p)
	p2 := new(p)
	gs := make([]g, len(p1.runq))
	for i := 0; i < len(p1.runq); i++ {
		for j := 0; j < i; j++ {
			gs[j].sig = 0
			runqput(p1, &gs[j], false)
		}
		gp := runqsteal(p2, p1, true)
		s := 0
		if gp != nil {
			s++
			gp.sig++
		}
		for {
			gp, _ = runqget(p2)
			if gp == nil {
				break
			}
			s++
			gp.sig++
		}
		for {
			gp, _ = runqget(p1)
			if gp == nil {
				break
			}
			gp.sig++
		}
		for j := 0; j < i; j++ {
			if gs[j].sig != 1 {
				print("bad element ", j, "(", gs[j].sig, ") at iter ", i, "\n")
				throw("bad element")
			}
		}
		if s != i/2 && s != i/2+1 {
			print("bad steal ", s, ", want ", i/2, " or ", i/2+1, ", iter ", i, "\n")
			throw("bad steal")
		}
	}
}

func RunSchedLocalQueueEmptyTest(iters int) {
	// Test that runq is not spuriously reported as empty.
	// Runq emptiness affects scheduling decisions and spurious emptiness
	// can lead to underutilization (both runnable Gs and idle Ps coexist
	// for arbitrary long time).
	done := make(chan bool, 1)
	p := new(p)
	gs := make([]g, 2)
	ready := new(uint32)
	for i := 0; i < iters; i++ {
		*ready = 0
		next0 := (i & 1) == 0
		next1 := (i & 2) == 0
		runqput(p, &gs[0], next0)
		go func() {
			for atomic.Xadd(ready, 1); atomic.Load(ready) != 2; {
			}
			if runqempty(p) {
				println("next:", next0, next1)
				throw("queue is empty")
			}
			done <- true
		}()
		for atomic.Xadd(ready, 1); atomic.Load(ready) != 2; {
		}
		runqput(p, &gs[1], next1)
		runqget(p)
		<-done
		runqget(p)
	}
}

var (
	StringHash = stringHash
	BytesHash  = bytesHash
	Int32Hash  = int32Hash
	Int64Hash  = int64Hash
	MemHash    = memhash
	MemHash32  = memhash32
	MemHash64  = memhash64
	EfaceHash  = efaceHash
	IfaceHash  = ifaceHash
)

var UseAeshash = &useAeshash

func MemclrBytes(b []byte) {
	s := (*slice)(unsafe.Pointer(&b))
	memclrNoHeapPointers(s.array, uintptr(s.len))
}

var HashLoad = &hashLoad

// entry point for testing
func GostringW(w []uint16) (s string) {
	systemstack(func() {
		s = gostringw(&w[0])
	})
	return
}

type Uintreg sys.Uintreg

var Open = open
var Close = closefd
var Read = read
var Write = write

func Envs() []string     { return envs }
func SetEnvs(e []string) { envs = e }

var BigEndian = sys.BigEndian

// For benchmarking.

func BenchSetType(n int, x interface{}) {
	e := *efaceOf(&x)
	t := e._type
	var size uintptr
	var p unsafe.Pointer
	switch t.kind & kindMask {
	case kindPtr:
		t = (*ptrtype)(unsafe.Pointer(t)).elem
		size = t.size
		p = e.data
	case kindSlice:
		slice := *(*struct {
			ptr      unsafe.Pointer
			len, cap uintptr
		})(e.data)
		t = (*slicetype)(unsafe.Pointer(t)).elem
		size = t.size * slice.len
		p = slice.ptr
	}
	allocSize := roundupsize(size)
	systemstack(func() {
		for i := 0; i < n; i++ {
			heapBitsSetType(uintptr(p), allocSize, size, t)
		}
	})
}

const PtrSize = sys.PtrSize

var ForceGCPeriod = &forcegcperiod

// SetTracebackEnv is like runtime/debug.SetTraceback, but it raises
// the "environment" traceback level, so later calls to
// debug.SetTraceback (e.g., from testing timeouts) can't lower it.
func SetTracebackEnv(level string) {
	setTraceback(level)
	traceback_env = traceback_cache
}

var ReadUnaligned32 = readUnaligned32
var ReadUnaligned64 = readUnaligned64

func CountPagesInUse() (pagesInUse, counted uintptr) {
	stopTheWorld("CountPagesInUse")

	pagesInUse = uintptr(mheap_.pagesInUse)

	for _, s := range mheap_.allspans {
		if s.state == mSpanInUse {
			counted += s.npages
		}
	}

	startTheWorld()

	return
}

func Fastrand() uint32          { return fastrand() }
func Fastrandn(n uint32) uint32 { return fastrandn(n) }

type ProfBuf profBuf

func NewProfBuf(hdrsize, bufwords, tags int) *ProfBuf {
	return (*ProfBuf)(newProfBuf(hdrsize, bufwords, tags))
}

func (p *ProfBuf) Write(tag *unsafe.Pointer, now int64, hdr []uint64, stk []uintptr) {
	(*profBuf)(p).write(tag, now, hdr, stk)
}

const (
	ProfBufBlocking    = profBufBlocking
	ProfBufNonBlocking = profBufNonBlocking
)

func (p *ProfBuf) Read(mode profBufReadMode) ([]uint64, []unsafe.Pointer, bool) {
	return (*profBuf)(p).read(profBufReadMode(mode))
}

func (p *ProfBuf) Close() {
	(*profBuf)(p).close()
}

// ReadMemStatsSlow returns both the runtime-computed MemStats and
// MemStats accumulated by scanning the heap.
func ReadMemStatsSlow() (base, slow MemStats) {
	stopTheWorld("ReadMemStatsSlow")

	// Run on the system stack to avoid stack growth allocation.
	systemstack(func() {
		// Make sure stats don't change.
		getg().m.mallocing++

		readmemstats_m(&base)

		// Initialize slow from base and zero the fields we're
		// recomputing.
		slow = base
		slow.Alloc = 0
		slow.TotalAlloc = 0
		slow.Mallocs = 0
		slow.Frees = 0
		slow.HeapReleased = 0
		var bySize [_NumSizeClasses]struct {
			Mallocs, Frees uint64
		}

		// Add up current allocations in spans.
		for _, s := range mheap_.allspans {
			if s.state != mSpanInUse {
				continue
			}
			if sizeclass := s.spanclass.sizeclass(); sizeclass == 0 {
				slow.Mallocs++
				slow.Alloc += uint64(s.elemsize)
			} else {
				slow.Mallocs += uint64(s.allocCount)
				slow.Alloc += uint64(s.allocCount) * uint64(s.elemsize)
				bySize[sizeclass].Mallocs += uint64(s.allocCount)
			}
		}

		// Add in frees. readmemstats_m flushed the cached stats, so
		// these are up-to-date.
		var smallFree uint64
		slow.Frees = mheap_.nlargefree
		for i := range mheap_.nsmallfree {
			slow.Frees += mheap_.nsmallfree[i]
			bySize[i].Frees = mheap_.nsmallfree[i]
			bySize[i].Mallocs += mheap_.nsmallfree[i]
			smallFree += mheap_.nsmallfree[i] * uint64(class_to_size[i])
		}
		slow.Frees += memstats.tinyallocs
		slow.Mallocs += slow.Frees

		slow.TotalAlloc = slow.Alloc + mheap_.largefree + smallFree

		for i := range slow.BySize {
			slow.BySize[i].Mallocs = bySize[i].Mallocs
			slow.BySize[i].Frees = bySize[i].Frees
		}

		for i := mheap_.pages.start; i < mheap_.pages.end; i++ {
			chunk := mheap_.pages.chunks[i]
			if chunk != nil {
				pg := chunk.scavenged.popcntRange(0, mallocChunkPages)
				slow.HeapReleased += uint64(pg) * pageSize
			}
		}
		for _, p := range allp {
			pg := bits.OnesCount64(p.pcache.cache & p.pcache.scav)
			slow.HeapReleased += uint64(pg) * pageSize
		}

		// Unused space in the current arena also counts as released space.
		slow.HeapReleased += uint64(mheap_.curArena.end - mheap_.curArena.base)

		getg().m.mallocing--
	})

	startTheWorld()
	return
}

// BlockOnSystemStack switches to the system stack, prints "x\n" to
// stderr, and blocks in a stack containing
// "runtime.blockOnSystemStackInternal".
func BlockOnSystemStack() {
	systemstack(blockOnSystemStackInternal)
}

func blockOnSystemStackInternal() {
	print("x\n")
	lock(&deadlock)
	lock(&deadlock)
}

type RWMutex struct {
	rw rwmutex
}

func (rw *RWMutex) RLock() {
	rw.rw.rlock()
}

func (rw *RWMutex) RUnlock() {
	rw.rw.runlock()
}

func (rw *RWMutex) Lock() {
	rw.rw.lock()
}

func (rw *RWMutex) Unlock() {
	rw.rw.unlock()
}

const RuntimeHmapSize = unsafe.Sizeof(hmap{})

func MapBucketsCount(m map[int]int) int {
	h := *(**hmap)(unsafe.Pointer(&m))
	return 1 << h.B
}

func MapBucketsPointerIsNil(m map[int]int) bool {
	h := *(**hmap)(unsafe.Pointer(&m))
	return h.buckets == nil
}

func LockOSCounts() (external, internal uint32) {
	g := getg()
	if g.m.lockedExt+g.m.lockedInt == 0 {
		if g.lockedm != 0 {
			panic("lockedm on non-locked goroutine")
		}
	} else {
		if g.lockedm == 0 {
			panic("nil lockedm on locked goroutine")
		}
	}
	return g.m.lockedExt, g.m.lockedInt
}

//go:noinline
func TracebackSystemstack(stk []uintptr, i int) int {
	if i == 0 {
		pc, sp := getcallerpc(), getcallersp()
		return gentraceback(pc, sp, 0, getg(), 0, &stk[0], len(stk), nil, nil, _TraceJumpStack)
	}
	n := 0
	systemstack(func() {
		n = TracebackSystemstack(stk, i-1)
	})
	return n
}

func KeepNArenaHints(n int) {
	hint := mheap_.arenaHints
	for i := 1; i < n; i++ {
		hint = hint.next
		if hint == nil {
			return
		}
	}
	hint.next = nil
}

// MapNextArenaHint reserves a page at the next arena growth hint,
// preventing the arena from growing there, and returns the range of
// addresses that are no longer viable.
func MapNextArenaHint() (start, end uintptr) {
	hint := mheap_.arenaHints
	addr := hint.addr
	if hint.down {
		start, end = addr-heapArenaBytes, addr
		addr -= physPageSize
	} else {
		start, end = addr, addr+heapArenaBytes
	}
	sysReserve(unsafe.Pointer(addr), physPageSize)
	return
}

func GetNextArenaHint() uintptr {
	return mheap_.arenaHints.addr
}

type G = g

func Getg() *G {
	return getg()
}

//go:noinline
func PanicForTesting(b []byte, i int) byte {
	return unexportedPanicForTesting(b, i)
}

//go:noinline
func unexportedPanicForTesting(b []byte, i int) byte {
	return b[i]
}

func G0StackOverflow() {
	systemstack(func() {
		stackOverflow(nil)
	})
}

func stackOverflow(x *byte) {
	var buf [256]byte
	stackOverflow(&buf[0])
}

func MapTombstoneCheck(m map[int]int) {
	// Make sure emptyOne and emptyRest are distributed correctly.
	// We should have a series of filled and emptyOne cells, followed by
	// a series of emptyRest cells.
	h := *(**hmap)(unsafe.Pointer(&m))
	i := interface{}(m)
	t := *(**maptype)(unsafe.Pointer(&i))

	for x := 0; x < 1<<h.B; x++ {
		b0 := (*bmap)(add(h.buckets, uintptr(x)*uintptr(t.bucketsize)))
		n := 0
		for b := b0; b != nil; b = b.overflow(t) {
			for i := 0; i < bucketCnt; i++ {
				if b.tophash[i] != emptyRest {
					n++
				}
			}
		}
		k := 0
		for b := b0; b != nil; b = b.overflow(t) {
			for i := 0; i < bucketCnt; i++ {
				if k < n && b.tophash[i] == emptyRest {
					panic("early emptyRest")
				}
				if k >= n && b.tophash[i] != emptyRest {
					panic("late non-emptyRest")
				}
				if k == n-1 && b.tophash[i] == emptyOne {
					panic("last non-emptyRest entry is emptyOne")
				}
				k++
			}
		}
	}
}

func RunGetgThreadSwitchTest() {
	// Test that getg works correctly with thread switch.
	// With gccgo, if we generate getg inlined, the backend
	// may cache the address of the TLS variable, which
	// will become invalid after a thread switch. This test
	// checks that the bad caching doesn't happen.

	ch := make(chan int)
	go func(ch chan int) {
		ch <- 5
		LockOSThread()
	}(ch)

	g1 := getg()

	// Block on a receive. This is likely to get us a thread
	// switch. If we yield to the sender goroutine, it will
	// lock the thread, forcing us to resume on a different
	// thread.
	<-ch

	g2 := getg()
	if g1 != g2 {
		panic("g1 != g2")
	}

	// Also test getg after some control flow, as the
	// backend is sensitive to control flow.
	g3 := getg()
	if g1 != g3 {
		panic("g1 != g3")
	}
}

const (
	PageSize         = pageSize
	MallocChunkPages = mallocChunkPages
)

// Expose mallocSum for testing.
type MallocSum mallocSum

func PackMallocSum(start, max, end int) MallocSum { return MallocSum(packMallocSum(start, max, end)) }
func (m MallocSum) Start() int                    { return mallocSum(m).start() }
func (m MallocSum) Max() int                      { return mallocSum(m).max() }
func (m MallocSum) End() int                      { return mallocSum(m).end() }

// Expose mallocBits for testing.
type MallocBits mallocBits

func (b *MallocBits) Find(npages uintptr, hint int) (int, int) {
	return (*mallocBits)(b).find(npages, hint)
}
func (b *MallocBits) AllocRange(i, n int)      { (*mallocBits)(b).allocRange(i, n) }
func (b *MallocBits) Free(i, n int)            { (*mallocBits)(b).free(i, n) }
func (b *MallocBits) Summarize() MallocSum     { return MallocSum((*mallocBits)(b).summarize()) }
func (b *MallocBits) PopcntRange(i, n int) int { return (*pageBits)(b).popcntRange(i, n) }

// Expose non-trivial alloc helpers for testing.
func SetConsecBits64(x uint64, i, n int) uint64   { return setConsecBits64(x, i, n) }
func ClearConsecBits64(x uint64, i, n int) uint64 { return clearConsecBits64(x, i, n) }
func FindConsecN64(c uint64, n int) int           { return findConsecN64(c, n) }

// Expose mallocData for testing.
type MallocData mallocData

func (d *MallocData) FindScavengeCandidate(hint, max int) (int, int) {
	return (*mallocData)(d).findScavengeCandidate(hint, max)
}
func (d *MallocData) AllocRange(i, n int) { (*mallocData)(d).allocRange(i, n) }
func (d *MallocData) ScavengeRange(i, n int) {
	(*mallocData)(d).scavengeRange(i, n)
}

// Expose pageCache for testing.
type PageCache pageCache

const PageCacheSize = pageCacheSize

func NewPageCache(base uintptr, cache, scav uint64) PageCache {
	return PageCache(pageCache{base: base, cache: cache, scav: scav})
}
func (c *PageCache) Base() uintptr { return (*pageCache)(c).base }
func (c *PageCache) Cache() uint64 { return (*pageCache)(c).cache }
func (c *PageCache) Scav() uint64  { return (*pageCache)(c).scav }
func (c *PageCache) Alloc(npages uintptr) (uintptr, uintptr) {
	return (*pageCache)(c).alloc(npages)
}
func (c *PageCache) Flush(s *PageAlloc) {
	(*pageCache)(c).flush((*pageAlloc)(s))
}

// Expose pageAlloc for testing. Note that because pageAlloc is
// not in the heap, so is PageAlloc.
//
//go:notinheap
type PageAlloc pageAlloc

func (p *PageAlloc) Alloc(npages uintptr) (uintptr, uintptr) {
	return (*pageAlloc)(p).alloc(npages)
}
func (p *PageAlloc) AllocToCache() PageCache {
	return PageCache((*pageAlloc)(p).allocToCache())
}
func (p *PageAlloc) Free(base, npages uintptr) {
	(*pageAlloc)(p).free(base, npages)
}
func (p *PageAlloc) Bounds() (uint, uint) {
	return uint((*pageAlloc)(p).start), uint((*pageAlloc)(p).end)
}
func (p *PageAlloc) HasChunk(i uint) bool {
	if (*pageAlloc)(p).chunks[i] != nil {
		return true
	}
	return false
}
func (p *PageAlloc) MallocBits(i uint) *MallocBits {
	return (*MallocBits)(&((*pageAlloc)(p).chunks[i].mallocBits))
}
func (p *PageAlloc) Scavenge(nbytes uintptr, locked bool) (r uintptr) {
	systemstack(func() {
		r = (*pageAlloc)(p).scavenge(nbytes, locked)
	})
	return
}

// InitScavState initializes the pageAlloc's scavenged bitmap by first clearing
// the bitmap and then applying 1 bits to the bit ranges for each arena in arenas.
func (p *PageAlloc) InitScavState(arenas map[int][]BitRange) {
	pp := (*pageAlloc)(p)
	for highaddr, init := range arenas {
		addr := uintptr(highaddr) * mallocChunkBytes
		ci := chunkIndex(addr)
		pp.chunks[ci].scavenged.clearRange(0, mallocChunkPages)
		for _, s := range init {
			pp.chunks[ci].scavengeRange(s.I, s.N)
		}
	}
	pp.resetScavengeAddr()
}

// BitRange represents a range over a bitmap.
type BitRange struct {
	I, N int // bit index and length in bits
}

type arenaL2Entry [1 << arenaL2Bits]*heapArena

// Must be notinheap because mheap and pageAlloc are.
//
//go:notinheap
type pageAllocDummy struct {
	pageAlloc pageAlloc
	dummyLock mutex
	inUse     bool
}

var pageAllocDummies [2]pageAllocDummy

// GetTestPageAlloc returns a cached dummy pageAlloc for testing.
//
// It initializes the dummy pageAlloc using a set of bit ranges
// for each chunk.
//
// All arenas are initialized to be completely scavenged, as any
// arena produced live would be. To initialize the arena's scavenged
// state, call InitScavState on the returned pageAlloc.
//
// Note that this function could fail if all cached pageAllocs are
// in-use. If more than the number above are required for a given
// test, increase the size of pageAllocDummies above. Furthermore,
// because of this caching mechanism, any pageAlloc tests using
// this function cannot be run in parallel with other pageAlloc tests.
//
// Although this caching mechanism seems arbitrary, it serves an
// important purpose. The biggest reason is that we cannot heap
// allocate any of pageAlloc's structures, because pageAlloc is
// go:notinheap. Furthermore, pageAlloc consists of enormous memory
// mappings that, while probably not problematic in small multiples,
// could have adverse effects if we could create as many as we want
// on-demand. This mechanism thus also helps limit the amount of
// memory mappings we create.
func GetTestPageAlloc(chunks map[int][]BitRange) *PageAlloc {
	// Search for an unused entry.
	var entry *pageAllocDummy
	for i := 0; i < len(pageAllocDummies); i++ {
		if !pageAllocDummies[i].inUse {
			entry = &pageAllocDummies[i]
			entry.inUse = true
			break
		}
	}
	if entry == nil {
		return nil
	}

	// We've got an entry, so initialize the pageAlloc.
	entry.pageAlloc.init(&entry.dummyLock, nil)
	entry.pageAlloc.test = true

	n := 0
	for i, init := range chunks {
		// Update mheap metadata.
		addr := uintptr(i) * mallocChunkBytes

		// Mark the chunk's existence in the pageAlloc.
		entry.pageAlloc.grow(addr, mallocChunkBytes, nil)

		// Initialize the bitmap and update pageAlloc metadata.
		//
		// Just like when running live, start with the arena
		// completely scavenged.
		chunk := entry.pageAlloc.chunks[chunkIndex(addr)]
		for _, s := range init {
			chunk.allocRange(s.I, s.N)
		}
		entry.pageAlloc.update(addr, mallocChunkPages, false, false)
		n++
	}
	return (*PageAlloc)(&entry.pageAlloc)
}

// PutTestPageAlloc returns a pageAlloc to the cache mechanism.
//
// It also frees/unmaps any resources associated with the pageAlloc.
func PutTestPageAlloc(pa *PageAlloc) {
	p := (*pageAlloc)(pa)

	// Find the corresponding entry.
	var entry *pageAllocDummy
	for i := 0; i < len(pageAllocDummies); i++ {
		if p == &pageAllocDummies[i].pageAlloc {
			entry = &pageAllocDummies[i]
		}
	}
	if entry == nil {
		panic("attempt to put back page alloc which wasn't taken?")
	}

	// Free all the mapped space for the summary levels.
	if pageAlloc64Bit != 0 {
		for l := 0; l < summaryLevels; l++ {
			sysFree(unsafe.Pointer(&p.summary[l][0]), uintptr(cap(p.summary[l]))*mallocSumBytes, nil)
		}
	} else {
		resSize := uintptr(0)
		for l := 0; l < summaryLevels; l++ {
			resSize += uintptr(cap(p.summary[l])) * mallocSumBytes
		}
		sysFree(unsafe.Pointer(&p.summary[0][0]), alignUp(resSize, physPageSize), nil)
	}

	// Free all the bitmap chunks in the fixalloc.
	//
	// There's no easy way to free the memory backing a fixalloc,
	// so we'll hold onto it and re-use it.
	for c := p.start; c < p.end; c++ {
		if p.chunks[c] != nil {
			p.mallocDataAlloc.free(unsafe.Pointer(p.chunks[c]))
		}
	}

	// Free the mapped space for chunks.
	sysFree(unsafe.Pointer(&p.chunks[0]), uintptr(cap(p.chunks))*sys.PtrSize, nil)

	// Clear all relevant fields, but hold on to the fixalloc.
	entry.pageAlloc = pageAlloc{mallocDataAlloc: p.mallocDataAlloc}
	entry.inUse = false
}

const BaseChunkIdx = 0xc000*pageAlloc64Bit + 0x200*pageAlloc32Bit

func PageBase(chunkIdx uintptr, pageIdx int) uintptr {
	return chunkIdx*mallocChunkBytes + uintptr(pageIdx*pageSize)
}

type BitsMismatch struct {
	Base      uintptr
	Got, Want uint64
}

func CheckScavengedBitsCleared(mismatches []BitsMismatch) (n int, ok bool) {
	ok = true

	// Run on the system stack to avoid stack growth allocation.
	systemstack(func() {
		getg().m.mallocing++

		// Lock so that we can safely access the bitmap.
		lock(&mheap_.lock)
	chunkLoop:
		for i := mheap_.pages.start; i < mheap_.pages.end; i++ {
			chunk := mheap_.pages.chunks[i]
			if chunk != nil {
				for j := 0; j < mallocChunkPages/64; j++ {
					// Run over each 64-bit bitmap section and ensure
					// scavenged is being cleared properly on allocation.
					// If a used bit and scavenged bit are both set, that's
					// an error, and could indicate a larger problem, or
					// an accounting problem.
					want := chunk.scavenged[j] &^ chunk.mallocBits[j]
					got := chunk.scavenged[j]
					if want != got {
						ok = false
						if n >= len(mismatches) {
							break chunkLoop
						}
						mismatches[n] = BitsMismatch{
							Base: chunkBase(i) + uintptr(j)*64*pageSize,
							Got:  got,
							Want: want,
						}
						n++
					}
				}
			}
		}
		unlock(&mheap_.lock)

		getg().m.mallocing--
	})
	return
}

func PageCachePagesLeaked() (leaked uintptr) {
	stopTheWorld("PageCachePagesLeaked")

	// Run on the system stack to avoid stack growth allocation.
	systemstack(func() {
		// Make sure nothing else could possibly be running.
		getg().m.mallocing++

		// Walk over destroyed Ps and look for unflushed caches.
		deadp := allp[len(allp):cap(allp)]
		for _, p := range deadp {
			leaked += uintptr(bits.OnesCount64(p.pcache.cache))
		}

		getg().m.mallocing--
	})

	startTheWorld()
	return
}
