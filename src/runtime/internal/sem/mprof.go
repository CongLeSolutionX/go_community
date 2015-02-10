// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Malloc profiling.
// Patterned after tcmalloc's algorithms; shorter code.

package sem

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
	"unsafe"
)

// NOTE(rsc): Everything here could use cas if contention became an issue.
var Proflock _core.Mutex

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
	Prev_allocs      uintptr
	Prev_frees       uintptr
	Prev_alloc_bytes uintptr
	Prev_free_bytes  uintptr

	// changes since last GC
	Recent_allocs      uintptr
	Recent_frees       uintptr
	Recent_alloc_bytes uintptr
	Recent_free_bytes  uintptr
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
		_lock.Throw("invalid profile bucket type")
	case MemProfile:
		size += unsafe.Sizeof(MemRecord{})
	case BlockProfile:
		size += unsafe.Sizeof(BlockRecord{})
	}

	b := (*Bucket)(_lock.Persistentalloc(size, 0, &_lock.Memstats.Buckhash_sys))
	bucketmem += size
	b.typ = typ
	b.Nstk = uintptr(nstk)
	return b
}

// stk returns the slice in b holding the stack.
func (b *Bucket) Stk() []uintptr {
	stk := (*[MaxStack]uintptr)(_core.Add(unsafe.Pointer(b), unsafe.Sizeof(*b)))
	return stk[:b.Nstk:b.Nstk]
}

// mp returns the memRecord associated with the memProfile bucket b.
func (b *Bucket) Mp() *MemRecord {
	if b.typ != MemProfile {
		_lock.Throw("bad use of bucket.mp")
	}
	data := _core.Add(unsafe.Pointer(b), unsafe.Sizeof(*b)+b.Nstk*unsafe.Sizeof(uintptr(0)))
	return (*MemRecord)(data)
}

// bp returns the blockRecord associated with the blockProfile bucket b.
func (b *Bucket) Bp() *BlockRecord {
	if b.typ != BlockProfile {
		_lock.Throw("bad use of bucket.bp")
	}
	data := _core.Add(unsafe.Pointer(b), unsafe.Sizeof(*b)+b.Nstk*unsafe.Sizeof(uintptr(0)))
	return (*BlockRecord)(data)
}

// Return the bucket for stk[0:nstk], allocating new bucket if needed.
func Stkbucket(typ bucketType, size uintptr, stk []uintptr, alloc bool) *Bucket {
	if buckhash == nil {
		buckhash = (*[BuckHashSize]*Bucket)(_lock.SysAlloc(unsafe.Sizeof(*buckhash), &_lock.Memstats.Buckhash_sys))
		if buckhash == nil {
			_lock.Throw("runtime: cannot allocate memory")
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

var Blockprofilerate uint64 // in CPU ticks

func Blockevent(cycles int64, skip int) {
	if cycles <= 0 {
		cycles = 1
	}
	rate := int64(_sched.Atomicload64(&Blockprofilerate))
	if rate <= 0 || (rate > cycles && int64(_lock.Fastrand1())%rate > cycles) {
		return
	}
	gp := _core.Getg()
	var nstk int
	var stk [MaxStack]uintptr
	if gp.M.Curg == nil || gp.M.Curg == gp {
		nstk = _sched.Callers(skip, &stk[0], len(stk))
	} else {
		nstk = _sched.Gcallers(gp.M.Curg, skip, &stk[0], len(stk))
	}
	_lock.Lock(&Proflock)
	b := Stkbucket(BlockProfile, 0, stk[:nstk], true)
	b.Bp().Count++
	b.Bp().Cycles += cycles
	_lock.Unlock(&Proflock)
}
