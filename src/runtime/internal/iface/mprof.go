// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Malloc profiling.
// Patterned after tcmalloc's algorithms; shorter code.

package iface

import (
	_base "runtime/internal/base"
	_gc "runtime/internal/gc"
	_lock "runtime/internal/lock"
	"unsafe"
)

// Called by malloc to record a profiled block.
func mProf_Malloc(p unsafe.Pointer, size uintptr) {
	var stk [_lock.MaxStack]uintptr
	nstk := _base.Callers(4, stk[:])
	_base.Lock(&_lock.Proflock)
	b := _lock.Stkbucket(_lock.MemProfile, size, stk[:nstk], true)
	mp := b.Mp()
	mp.Recent_allocs++
	mp.Recent_alloc_bytes += size
	_base.Unlock(&_lock.Proflock)

	// Setprofilebucket locks a bunch of other mutexes, so we call it outside of proflock.
	// This reduces potential contention and chances of deadlocks.
	// Since the object must be alive during call to mProf_Malloc,
	// it's fine to do this non-atomically.
	_base.Systemstack(func() {
		setprofilebucket(p, b)
	})
}

func tracealloc(p unsafe.Pointer, size uintptr, typ *_base.Type) {
	_base.Lock(&_gc.Tracelock)
	gp := _base.Getg()
	gp.M.Traceback = 2
	if typ == nil {
		print("tracealloc(", p, ", ", _base.Hex(size), ")\n")
	} else {
		print("tracealloc(", p, ", ", _base.Hex(size), ", ", *typ.String, ")\n")
	}
	if gp.M.Curg == nil || gp == gp.M.Curg {
		_base.Goroutineheader(gp)
		pc := _base.Getcallerpc(unsafe.Pointer(&p))
		sp := _base.Getcallersp(unsafe.Pointer(&p))
		_base.Systemstack(func() {
			_base.Traceback(pc, sp, 0, gp)
		})
	} else {
		_base.Goroutineheader(gp.M.Curg)
		_base.Traceback(^uintptr(0), ^uintptr(0), 0, gp.M.Curg)
	}
	print("\n")
	gp.M.Traceback = 0
	_base.Unlock(&_gc.Tracelock)
}
