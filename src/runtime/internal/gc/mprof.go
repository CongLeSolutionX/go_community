// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Malloc profiling.
// Patterned after tcmalloc's algorithms; shorter code.

package gc

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	_sem "runtime/internal/sem"
	"unsafe"
)

func Mprof_GC() {
	for b := _sem.Mbuckets; b != nil; b = b.Allnext {
		mp := b.Mp()
		mp.Allocs += mp.Prev_allocs
		mp.Frees += mp.Prev_frees
		mp.Alloc_bytes += mp.Prev_alloc_bytes
		mp.Free_bytes += mp.Prev_free_bytes

		mp.Prev_allocs = mp.Recent_allocs
		mp.Prev_frees = mp.Recent_frees
		mp.Prev_alloc_bytes = mp.Recent_alloc_bytes
		mp.Prev_free_bytes = mp.Recent_free_bytes

		mp.Recent_allocs = 0
		mp.Recent_frees = 0
		mp.Recent_alloc_bytes = 0
		mp.Recent_free_bytes = 0
	}
}

// Record that a gc just happened: all the 'recent' statistics are now real.
func mProf_GC() {
	_lock.Lock(&_sem.Proflock)
	Mprof_GC()
	_lock.Unlock(&_sem.Proflock)
}

// Called when freeing a profiled block.
func mProf_Free(b *_sem.Bucket, size uintptr, freed bool) {
	_lock.Lock(&_sem.Proflock)
	mp := b.Mp()
	if freed {
		mp.Recent_frees++
		mp.Recent_free_bytes += size
	} else {
		mp.Prev_frees++
		mp.Prev_free_bytes += size
	}
	_lock.Unlock(&_sem.Proflock)
}

// Tracing of alloc/free/gc.

var Tracelock _core.Mutex

func tracefree(p unsafe.Pointer, size uintptr) {
	_lock.Lock(&Tracelock)
	gp := _core.Getg()
	gp.M.Traceback = 2
	print("tracefree(", p, ", ", _core.Hex(size), ")\n")
	_lock.Goroutineheader(gp)
	pc := _lock.Getcallerpc(unsafe.Pointer(&p))
	sp := _lock.Getcallersp(unsafe.Pointer(&p))
	_lock.Systemstack(func() {
		_lock.Traceback(pc, sp, 0, gp)
	})
	print("\n")
	gp.M.Traceback = 0
	_lock.Unlock(&Tracelock)
}

func tracegc() {
	_lock.Lock(&Tracelock)
	gp := _core.Getg()
	gp.M.Traceback = 2
	print("tracegc()\n")
	// running on m->g0 stack; show all non-g0 goroutines
	_lock.Tracebackothers(gp)
	print("end tracegc\n")
	print("\n")
	gp.M.Traceback = 0
	_lock.Unlock(&Tracelock)
}
