// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Malloc profiling.
// Patterned after tcmalloc's algorithms; shorter code.

package runtime

import (
	_base "runtime/internal/base"
	_gc "runtime/internal/gc"
	_iface "runtime/internal/iface"
	_lock "runtime/internal/lock"
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
		r = int64(float64(rate) * float64(tickspersecond()) / (1000 * 1000 * 1000))
		if r == 0 {
			r = 1
		}
	}

	_base.Atomicstore64(&_lock.Blockprofilerate, uint64(r))
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
	_base.Lock(&_lock.Proflock)
	clear := true
	for b := _lock.Mbuckets; b != nil; b = b.Allnext {
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
		for b := _lock.Mbuckets; b != nil; b = b.Allnext {
			mp := b.Mp()
			if inuseZero || mp.Alloc_bytes != mp.Free_bytes {
				n++
			}
		}
	}
	if n <= len(p) {
		ok = true
		idx := 0
		for b := _lock.Mbuckets; b != nil; b = b.Allnext {
			mp := b.Mp()
			if inuseZero || mp.Alloc_bytes != mp.Free_bytes {
				record(&p[idx], b)
				idx++
			}
		}
	}
	_base.Unlock(&_lock.Proflock)
	return
}

// Write b's data to r.
func record(r *MemProfileRecord, b *_lock.Bucket) {
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

func iterate_memprof(fn func(*_lock.Bucket, uintptr, *uintptr, uintptr, uintptr, uintptr)) {
	_base.Lock(&_lock.Proflock)
	for b := _lock.Mbuckets; b != nil; b = b.Allnext {
		mp := b.Mp()
		fn(b, uintptr(b.Nstk), &b.Stk()[0], b.Size, mp.Allocs, mp.Frees)
	}
	_base.Unlock(&_lock.Proflock)
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
	_base.Lock(&_lock.Proflock)
	for b := _lock.Bbuckets; b != nil; b = b.Allnext {
		n++
	}
	if n <= len(p) {
		ok = true
		for b := _lock.Bbuckets; b != nil; b = b.Allnext {
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
	_base.Unlock(&_lock.Proflock)
	return
}

// ThreadCreateProfile returns n, the number of records in the thread creation profile.
// If len(p) >= n, ThreadCreateProfile copies the profile into p and returns n, true.
// If len(p) < n, ThreadCreateProfile does not change p and returns n, false.
//
// Most clients should use the runtime/pprof package instead
// of calling ThreadCreateProfile directly.
func ThreadCreateProfile(p []StackRecord) (n int, ok bool) {
	first := (*_base.M)(_iface.Atomicloadp(unsafe.Pointer(&_base.Allm)))
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

// GoroutineProfile returns n, the number of records in the active goroutine stack profile.
// If len(p) >= n, GoroutineProfile copies the profile into p and returns n, true.
// If len(p) < n, GoroutineProfile does not change p and returns n, false.
//
// Most clients should use the runtime/pprof package instead
// of calling GoroutineProfile directly.
func GoroutineProfile(p []StackRecord) (n int, ok bool) {

	n = NumGoroutine()
	if n <= len(p) {
		gp := _base.Getg()
		stopTheWorld("profile")

		n = NumGoroutine()
		if n <= len(p) {
			ok = true
			r := p
			sp := _base.Getcallersp(unsafe.Pointer(&p))
			pc := _base.Getcallerpc(unsafe.Pointer(&p))
			_base.Systemstack(func() {
				saveg(pc, sp, gp, &r[0])
			})
			r = r[1:]
			for _, gp1 := range _base.Allgs {
				if gp1 == gp || _base.Readgstatus(gp1) == _base.Gdead {
					continue
				}
				saveg(^uintptr(0), ^uintptr(0), gp1, &r[0])
				r = r[1:]
			}
		}

		startTheWorld()
	}

	return n, ok
}

func saveg(pc, sp uintptr, gp *_base.G, r *StackRecord) {
	n := _base.Gentraceback(pc, sp, 0, gp, 0, &r.Stack0[0], len(r.Stack0), nil, nil, 0)
	if n < len(r.Stack0) {
		r.Stack0[n] = 0
	}
}

// Stack formats a stack trace of the calling goroutine into buf
// and returns the number of bytes written to buf.
// If all is true, Stack formats stack traces of all other goroutines
// into buf after the trace for the current goroutine.
func Stack(buf []byte, all bool) int {
	if all {
		stopTheWorld("stack trace")
	}

	n := 0
	if len(buf) > 0 {
		gp := _base.Getg()
		sp := _base.Getcallersp(unsafe.Pointer(&buf))
		pc := _base.Getcallerpc(unsafe.Pointer(&buf))
		_base.Systemstack(func() {
			g0 := _base.Getg()
			g0.Writebuf = buf[0:0:len(buf)]
			_base.Goroutineheader(gp)
			_base.Traceback(pc, sp, 0, gp)
			if all {
				_base.Tracebackothers(gp)
			}
			n = len(g0.Writebuf)
			g0.Writebuf = nil
		})
	}

	if all {
		startTheWorld()
	}
	return n
}
