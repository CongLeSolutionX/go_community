// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Malloc profiling.
// Patterned after tcmalloc's algorithms; shorter code.

package gc

import (
	_base "runtime/internal/base"
	_lock "runtime/internal/lock"
	"unsafe"
)

func Mprof_GC() {
	for b := _lock.Mbuckets; b != nil; b = b.Allnext {
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
	_base.Lock(&_lock.Proflock)
	Mprof_GC()
	_base.Unlock(&_lock.Proflock)
}

// Called when freeing a profiled block.
func mProf_Free(b *_lock.Bucket, size uintptr, freed bool) {
	_base.Lock(&_lock.Proflock)
	mp := b.Mp()
	if freed {
		mp.Recent_frees++
		mp.Recent_free_bytes += size
	} else {
		mp.Prev_frees++
		mp.Prev_free_bytes += size
	}
	_base.Unlock(&_lock.Proflock)
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
