// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sync

import (
	"internal/race"
	"runtime"
	"sync/atomic"
	"unsafe"
)

// A Pool is a set of temporary objects that may be individually saved and
// retrieved.
//
// Any item stored in the Pool may be removed automatically at any time without
// notification. If the Pool holds the only reference when this happens, the
// item might be deallocated.
//
// A Pool is safe for use by multiple goroutines simultaneously.
//
// Pool's purpose is to cache allocated but unused items for later reuse,
// relieving pressure on the garbage collector. That is, it makes it easy to
// build efficient, thread-safe free lists. However, it is not suitable for all
// free lists.
//
// An appropriate use of a Pool is to manage a group of temporary items
// silently shared among and potentially reused by concurrent independent
// clients of a package. Pool provides a way to amortize allocation overhead
// across many clients.
//
// An example of good use of a Pool is in the fmt package, which maintains a
// dynamically-sized store of temporary output buffers. The store scales under
// load (when many goroutines are actively printing) and shrinks when
// quiescent.
//
// On the other hand, a free list maintained as part of a short-lived object is
// not a suitable use for a Pool, since the overhead does not amortize well in
// that scenario. It is more efficient to have such objects implement their own
// free list.
//
// A Pool must not be copied after first use.
type Pool struct {
	noCopy noCopy

	local     unsafe.Pointer // local fixed-size per-P pool, actual type is [P]poolLocal
	localSize uintptr        // size of the local array

	globalLock  uintptr    // mutex for access to global/globalEmpty
	global      *poolShard // global pool of full shards (elems==shardSize)
	globalEmpty *poolShard // global pool of empty shards (elems==0)

	// New optionally specifies a function to generate
	// a value when Get would otherwise return nil.
	// It may not be changed concurrently with calls to Get.
	New func() interface{}
}

const (
	globalLocked   uintptr = 1
	globalUnlocked         = 0

	shardSize = 32 // number of elements per shard
)

type poolShardInternal struct {
	next  *poolShard
	elems int
	elem  [shardSize]interface{}
}

type poolShard struct {
	poolShardInternal

	// Prevents false sharing on widespread platforms with
	// 128 mod (cache line size) = 0.
	_ [128 - unsafe.Sizeof(poolShardInternal{})%128]byte
}

// Local per-P Pool appendix.
type poolLocal struct {
	private poolShard
}

// from runtime
func fastrand() uint32

var poolRaceHash [128]uint64

// poolRaceAddr returns an address to use as the synchronization point
// for race detector logic. We don't use the actual pointer stored in x
// directly, for fear of conflicting with other synchronization on that address.
// Instead, we hash the pointer to get an index into poolRaceHash.
// See discussion on golang.org/cl/31589.
func poolRaceAddr(x interface{}) unsafe.Pointer {
	ptr := uintptr((*[2]unsafe.Pointer)(unsafe.Pointer(&x))[1])
	h := uint32((uint64(uint32(ptr)) * 0x85ebca6b) >> 16)
	return unsafe.Pointer(&poolRaceHash[h%uint32(len(poolRaceHash))])
}

// Put adds x to the pool.
func (p *Pool) Put(x interface{}) {
	if x == nil {
		return
	}

	if race.Enabled {
		if fastrand()%4 == 0 {
			// Randomly drop x on floor.
			return
		}
		race.ReleaseMerge(poolRaceAddr(x))
		race.Disable()
	}

	l := p.pin()
	if l.private.elems < shardSize {
		l.private.elem[l.private.elems] = x
		l.private.elems++
	} else if next := l.private.next; next != nil && next.elems < shardSize {
		next.elem[next.elems] = x
		next.elems++
	} else if p.globalLockIfUnlocked() {
		// There is no space in the private pool but we were able to acquire
		// the globalLock, so we can try to move shards to/from the global pools.
		if l.private.next != nil {
			// The l.private.next shard is full: move it to the global pool.
			full := l.private.next
			l.private.next = nil
			full.next = p.global
			p.global = full
		}
		if p.globalEmpty != nil {
			// Grab a reusable empty shard from the globalEmpty pool and move it
			// to the private pool.
			empty := p.globalEmpty
			p.globalEmpty = empty.next
			empty.next = nil
			l.private.next = empty
			p.globalUnlock()
		} else {
			// The globalEmpty pool contains no reusable shards: allocate a new
			// empty shard.
			p.globalUnlock()
			l.private.next = &poolShard{}
		}
		l.private.next.elem[0] = x
		l.private.next.elems = 1
	} else {
		// We could not acquire the globalLock to recycle x: drop it on the floor.
	}
	runtime_procUnpin()

	if race.Enabled {
		race.Enable()
	}
}

// Get selects an arbitrary item from the Pool, removes it from the
// Pool, and returns it to the caller.
// Get may choose to ignore the pool and treat it as empty.
// Callers should not assume any relation between values passed to Put and
// the values returned by Get.
//
// If Get would otherwise return nil and p.New is non-nil, Get returns
// the result of calling p.New.
func (p *Pool) Get() interface{} {
	if race.Enabled {
		race.Disable()
	}

	l := p.pin()
	var x interface{}
	if l.private.elems > 0 {
		l.private.elems--
		x = l.private.elem[l.private.elems]
	} else if next := l.private.next; next != nil && next.elems > 0 {
		next.elems--
		x = next.elem[next.elems]
	} else if p.globalLockIfUnlocked() {
		// The private pool is empty but we were able to acquire the globalLock,
		// so we can try to move shards to/from the global pools.
		if l.private.next != nil {
			// The l.private.next shard is empty: move it to the globalFree pool.
			empty := l.private.next
			l.private.next = nil
			empty.next = p.globalEmpty
			p.globalEmpty = empty
		}
		if p.global != nil {
			// Grab one full shard from the global pool and move it to the private
			// pool.
			full := p.global
			p.global = full.next
			full.next = nil
			l.private.next = full
			if full.elems > 0 {
				full.elems--
				x = full.elem[full.elems]
			}
		}
		p.globalUnlock()
	} else {
		// The local pool was empty and we could not acquire the globalLock.
	}
	runtime_procUnpin()

	if race.Enabled {
		race.Enable()
		if x != nil {
			race.Acquire(poolRaceAddr(x))
		}
	}

	if x == nil && p.New != nil {
		x = p.New()
	}
	return x
}

// pin pins the current goroutine to P, disables preemption and returns poolLocal pool for the P.
// Caller must call runtime_procUnpin() when done with the pool.
func (p *Pool) pin() *poolLocal {
	pid := runtime_procPin()
	// In pinSlow we store to localSize and then to local, here we load in opposite order.
	// Since we've disabled preemption, GC cannot happen in between.
	// Thus here we must observe local at least as large localSize.
	// We can observe a newer/larger local, it is fine (we must observe its zero-initialized-ness).
	s := atomic.LoadUintptr(&p.localSize) // load-acquire
	l := p.local                          // load-consume
	if uintptr(pid) < s {
		return indexLocal(l, pid)
	}
	return p.pinSlow()
}

func (p *Pool) pinSlow() *poolLocal {
	// Retry under the mutex.
	// Can not lock the mutex while pinned.
	runtime_procUnpin()
	allPoolsMu.Lock()
	defer allPoolsMu.Unlock()
	pid := runtime_procPin()
	// poolCleanup won't be called while we are pinned.
	s := p.localSize
	l := p.local
	if uintptr(pid) < s {
		return indexLocal(l, pid)
	}
	if p.local == nil {
		allPools = append(allPools, p)
	}
	// If GOMAXPROCS changes between GCs, we re-allocate the array and lose the old one.
	size := runtime.GOMAXPROCS(0)
	local := make([]poolLocal, size)
	atomic.StorePointer(&p.local, unsafe.Pointer(&local[0])) // store-release
	atomic.StoreUintptr(&p.localSize, uintptr(size))         // store-release
	return &local[pid]
}

// globalLockIfUnlocked attempts to lock the globalLock. If the globalLock is
// already locked it returns false. Otherwise it locks it and returns true.
// This function is very similar to try_lock in POSIX, and it is equivalent
// to the uncontended fast path of Mutex.Lock. If this function returns true
// the caller has to call p.globalUnlock() to unlock the globalLock.
func (p *Pool) globalLockIfUnlocked() bool {
	if atomic.CompareAndSwapUintptr(&p.globalLock, globalUnlocked, globalLocked) {
		if race.Enabled {
			race.Acquire(unsafe.Pointer(p))
		}
		return true
	}
	return false
}

// globalUnlcok unlocks the globalLock. Calling this function should be done
// only if the last call to p.globalLockIfUnlocked() returned true: its behavior
// is otherwise undefined.
func (p *Pool) globalUnlock() {
	if race.Enabled {
		_, _ = p.global, p.globalEmpty
		race.Release(unsafe.Pointer(p))
	}
	atomic.StoreUintptr(&p.globalLock, globalUnlocked)
}

func poolCleanup() {
	// This function is called with the world stopped, at the beginning of a garbage collection.
	// It must not allocate and probably should not call any runtime functions.
	// Defensively zero out everything, 2 reasons:
	// 1. To prevent false retention of whole Pools.
	// 2. If GC happens while a goroutine works with l.shared in Put/Get,
	//    it will retain whole Pool. So next cycle memory consumption would be doubled.
	for i, p := range allPools {
		allPools[i] = nil
		for i := 0; i < int(p.localSize); i++ {
			l := indexLocal(p.local, i)
			for j := range l.private.elem {
				l.private.elem[j] = nil
			}
			l.private.elems = 0
			l.private.next = nil
		}
		p.global = nil
		p.globalEmpty = nil
		p.local = nil
		p.localSize = 0
	}
	allPools = []*Pool{}
}

var (
	allPoolsMu Mutex
	allPools   []*Pool
)

func init() {
	runtime_registerPoolCleanup(poolCleanup)
}

func indexLocal(l unsafe.Pointer, i int) *poolLocal {
	lp := unsafe.Pointer(uintptr(l) + uintptr(i)*unsafe.Sizeof(poolLocal{}))
	return (*poolLocal)(lp)
}

// Implemented in runtime.
func runtime_registerPoolCleanup(cleanup func())
func runtime_procPin() int
func runtime_procUnpin()
