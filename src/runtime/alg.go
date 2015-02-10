// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	_core "runtime/internal/core"
	_sched "runtime/internal/sched"
	"unsafe"
)

// memhash_varlen is defined in assembly because it needs access
// to the closure.  It appears here to provide an argument
// signature for the assembly routine.
func memhash_varlen(p unsafe.Pointer, h uintptr) uintptr

// Testing adapter for memclr
func memclrBytes(b []byte) {
	s := (*_sched.SliceStruct)(unsafe.Pointer(&b))
	_core.Memclr(s.Array, uintptr(s.Len))
}
