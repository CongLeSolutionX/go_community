// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build amd64 amd64p32

package lock

// The calls to nop are to keep these functions from being inlined.
// If they are inlined we have no guarantee that later rewrites of the
// code by optimizers will preserve the relative order of memory accesses.

//go:nosplit
func Atomicload(ptr *uint32) uint32 {
	Nop()
	return *ptr
}

//go:noescape
func Xadd(ptr *uint32, delta int32) uint32

//go:noescape
func Xadd64(ptr *uint64, delta int64) uint64

//go:noescape
func Atomicstore(ptr *uint32, val uint32)
