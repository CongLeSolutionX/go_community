// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Malloc profiling.
// Patterned after tcmalloc's algorithms; shorter code.

package prof

import (
	_core "runtime/internal/core"
	_gc "runtime/internal/gc"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
	_sem "runtime/internal/sem"
	"unsafe"
)

// SetBlockProfileRate controls the fraction of goroutine blocking events
// that are reported in the blocking profile.  The profiler aims to sample
// an average of one blocking event per rate nanoseconds spent blocked.
//
// To include every blocking event in the profile, pass rate = 1.
// To turn off profiling entirely, pass rate <= 0.
func SetBlockProfileRate(rate int) {
	var r int64
	if rate <= 0 {
		r = 0 // disable profiling
	} else if rate == 1 {
		r = 1 // profile everything
	} else {
		// convert ns to cycles, use float64 to prevent overflow during multiplication
		r = int64(float64(rate) * float64(Tickspersecond()) / (1000 * 1000 * 1000))
		if r == 0 {
			r = 1
		}
	}

	_sched.Atomicstore64(&_sem.Blockprofilerate, uint64(r))
}

// Go interface to profile data.

// A StackRecord describes a single execution stack.
type StackRecord struct {
	Stack0 [32]uintptr // stack trace for this record; ends at first 0 entry
}

// Stack returns the stack trace associated with the record,
// a prefix of r.Stack0.
func (r *StackRecord) Stack() []uintptr {
	for i, v := range r.Stack0 {
		if v == 0 {
			return r.Stack0[0:i]
		}
	}
	return r.Stack0[0:]
}

// A MemProfileRecord describes the live objects allocated
// by a particular call sequence (stack trace).
type MemProfileRecord struct {
	AllocBytes, FreeBytes     int64       // number of bytes allocated, freed
	AllocObjects, FreeObjects int64       // number of objects allocated, freed
	Stack0                    [32]uintptr // stack trace for this record; ends at first 0 entry
}

// InUseBytes returns the number of bytes in use (AllocBytes - FreeBytes).
func (r *MemProfileRecord) InUseBytes() int64 { return r.AllocBytes - r.FreeBytes }

// InUseObjects returns the number of objects in use (AllocObjects - FreeObjects).
func (r *MemProfileRecord) InUseObjects() int64 {
	return r.AllocObjects - r.FreeObjects
}

// Stack returns the stack trace associated with the record,
// a prefix of r.Stack0.
func (r *MemProfileRecord) Stack() []uintptr {
	for i, v := range r.Stack0 {
		if v == 0 {
			return r.Stack0[0:i]
		}
	}
	return r.Stack0[0:]
}

// MemProfile returns n, the number of records in the current memory profile.
// If len(p) >= n, MemProfile copies the profile into p and returns n, true.
// If len(p) < n, MemProfile does not change p and returns n, false.
//
// If inuseZero is true, the profile includes allocation records
// where r.AllocBytes > 0 but r.AllocBytes == r.FreeBytes.
// These are sites where memory was allocated, but it has all
// been released back to the runtime.
//
// Most clients should use the runtime/pprof package or
// the testing package's -test.memprofile flag instead
// of calling MemProfile directly.
func MemProfile(p []MemProfileRecord, inuseZero bool) (n int, ok bool) {
	_lock.Lock(&_sem.Proflock)
	clear := true
	for b := _sem.Mbuckets; b != nil; b = b.Allnext {
		mp := b.Mp()
		if inuseZero || mp.Alloc_bytes != mp.Free_bytes {
			n++
		}
		if mp.Allocs != 0 || mp.Frees != 0 {
			clear = false
		}
	}
	if clear {
		// Absolutely no data, suggesting that a garbage collection
		// has not yet happened. In order to allow profiling when
		// garbage collection is disabled from the beginning of execution,
		// accumulate stats as if a GC just happened, and recount buckets.
		_gc.Mprof_GC()
		_gc.Mprof_GC()
		n = 0
		for b := _sem.Mbuckets; b != nil; b = b.Allnext {
			mp := b.Mp()
			if inuseZero || mp.Alloc_bytes != mp.Free_bytes {
				n++
			}
		}
	}
	if n <= len(p) {
		ok = true
		idx := 0
		for b := _sem.Mbuckets; b != nil; b = b.Allnext {
			mp := b.Mp()
			if inuseZero || mp.Alloc_bytes != mp.Free_bytes {
				record(&p[idx], b)
				idx++
			}
		}
	}
	_lock.Unlock(&_sem.Proflock)
	return
}

// Write b's data to r.
func record(r *MemProfileRecord, b *_sem.Bucket) {
	mp := b.Mp()
	r.AllocBytes = int64(mp.Alloc_bytes)
	r.FreeBytes = int64(mp.Free_bytes)
	r.AllocObjects = int64(mp.Allocs)
	r.FreeObjects = int64(mp.Frees)
	copy(r.Stack0[:], b.Stk())
	for i := int(b.Nstk); i < len(r.Stack0); i++ {
		r.Stack0[i] = 0
	}
}

// BlockProfileRecord describes blocking events originated
// at a particular call sequence (stack trace).
type BlockProfileRecord struct {
	Count  int64
	Cycles int64
	StackRecord
}

// BlockProfile returns n, the number of records in the current blocking profile.
// If len(p) >= n, BlockProfile copies the profile into p and returns n, true.
// If len(p) < n, BlockProfile does not change p and returns n, false.
//
// Most clients should use the runtime/pprof package or
// the testing package's -test.blockprofile flag instead
// of calling BlockProfile directly.
func BlockProfile(p []BlockProfileRecord) (n int, ok bool) {
	_lock.Lock(&_sem.Proflock)
	for b := _sem.Bbuckets; b != nil; b = b.Allnext {
		n++
	}
	if n <= len(p) {
		ok = true
		for b := _sem.Bbuckets; b != nil; b = b.Allnext {
			bp := b.Bp()
			r := &p[0]
			r.Count = int64(bp.Count)
			r.Cycles = int64(bp.Cycles)
			i := copy(r.Stack0[:], b.Stk())
			for ; i < len(r.Stack0); i++ {
				r.Stack0[i] = 0
			}
			p = p[1:]
		}
	}
	_lock.Unlock(&_sem.Proflock)
	return
}

// ThreadCreateProfile returns n, the number of records in the thread creation profile.
// If len(p) >= n, ThreadCreateProfile copies the profile into p and returns n, true.
// If len(p) < n, ThreadCreateProfile does not change p and returns n, false.
//
// Most clients should use the runtime/pprof package instead
// of calling ThreadCreateProfile directly.
func ThreadCreateProfile(p []StackRecord) (n int, ok bool) {
	first := (*_core.M)(Atomicloadp(unsafe.Pointer(&_lock.Allm)))
	for mp := first; mp != nil; mp = mp.Alllink {
		n++
	}
	if n <= len(p) {
		ok = true
		i := 0
		for mp := first; mp != nil; mp = mp.Alllink {
			for s := range mp.Createstack {
				p[i].Stack0[s] = uintptr(mp.Createstack[s])
			}
			i++
		}
	}
	return
}
