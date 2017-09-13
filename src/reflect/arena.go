// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package reflect

import (
	"unsafe"
)

type Arena struct {
	ptr  unsafe.Pointer // allocation point
	size uintptr        // remaining size
}

type arenaStore struct {
	// This type must contain at least one pointer so the arena
	// allocation doesn't happen in a noscan span.
	// As soon as we allocate the arena, we will erase its pointer bit.
	_ unsafe.Pointer
}

var (
	arenaStoreIface = interface{}(arenaStore{})
	arenaStoreType  = (*emptyInterface)(unsafe.Pointer(&arenaStoreIface)).typ
	uintptrIface    = interface{}(uintptr(0))
	uintptrType     = (*emptyInterface)(unsafe.Pointer(&uintptrIface)).typ
)

func NewArena(bytes int) *Arena {
	if bytes <= 0 {
		return new(Arena)
	}
	b := uintptr(bytes)
	if b < unsafe.Sizeof(arenaStore{}) {
		b = unsafe.Sizeof(arenaStore{})
	}

	// TODO: Might as well round up to size class.

	typ := *arenaStoreType
	typ.size = b
	base := unsafe_New(&typ)
	// Overwrite the single ptr bit.
	// We need to clobber the pointer bit in case the entire arena is
	// filled with pointerless allocations.
	unsafe_NewAt(uintptrType, base)

	// TODO: Arenas are allocated in the heap. Return them by value instead?
	// Then we'd have to return updated arenas on every operation.
	return &Arena{ptr: base, size: b}
}

var zeroBase byte

func (a *Arena) reservePointer(size uintptr, align uint8) unsafe.Pointer {
	// TODO: Atomic to make this thread-safe. (Otherwise this is
	// an easy way to escape memory safety, even though
	// technically we could call it a race.)
	p := a.ptr
	s := a.size

	// Allocate from the beginning.
	pad := (-uintptr(p)) & (uintptr(align) - 1) // alignment padding
	size += pad
	if size > s {
		return nil
	}
	a.size = s - size
	if a.size != 0 {
		a.ptr = add(p, size, "remaining size > 0")
	} else {
		// Prevent pointer-to-next-object-in-heap.
		a.ptr = nil
	}
	return add(p, pad, "size > 0")
}

func (a *Arena) reserveScalar(size uintptr, align uint8) unsafe.Pointer {
	if size == 0 {
		return unsafe.Pointer(&zeroBase)
	}
	p := a.ptr
	s := a.size
	// Allocate from the end.
	pad := (uintptr(p) + s) & (uintptr(align) - 1) // alignment padding
	size += pad
	if size > s {
		return nil
	}
	a.size = s - size
	return add(p, s-size, "size > 0")
}

// New allocates a value of type typ and returns a pointer to it.
// If there isn't enough space left in the arena, returns nil.
func (a *Arena) New(typ Type) interface{} {
	if typ == nil {
		panic("reflect: Arena.New(nil)")
	}
	rt := typ.(*rtype)
	var ptr unsafe.Pointer
	if rt.pointers() {
		ptr = a.reservePointer(rt.size, rt.align)
		if ptr != nil {
			unsafe_NewAt(rt, ptr)
		}
	} else {
		ptr = a.reserveScalar(rt.size, rt.align)
	}
	if ptr == nil {
		return nil
	}
	i := emptyInterface{
		typ:  rt.ptrTo(),
		word: ptr,
	}
	return *(*interface{})(unsafe.Pointer(&i))
}

// res must be *[]T.  Allocates an array a = [cap]T, assigns *res = a[:].
// The weird argument convention is so the storage for the slice header
// can be allocated in the caller's stack frame.
// If the array won't fit in the arena, returns without allocating or assigning anything.
func (a *Arena) Slice(cap int, res interface{}) {
	if cap < 0 {
		panic("reflect.Arena.MakeSlice: negative cap")
	}
	i := (*emptyInterface)(unsafe.Pointer(&res))
	t := i.typ
	if t.Kind() != Ptr {
		panic("reflect.Arena.Slice result of non-ptr type")
	}
	t = (*ptrType)(unsafe.Pointer(t)).elem
	if t.Kind() != Slice {
		panic("reflect.Arena.Slice of non-ptr-to-slice type")
	}
	t = (*sliceType)(unsafe.Pointer(t)).elem
	// t is now the element type of the slice we want to allocate.

	size := t.size * uintptr(cap) // TODO: Check for overflow

	var ptr unsafe.Pointer
	if t.pointers() {
		ptr = a.reservePointer(size, t.align)
		if ptr != nil {
			unsafe_NewArrayAt(t, cap, ptr)
		}
	} else {
		ptr = a.reserveScalar(size, t.align)
	}
	if ptr != nil {
		*(*sliceHeader)(i.word) = sliceHeader{Data: ptr, Len: cap, Cap: cap}
	}
}

// implemented in package runtime
func unsafe_NewAt(*rtype, unsafe.Pointer)
func unsafe_NewArrayAt(*rtype, int, unsafe.Pointer)
