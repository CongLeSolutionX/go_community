// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lock

import (
	_core "runtime/internal/core"
	"unsafe"
)

// round n up to a multiple of a.  a must be a power of 2.
func Round(n, a uintptr) uintptr {
	return (n + a - 1) &^ (a - 1)
}

var persistent struct {
	lock _core.Mutex
	base unsafe.Pointer
	off  uintptr
}

// Wrapper around sysAlloc that can allocate small chunks.
// There is no associated free operation.
// Intended for things like function/type/debug-related persistent data.
// If align is 0, uses default align (currently 8).
func Persistentalloc(size, align uintptr, stat *uint64) unsafe.Pointer {
	const (
		chunk    = 256 << 10
		maxBlock = 64 << 10 // VM reservation granularity is 64K on windows
	)

	if size == 0 {
		Throw("persistentalloc: size == 0")
	}
	if align != 0 {
		if align&(align-1) != 0 {
			Throw("persistentalloc: align is not a power of 2")
		}
		if align > _core.PageSize {
			Throw("persistentalloc: align is too large")
		}
	} else {
		align = 8
	}

	if size >= maxBlock {
		return SysAlloc(size, stat)
	}

	Lock(&persistent.lock)
	persistent.off = Round(persistent.off, align)
	if persistent.off+size > chunk || persistent.base == nil {
		persistent.base = SysAlloc(chunk, &Memstats.Other_sys)
		if persistent.base == nil {
			Unlock(&persistent.lock)
			Throw("runtime: cannot allocate memory")
		}
		persistent.off = 0
	}
	p := _core.Add(persistent.base, persistent.off)
	persistent.off += size
	Unlock(&persistent.lock)

	if stat != &Memstats.Other_sys {
		Xadd64(stat, int64(size))
		Xadd64(&Memstats.Other_sys, -int64(size))
	}
	return p
}
