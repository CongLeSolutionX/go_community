// Copyright 2009 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Garbage collector: finalizers and block profiling.

package gc

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
	"unsafe"
)

var Finq *Finblock // list of finalizers that are to be executed
var Finc *Finblock // cache of free blocks
var finptrmask [_sched.FinBlockSize / _sched.TypeBitmapScale]byte
var Allfin *Finblock // list of all blocks

var finalizer1 = [...]byte{
	// Each Finalizer is 5 words, ptr ptr uintptr ptr ptr.
	// Each byte describes 4 words.
	// Need 4 Finalizers described by 5 bytes before pattern repeats:
	//	ptr ptr uintptr ptr ptr
	//	ptr ptr uintptr ptr ptr
	//	ptr ptr uintptr ptr ptr
	//	ptr ptr uintptr ptr ptr
	// aka
	//	ptr ptr uintptr ptr
	//	ptr ptr ptr uintptr
	//	ptr ptr ptr ptr
	//	uintptr ptr ptr ptr
	//	ptr uintptr ptr ptr
	// Assumptions about Finalizer layout checked below.
	_sched.TypePointer | _sched.TypePointer<<2 | _sched.TypeScalar<<4 | _sched.TypePointer<<6,
	_sched.TypePointer | _sched.TypePointer<<2 | _sched.TypePointer<<4 | _sched.TypeScalar<<6,
	_sched.TypePointer | _sched.TypePointer<<2 | _sched.TypePointer<<4 | _sched.TypePointer<<6,
	_sched.TypeScalar | _sched.TypePointer<<2 | _sched.TypePointer<<4 | _sched.TypePointer<<6,
	_sched.TypePointer | _sched.TypeScalar<<2 | _sched.TypePointer<<4 | _sched.TypePointer<<6,
}

func queuefinalizer(p unsafe.Pointer, fn *_core.Funcval, nret uintptr, fint *_core.Type, ot *Ptrtype) {
	_lock.Lock(&_sched.Finlock)
	if Finq == nil || Finq.Cnt == int32(len(Finq.Fin)) {
		if Finc == nil {
			// Note: write barrier here, assigning to finc, but should be okay.
			Finc = (*Finblock)(_lock.Persistentalloc(_sched.FinBlockSize, 0, &_lock.Memstats.Gc_sys))
			Finc.Alllink = Allfin
			Allfin = Finc
			if finptrmask[0] == 0 {
				// Build pointer mask for Finalizer array in block.
				// Check assumptions made in finalizer1 array above.
				if (unsafe.Sizeof(Finalizer{}) != 5*_core.PtrSize ||
					unsafe.Offsetof(Finalizer{}.Fn) != 0 ||
					unsafe.Offsetof(Finalizer{}.Arg) != _core.PtrSize ||
					unsafe.Offsetof(Finalizer{}.Nret) != 2*_core.PtrSize ||
					unsafe.Offsetof(Finalizer{}.Fint) != 3*_core.PtrSize ||
					unsafe.Offsetof(Finalizer{}.Ot) != 4*_core.PtrSize ||
					_sched.TypeBitsWidth != 2) {
					_lock.Throw("finalizer out of sync")
				}
				for i := range finptrmask {
					finptrmask[i] = finalizer1[i%len(finalizer1)]
				}
			}
		}
		block := Finc
		Finc = block.Next
		block.Next = Finq
		Finq = block
	}
	f := &Finq.Fin[Finq.Cnt]
	Finq.Cnt++
	f.Fn = fn
	f.Nret = nret
	f.Fint = fint
	f.Ot = ot
	f.Arg = p
	_sched.Fingwake = true
	_lock.Unlock(&_sched.Finlock)
}
