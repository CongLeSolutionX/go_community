// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Garbage collector: write barriers.
//
// For the concurrent garbage collector, the Go compiler implements
// updates to pointer-valued fields that may be in heap objects by
// emitting calls to write barriers. This file contains the actual write barrier
// implementation, markwb, and the various wrappers called by the
// compiler to implement pointer assignment, slice assignment,
// typed memmove, and so on.

package iface

import (
	_base "runtime/internal/base"
	"unsafe"
)

// NOTE: Really dst *unsafe.Pointer, src unsafe.Pointer,
// but if we do that, Go inserts a write barrier on *dst = src.
//go:nosplit
func Writebarrierptr(dst *uintptr, src uintptr) {
	*dst = src
	if !_base.WriteBarrierEnabled {
		return
	}
	if src != 0 && (src < _base.PhysPageSize || src == _base.PoisonStack) {
		_base.Systemstack(func() {
			print("runtime: writebarrierptr *", dst, " = ", _base.Hex(src), "\n")
			_base.Throw("bad pointer in write barrier")
		})
	}
	_base.Writebarrierptr_nostore1(dst, src)
}

//go:generate go run wbfat_gen.go -- wbfat.go
//
// The above line generates multiword write barriers for
// all the combinations of ptr+scalar up to four words.
// The implementations are written to wbfat.go.

// typedmemmove copies a value of type t to dst from src.
//go:nosplit
func Typedmemmove(typ *_base.Type, dst, src unsafe.Pointer) {
	_base.Memmove(dst, src, typ.Size)
	if typ.Kind&KindNoPointers != 0 {
		return
	}
	HeapBitsBulkBarrier(uintptr(dst), typ.Size)
}
