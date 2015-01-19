// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package check

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
	"unsafe"
)

func testAtomic64() {
	var z64, x64 uint64

	z64 = 42
	x64 = 0
	prefetcht0(uintptr(unsafe.Pointer(&z64)))
	prefetcht1(uintptr(unsafe.Pointer(&z64)))
	prefetcht2(uintptr(unsafe.Pointer(&z64)))
	prefetchnta(uintptr(unsafe.Pointer(&z64)))
	if _sched.Cas64(&z64, x64, 1) {
		_lock.Throw("cas64 failed")
	}
	if x64 != 0 {
		_lock.Throw("cas64 failed")
	}
	x64 = 42
	if !_sched.Cas64(&z64, x64, 1) {
		_lock.Throw("cas64 failed")
	}
	if x64 != 42 || z64 != 1 {
		_lock.Throw("cas64 failed")
	}
	if _sched.Atomicload64(&z64) != 1 {
		_lock.Throw("load64 failed")
	}
	_sched.Atomicstore64(&z64, (1<<40)+1)
	if _sched.Atomicload64(&z64) != (1<<40)+1 {
		_lock.Throw("store64 failed")
	}
	if _lock.Xadd64(&z64, (1<<40)+1) != (2<<40)+2 {
		_lock.Throw("xadd64 failed")
	}
	if _sched.Atomicload64(&z64) != (2<<40)+2 {
		_lock.Throw("xadd64 failed")
	}
	if _sched.Xchg64(&z64, (3<<40)+3) != (2<<40)+2 {
		_lock.Throw("xchg64 failed")
	}
	if _sched.Atomicload64(&z64) != (3<<40)+3 {
		_lock.Throw("xchg64 failed")
	}
}

func check() {
	var (
		a     int8
		b     uint8
		c     int16
		d     uint16
		e     int32
		f     uint32
		g     int64
		h     uint64
		i, i1 float32
		j, j1 float64
		k, k1 unsafe.Pointer
		l     *uint16
		m     [4]byte
	)
	type x1t struct {
		x uint8
	}
	type y1t struct {
		x1 x1t
		y  uint8
	}
	var x1 x1t
	var y1 y1t

	if unsafe.Sizeof(a) != 1 {
		_lock.Throw("bad a")
	}
	if unsafe.Sizeof(b) != 1 {
		_lock.Throw("bad b")
	}
	if unsafe.Sizeof(c) != 2 {
		_lock.Throw("bad c")
	}
	if unsafe.Sizeof(d) != 2 {
		_lock.Throw("bad d")
	}
	if unsafe.Sizeof(e) != 4 {
		_lock.Throw("bad e")
	}
	if unsafe.Sizeof(f) != 4 {
		_lock.Throw("bad f")
	}
	if unsafe.Sizeof(g) != 8 {
		_lock.Throw("bad g")
	}
	if unsafe.Sizeof(h) != 8 {
		_lock.Throw("bad h")
	}
	if unsafe.Sizeof(i) != 4 {
		_lock.Throw("bad i")
	}
	if unsafe.Sizeof(j) != 8 {
		_lock.Throw("bad j")
	}
	if unsafe.Sizeof(k) != _core.PtrSize {
		_lock.Throw("bad k")
	}
	if unsafe.Sizeof(l) != _core.PtrSize {
		_lock.Throw("bad l")
	}
	if unsafe.Sizeof(x1) != 1 {
		_lock.Throw("bad unsafe.Sizeof x1")
	}
	if unsafe.Offsetof(y1.y) != 1 {
		_lock.Throw("bad offsetof y1.y")
	}
	if unsafe.Sizeof(y1) != 2 {
		_lock.Throw("bad unsafe.Sizeof y1")
	}

	if _lock.Timediv(12345*1000000000+54321, 1000000000, &e) != 12345 || e != 54321 {
		_lock.Throw("bad timediv")
	}

	var z uint32
	z = 1
	if !_sched.Cas(&z, 1, 2) {
		_lock.Throw("cas1")
	}
	if z != 2 {
		_lock.Throw("cas2")
	}

	z = 4
	if _sched.Cas(&z, 5, 6) {
		_lock.Throw("cas3")
	}
	if z != 4 {
		_lock.Throw("cas4")
	}

	z = 0xffffffff
	if !_sched.Cas(&z, 0xffffffff, 0xfffffffe) {
		_lock.Throw("cas5")
	}
	if z != 0xfffffffe {
		_lock.Throw("cas6")
	}

	k = unsafe.Pointer(uintptr(0xfedcb123))
	if _core.PtrSize == 8 {
		k = unsafe.Pointer(uintptr(unsafe.Pointer(k)) << 10)
	}
	if casp(&k, nil, nil) {
		_lock.Throw("casp1")
	}
	k1 = _core.Add(k, 1)
	if !casp(&k, k, k1) {
		_lock.Throw("casp2")
	}
	if k != k1 {
		_lock.Throw("casp3")
	}

	m = [4]byte{1, 1, 1, 1}
	_sched.Atomicor8(&m[1], 0xf0)
	if m[0] != 1 || m[1] != 0xf1 || m[2] != 1 || m[3] != 1 {
		_lock.Throw("atomicor8")
	}

	*(*uint64)(unsafe.Pointer(&j)) = ^uint64(0)
	if j == j {
		_lock.Throw("float64nan")
	}
	if !(j != j) {
		_lock.Throw("float64nan1")
	}

	*(*uint64)(unsafe.Pointer(&j1)) = ^uint64(1)
	if j == j1 {
		_lock.Throw("float64nan2")
	}
	if !(j != j1) {
		_lock.Throw("float64nan3")
	}

	*(*uint32)(unsafe.Pointer(&i)) = ^uint32(0)
	if i == i {
		_lock.Throw("float32nan")
	}
	if i == i {
		_lock.Throw("float32nan1")
	}

	*(*uint32)(unsafe.Pointer(&i1)) = ^uint32(1)
	if i == i1 {
		_lock.Throw("float32nan2")
	}
	if i == i1 {
		_lock.Throw("float32nan3")
	}

	testAtomic64()

	if _core.FixedStack != _sched.Round2(_core.FixedStack) {
		_lock.Throw("FixedStack is not power-of-2")
	}
}
