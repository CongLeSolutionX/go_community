// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gc

import (
	_core "runtime/internal/core"
	_sched "runtime/internal/sched"
	_sem "runtime/internal/sem"
	"unsafe"
)

// Size classes.  Computed and initialized by InitSizes.
//
// SizeToClass(0 <= n <= MaxSmallSize) returns the size class,
//	1 <= sizeclass < NumSizeClasses, for n.
//	Size class 0 is reserved to mean "not small".
//
// class_to_size[i] = largest size in class i
// class_to_allocnpages[i] = number of pages to allocate when
//	making new objects in class i

var Class_to_size [_core.NumSizeClasses]int32
var Class_to_allocnpages [_core.NumSizeClasses]int32

const (
	KindSpecialFinalizer = 1
	KindSpecialProfile   = 2
	// Note: The finalizer special must be first because if we're freeing
	// an object, a finalizer special will cause the freeing operation
	// to abort, and we want to keep the other special records around
	// if that happens.
)

// The described object has a finalizer set for it.
type Specialfinalizer struct {
	Special _core.Special
	Fn      *_core.Funcval
	Nret    uintptr
	Fint    *_core.Type
	Ot      *Ptrtype
}

// The described object is being heap profiled.
type Specialprofile struct {
	Special _core.Special
	B       *_sem.Bucket
}

// NOTE: Layout known to queuefinalizer.
type Finalizer struct {
	Fn   *_core.Funcval // function to call
	Arg  unsafe.Pointer // ptr to object
	Nret uintptr        // bytes of return values from fn
	Fint *_core.Type    // type of first argument of fn
	Ot   *Ptrtype       // type of ptr to object
}

type Finblock struct {
	Alllink *Finblock
	Next    *Finblock
	Cnt     int32
	_       int32
	Fin     [(_sched.FinBlockSize - 2*_core.PtrSize - 2*4) / unsafe.Sizeof(Finalizer{})]Finalizer
}

type Stackmap struct {
	N        int32   // number of bitmaps
	nbit     int32   // number of bits in each bitmap
	bytedata [1]byte // bitmaps, each starting on a 32-bit boundary
}
