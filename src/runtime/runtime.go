// Copyright 2009 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import _ "unsafe"

import (
	_base "runtime/internal/base"
	_gc "runtime/internal/gc"
)

//go:generate go run wincallback.go
//go:generate go run mkduff.go

var ticks struct {
	lock _base.Mutex
	pad  uint32 // ensure 8-byte alignment of val on 386
	val  uint64
}

var tls0 [8]uintptr // available storage for m0's TLS; not necessarily used; opaque to GC

// Note: Called by runtime/pprof in addition to runtime code.
func tickspersecond() int64 {
	r := int64(_base.Atomicload64(&ticks.val))
	if r != 0 {
		return r
	}
	_base.Lock(&ticks.lock)
	r = int64(ticks.val)
	if r == 0 {
		t0 := _base.Nanotime()
		c0 := _base.Cputicks()
		_base.Usleep(100 * 1000)
		t1 := _base.Nanotime()
		c1 := _base.Cputicks()
		if t1 == t0 {
			t1++
		}
		r = (c1 - c0) * 1000 * 1000 * 1000 / (t1 - t0)
		if r == 0 {
			r++
		}
		_base.Atomicstore64(&ticks.val, uint64(r))
	}
	_base.Unlock(&ticks.lock)
	return r
}

var argslice []string

//go:linkname syscall_runtime_envs syscall.runtime_envs
func syscall_runtime_envs() []string { return append([]string{}, _gc.Envs...) }

//go:linkname os_runtime_args os.runtime_args
func os_runtime_args() []string { return append([]string{}, argslice...) }
