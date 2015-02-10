// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Malloc profiling.
// Patterned after tcmalloc's algorithms; shorter code.

package maps

import (
	_core "runtime/internal/core"
	_gc "runtime/internal/gc"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
	_sem "runtime/internal/sem"
	"unsafe"
)

// Called by malloc to record a profiled block.
func mProf_Malloc(p unsafe.Pointer, size uintptr) {
	var stk [_sem.MaxStack]uintptr
	nstk := _sched.Callers(4, &stk[0], len(stk))
	_lock.Lock(&_sem.Proflock)
	b := _sem.Stkbucket(_sem.MemProfile, size, stk[:nstk], true)
	mp := b.Mp()
	mp.Recent_allocs++
	mp.Recent_alloc_bytes += size
	_lock.Unlock(&_sem.Proflock)

	// Setprofilebucket locks a bunch of other mutexes, so we call it outside of proflock.
	// This reduces potential contention and chances of deadlocks.
	// Since the object must be alive during call to mProf_Malloc,
	// it's fine to do this non-atomically.
	_lock.Systemstack(func() {
		setprofilebucket(p, b)
	})
}

func tracealloc(p unsafe.Pointer, size uintptr, typ *_core.Type) {
	_lock.Lock(&_gc.Tracelock)
	gp := _core.Getg()
	gp.M.Traceback = 2
	if typ == nil {
		print("tracealloc(", p, ", ", _core.Hex(size), ")\n")
	} else {
		print("tracealloc(", p, ", ", _core.Hex(size), ", ", *typ.String, ")\n")
	}
	if gp.M.Curg == nil || gp == gp.M.Curg {
		_lock.Goroutineheader(gp)
		pc := _lock.Getcallerpc(unsafe.Pointer(&p))
		sp := _lock.Getcallersp(unsafe.Pointer(&p))
		_lock.Systemstack(func() {
			_lock.Traceback(pc, sp, 0, gp)
		})
	} else {
		_lock.Goroutineheader(gp.M.Curg)
		_lock.Traceback(^uintptr(0), ^uintptr(0), 0, gp.M.Curg)
	}
	print("\n")
	gp.M.Traceback = 0
	_lock.Unlock(&_gc.Tracelock)
}
