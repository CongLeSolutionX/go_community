// Copyright 2010 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lock

import (
	"unsafe"
)

//go:nosplit
func SysAlloc(n uintptr, stat *uint64) unsafe.Pointer {
	v := (unsafe.Pointer)(Mmap(nil, n, PROT_READ|PROT_WRITE, MAP_ANON|MAP_PRIVATE, -1, 0))
	if uintptr(v) < 4096 {
		return nil
	}
	Xadd64(stat, int64(n))
	return v
}
