// Copyright 2009 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Garbage collector: finalizers and block profiling.

package gc

import (
	_base "runtime/internal/base"
	"unsafe"
)

type Finblock struct {
	Alllink *Finblock
	Next    *Finblock
	Cnt     int32
	_       int32
	Fin     [(_base.FinBlockSize - 2*_base.PtrSize - 2*4) / unsafe.Sizeof(Finalizer{})]Finalizer
}

var Finq *Finblock // list of finalizers that are to be executed
var Finc *Finblock // cache of free blocks
var finptrmask [_base.FinBlockSize / _base.PtrSize / 8]byte
var Allfin *Finblock // list of all blocks

// NOTE: Layout known to queuefinalizer.
type Finalizer struct {
	Fn   *_base.Funcval // function to call
	Arg  unsafe.Pointer // ptr to object
	Nret uintptr        // bytes of return values from fn
	Fint *_base.Type    // type of first argument of fn
	Ot   *Ptrtype       // type of ptr to object
}

var finalizer1 = [...]byte{
	// Each Finalizer is 5 words, ptr ptr INT ptr ptr (INT = uintptr here)
	// Each byte describes 8 words.
	// Need 8 Finalizers described by 5 bytes before pattern repeats:
	//	ptr ptr INT ptr ptr
	//	ptr ptr INT ptr ptr
	//	ptr ptr INT ptr ptr
	//	ptr ptr INT ptr ptr
	//	ptr ptr INT ptr ptr
	//	ptr ptr INT ptr ptr
	//	ptr ptr INT ptr ptr
	//	ptr ptr INT ptr ptr
	// aka
	//
	//	ptr ptr INT ptr ptr ptr ptr INT
	//	ptr ptr ptr ptr INT ptr ptr ptr
	//	ptr INT ptr ptr ptr ptr INT ptr
	//	ptr ptr ptr INT ptr ptr ptr ptr
	//	INT ptr ptr ptr ptr INT ptr ptr
	//
	// Assumptions about Finalizer layout checked below.
	1<<0 | 1<<1 | 0<<2 | 1<<3 | 1<<4 | 1<<5 | 1<<6 | 0<<7,
	1<<0 | 1<<1 | 1<<2 | 1<<3 | 0<<4 | 1<<5 | 1<<6 | 1<<7,
	1<<0 | 0<<1 | 1<<2 | 1<<3 | 1<<4 | 1<<5 | 0<<6 | 1<<7,
	1<<0 | 1<<1 | 1<<2 | 0<<3 | 1<<4 | 1<<5 | 1<<6 | 1<<7,
	0<<0 | 1<<1 | 1<<2 | 1<<3 | 1<<4 | 0<<5 | 1<<6 | 1<<7,
}

func queuefinalizer(p unsafe.Pointer, fn *_base.Funcval, nret uintptr, fint *_base.Type, ot *Ptrtype) {
	_base.Lock(&_base.Finlock)
	if Finq == nil || Finq.Cnt == int32(len(Finq.Fin)) {
		if Finc == nil {
			// Note: write barrier here, assigning to finc, but should be okay.
			Finc = (*Finblock)(_base.Persistentalloc(_base.FinBlockSize, 0, &_base.Memstats.Gc_sys))
			Finc.Alllink = Allfin
			Allfin = Finc
			if finptrmask[0] == 0 {
				// Build pointer mask for Finalizer array in block.
				// Check assumptions made in finalizer1 array above.
				if (unsafe.Sizeof(Finalizer{}) != 5*_base.PtrSize ||
					unsafe.Offsetof(Finalizer{}.Fn) != 0 ||
					unsafe.Offsetof(Finalizer{}.Arg) != _base.PtrSize ||
					unsafe.Offsetof(Finalizer{}.Nret) != 2*_base.PtrSize ||
					unsafe.Offsetof(Finalizer{}.Fint) != 3*_base.PtrSize ||
					unsafe.Offsetof(Finalizer{}.Ot) != 4*_base.PtrSize) {
					_base.Throw("finalizer out of sync")
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
	_base.Fingwake = true
	_base.Unlock(&_base.Finlock)
}
