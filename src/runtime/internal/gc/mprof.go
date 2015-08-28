// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Malloc profiling.
// Patterned after tcmalloc's algorithms; shorter code.

package gc

import (
	_base "runtime/internal/base"
	"unsafe"
)

// NOTE(rsc): Everything here could use cas if contention became an issue.
var Proflock _base.Mutex

// All memory allocations are local and do not escape outside of the profiler.
// The profiler is forbidden from referring to garbage-collected memory.

const (
	// profile types
	MemProfile bucketType = 1 + iota
	BlockProfile

	// size of bucket hash table
	BuckHashSize = 179999

	// max depth of stack to record in bucket
	MaxStack = 32
)

type bucketType int

// A bucket holds per-call-stack profiling information.
// The representation is a bit sleazy, inherited from C.
// This struct defines the bucket header. It is followed in
// memory by the stack words and then the actual record
// data, either a memRecord or a blockRecord.
//
// Per-call-stack profiling information.
// Lookup by hashing call stack into a linked-list hash table.
type Bucket struct {
	next    *Bucket
	Allnext *Bucket
	typ     bucketType // memBucket or blockBucket
	hash    uintptr
	Size    uintptr
	Nstk    uintptr
}

// A memRecord is the bucket data for a bucket of type memProfile,
// part of the memory profile.
type MemRecord struct {
	// The following complex 3-stage scheme of stats accumulation
	// is required to obtain a consistent picture of mallocs and frees
	// for some point in time.
	// The problem is that mallocs come in real time, while frees
	// come only after a GC during concurrent sweeping. So if we would
	// naively count them, we would get a skew toward mallocs.
	//
	// Mallocs are accounted in recent stats.
	// Explicit frees are accounted in recent stats.
	// GC frees are accounted in prev stats.
	// After GC prev stats are added to final stats and
	// recent stats are moved into prev stats.
	Allocs      uintptr
	Frees       uintptr
	Alloc_bytes uintptr
	Free_bytes  uintptr

	// changes between next-to-last GC and last GC
	prev_allocs      uintptr
	prev_frees       uintptr
	prev_alloc_bytes uintptr
	prev_free_bytes  uintptr

	// changes since last GC
	Recent_allocs      uintptr
	recent_frees       uintptr
	Recent_alloc_bytes uintptr
	recent_free_bytes  uintptr
}

// A blockRecord is the bucket data for a bucket of type blockProfile,
// part of the blocking profile.
type BlockRecord struct {
	Count  int64
	Cycles int64
}

var (
	Mbuckets  *Bucket // memory profile buckets
	Bbuckets  *Bucket // blocking profile buckets
	buckhash  *[179999]*Bucket
	bucketmem uintptr
)

// newBucket allocates a bucket with the given type and number of stack entries.
func newBucket(typ bucketType, nstk int) *Bucket {
	size := unsafe.Sizeof(Bucket{}) + uintptr(nstk)*unsafe.Sizeof(uintptr(0))
	switch typ {
	default:
		_base.Throw("invalid profile bucket type")
	case MemProfile:
		size += unsafe.Sizeof(MemRecord{})
	case BlockProfile:
		size += unsafe.Sizeof(BlockRecord{})
	}

	b := (*Bucket)(_base.Persistentalloc(size, 0, &_base.Memstats.Buckhash_sys))
	bucketmem += size
	b.typ = typ
	b.Nstk = uintptr(nstk)
	return b
}

// stk returns the slice in b holding the stack.
func (b *Bucket) Stk() []uintptr {
	stk := (*[MaxStack]uintptr)(_base.Add(unsafe.Pointer(b), unsafe.Sizeof(*b)))
	return stk[:b.Nstk:b.Nstk]
}

// mp returns the memRecord associated with the memProfile bucket b.
func (b *Bucket) Mp() *MemRecord {
	if b.typ != MemProfile {
		_base.Throw("bad use of bucket.mp")
	}
	data := _base.Add(unsafe.Pointer(b), unsafe.Sizeof(*b)+b.Nstk*unsafe.Sizeof(uintptr(0)))
	return (*MemRecord)(data)
}

// bp returns the blockRecord associated with the blockProfile bucket b.
func (b *Bucket) Bp() *BlockRecord {
	if b.typ != BlockProfile {
		_base.Throw("bad use of bucket.bp")
	}
	data := _base.Add(unsafe.Pointer(b), unsafe.Sizeof(*b)+b.Nstk*unsafe.Sizeof(uintptr(0)))
	return (*BlockRecord)(data)
}

// Return the bucket for stk[0:nstk], allocating new bucket if needed.
func Stkbucket(typ bucketType, size uintptr, stk []uintptr, alloc bool) *Bucket {
	if buckhash == nil {
		buckhash = (*[BuckHashSize]*Bucket)(_base.SysAlloc(unsafe.Sizeof(*buckhash), &_base.Memstats.Buckhash_sys))
		if buckhash == nil {
			_base.Throw("runtime: cannot allocate memory")
		}
	}

	// Hash stack.
	var h uintptr
	for _, pc := range stk {
		h += pc
		h += h << 10
		h ^= h >> 6
	}
	// hash in size
	h += size
	h += h << 10
	h ^= h >> 6
	// finalize
	h += h << 3
	h ^= h >> 11

	i := int(h % BuckHashSize)
	for b := buckhash[i]; b != nil; b = b.next {
		if b.typ == typ && b.hash == h && b.Size == size && eqslice(b.Stk(), stk) {
			return b
		}
	}

	if !alloc {
		return nil
	}

	// Create new bucket.
	b := newBucket(typ, len(stk))
	copy(b.Stk(), stk)
	b.hash = h
	b.Size = size
	b.next = buckhash[i]
	buckhash[i] = b
	if typ == MemProfile {
		b.Allnext = Mbuckets
		Mbuckets = b
	} else {
		b.Allnext = Bbuckets
		Bbuckets = b
	}
	return b
}

func eqslice(x, y []uintptr) bool {
	if len(x) != len(y) {
		return false
	}
	for i, xi := range x {
		if xi != y[i] {
			return false
		}
	}
	return true
}

func Mprof_GC() {
	for b := Mbuckets; b != nil; b = b.Allnext {
		mp := b.Mp()
		mp.Allocs += mp.prev_allocs
		mp.Frees += mp.prev_frees
		mp.Alloc_bytes += mp.prev_alloc_bytes
		mp.Free_bytes += mp.prev_free_bytes

		mp.prev_allocs = mp.Recent_allocs
		mp.prev_frees = mp.recent_frees
		mp.prev_alloc_bytes = mp.Recent_alloc_bytes
		mp.prev_free_bytes = mp.recent_free_bytes

		mp.Recent_allocs = 0
		mp.recent_frees = 0
		mp.Recent_alloc_bytes = 0
		mp.recent_free_bytes = 0
	}
}

// Record that a gc just happened: all the 'recent' statistics are now real.
func mProf_GC() {
	_base.Lock(&Proflock)
	Mprof_GC()
	_base.Unlock(&Proflock)
}

// Called when freeing a profiled block.
func mProf_Free(b *Bucket, size uintptr, freed bool) {
	_base.Lock(&Proflock)
	mp := b.Mp()
	if freed {
		mp.recent_frees++
		mp.recent_free_bytes += size
	} else {
		mp.prev_frees++
		mp.prev_free_bytes += size
	}
	_base.Unlock(&Proflock)
}

var Blockprofilerate uint64 // in CPU ticks

func Blockevent(cycles int64, skip int) {
	if cycles <= 0 {
		cycles = 1
	}
	rate := int64(_base.Atomicload64(&Blockprofilerate))
	if rate <= 0 || (rate > cycles && int64(_base.Fastrand1())%rate > cycles) {
		return
	}
	gp := _base.Getg()
	var nstk int
	var stk [MaxStack]uintptr
	if gp.M.Curg == nil || gp.M.Curg == gp {
		nstk = _base.Callers(skip, stk[:])
	} else {
		nstk = _base.Gcallers(gp.M.Curg, skip, stk[:])
	}
	_base.Lock(&Proflock)
	b := Stkbucket(BlockProfile, 0, stk[:nstk], true)
	b.Bp().Count++
	b.Bp().Cycles += cycles
	_base.Unlock(&Proflock)
}

// Tracing of alloc/free/gc.

var Tracelock _base.Mutex

func tracefree(p unsafe.Pointer, size uintptr) {
	_base.Lock(&Tracelock)
	gp := _base.Getg()
	gp.M.Traceback = 2
	print("tracefree(", p, ", ", _base.Hex(size), ")\n")
	_base.Goroutineheader(gp)
	pc := _base.Getcallerpc(unsafe.Pointer(&p))
	sp := _base.Getcallersp(unsafe.Pointer(&p))
	_base.Systemstack(func() {
		_base.Traceback(pc, sp, 0, gp)
	})
	print("\n")
	gp.M.Traceback = 0
	_base.Unlock(&Tracelock)
}

func tracegc() {
	_base.Lock(&Tracelock)
	gp := _base.Getg()
	gp.M.Traceback = 2
	print("tracegc()\n")
	// running on m->g0 stack; show all non-g0 goroutines
	_base.Tracebackothers(gp)
	print("end tracegc\n")
	print("\n")
	gp.M.Traceback = 0
	_base.Unlock(&Tracelock)
}
