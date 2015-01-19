// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lock

import (
	_core "runtime/internal/core"
	"unsafe"
)

var persistent struct {
	lock _core.Mutex
	pos  unsafe.Pointer
	end  unsafe.Pointer
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
	persistent.pos = Roundup(persistent.pos, align)
	if uintptr(persistent.pos)+size > uintptr(persistent.end) {
		persistent.pos = SysAlloc(chunk, &Memstats.Other_sys)
		if persistent.pos == nil {
			Unlock(&persistent.lock)
			Throw("runtime: cannot allocate memory")
		}
		persistent.end = _core.Add(persistent.pos, chunk)
	}
	p := persistent.pos
	persistent.pos = _core.Add(persistent.pos, size)
	Unlock(&persistent.lock)

	if stat != &Memstats.Other_sys {
		Xadd64(stat, int64(size))
		Xadd64(&Memstats.Other_sys, -int64(size))
	}
	return p
}
