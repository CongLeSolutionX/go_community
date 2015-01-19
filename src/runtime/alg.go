// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	_core "runtime/internal/core"
	_sched "runtime/internal/sched"
	"unsafe"
)

// Testing adapter for memclr
func memclrBytes(b []byte) {
	s := (*_sched.SliceStruct)(unsafe.Pointer(&b))
	_core.Memclr(s.Array, uintptr(s.Len))
}
