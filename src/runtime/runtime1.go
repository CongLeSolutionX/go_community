// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	_base "runtime/internal/base"
	_gc "runtime/internal/gc"
	"unsafe"
)

var (
	argc int32
	argv **byte
)

// nosplit for use in linux/386 startup linux_setup_vdso
//go:nosplit
func argv_index(argv **byte, i int32) *byte {
	return *(**byte)(_base.Add(unsafe.Pointer(argv), uintptr(i)*_base.PtrSize))
}

func args(c int32, v **byte) {
	argc = c
	argv = v
	sysargs(c, v)
}

var (
	// TODO: Retire in favor of GOOS== checks.
	isplan9   int32
	issolaris int32
	iswindows int32
)

func goargs() {
	if _base.GOOS == "windows" {
		return
	}

	argslice = make([]string, argc)
	for i := int32(0); i < argc; i++ {
		argslice[i] = _base.Gostringnocopy(argv_index(argv, i))
	}
}

func goenvs_unix() {
	// TODO(austin): ppc64 in dynamic linking mode doesn't
	// guarantee env[] will immediately follow argv.  Might cause
	// problems.
	n := int32(0)
	for argv_index(argv, argc+1+n) != nil {
		n++
	}

	_gc.Envs = make([]string, n)
	for i := int32(0); i < n; i++ {
		_gc.Envs[i] = gostring(argv_index(argv, argc+1+i))
	}
}

// TODO: These should be locals in testAtomic64, but we don't 8-byte
// align stack variables on 386.
var test_z64, test_x64 uint64

func testAtomic64() {
	test_z64 = 42
	test_x64 = 0
	prefetcht0(uintptr(unsafe.Pointer(&test_z64)))
	prefetcht1(uintptr(unsafe.Pointer(&test_z64)))
	prefetcht2(uintptr(unsafe.Pointer(&test_z64)))
	_base.Prefetchnta(uintptr(unsafe.Pointer(&test_z64)))
	if _base.Cas64(&test_z64, test_x64, 1) {
		_base.Throw("cas64 failed")
	}
	if test_x64 != 0 {
		_base.Throw("cas64 failed")
	}
	test_x64 = 42
	if !_base.Cas64(&test_z64, test_x64, 1) {
		_base.Throw("cas64 failed")
	}
	if test_x64 != 42 || test_z64 != 1 {
		_base.Throw("cas64 failed")
	}
	if _base.Atomicload64(&test_z64) != 1 {
		_base.Throw("load64 failed")
	}
	_base.Atomicstore64(&test_z64, (1<<40)+1)
	if _base.Atomicload64(&test_z64) != (1<<40)+1 {
		_base.Throw("store64 failed")
	}
	if _base.Xadd64(&test_z64, (1<<40)+1) != (2<<40)+2 {
		_base.Throw("xadd64 failed")
	}
	if _base.Atomicload64(&test_z64) != (2<<40)+2 {
		_base.Throw("xadd64 failed")
	}
	if _base.Xchg64(&test_z64, (3<<40)+3) != (2<<40)+2 {
		_base.Throw("xchg64 failed")
	}
	if _base.Atomicload64(&test_z64) != (3<<40)+3 {
		_base.Throw("xchg64 failed")
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
		_base.Throw("bad a")
	}
	if unsafe.Sizeof(b) != 1 {
		_base.Throw("bad b")
	}
	if unsafe.Sizeof(c) != 2 {
		_base.Throw("bad c")
	}
	if unsafe.Sizeof(d) != 2 {
		_base.Throw("bad d")
	}
	if unsafe.Sizeof(e) != 4 {
		_base.Throw("bad e")
	}
	if unsafe.Sizeof(f) != 4 {
		_base.Throw("bad f")
	}
	if unsafe.Sizeof(g) != 8 {
		_base.Throw("bad g")
	}
	if unsafe.Sizeof(h) != 8 {
		_base.Throw("bad h")
	}
	if unsafe.Sizeof(i) != 4 {
		_base.Throw("bad i")
	}
	if unsafe.Sizeof(j) != 8 {
		_base.Throw("bad j")
	}
	if unsafe.Sizeof(k) != _base.PtrSize {
		_base.Throw("bad k")
	}
	if unsafe.Sizeof(l) != _base.PtrSize {
		_base.Throw("bad l")
	}
	if unsafe.Sizeof(x1) != 1 {
		_base.Throw("bad unsafe.Sizeof x1")
	}
	if unsafe.Offsetof(y1.y) != 1 {
		_base.Throw("bad offsetof y1.y")
	}
	if unsafe.Sizeof(y1) != 2 {
		_base.Throw("bad unsafe.Sizeof y1")
	}

	if _base.Timediv(12345*1000000000+54321, 1000000000, &e) != 12345 || e != 54321 {
		_base.Throw("bad timediv")
	}

	var z uint32
	z = 1
	if !_base.Cas(&z, 1, 2) {
		_base.Throw("cas1")
	}
	if z != 2 {
		_base.Throw("cas2")
	}

	z = 4
	if _base.Cas(&z, 5, 6) {
		_base.Throw("cas3")
	}
	if z != 4 {
		_base.Throw("cas4")
	}

	z = 0xffffffff
	if !_base.Cas(&z, 0xffffffff, 0xfffffffe) {
		_base.Throw("cas5")
	}
	if z != 0xfffffffe {
		_base.Throw("cas6")
	}

	k = unsafe.Pointer(uintptr(0xfedcb123))
	if _base.PtrSize == 8 {
		k = unsafe.Pointer(uintptr(unsafe.Pointer(k)) << 10)
	}
	if casp(&k, nil, nil) {
		_base.Throw("casp1")
	}
	k1 = _base.Add(k, 1)
	if !casp(&k, k, k1) {
		_base.Throw("casp2")
	}
	if k != k1 {
		_base.Throw("casp3")
	}

	m = [4]byte{1, 1, 1, 1}
	_base.Atomicor8(&m[1], 0xf0)
	if m[0] != 1 || m[1] != 0xf1 || m[2] != 1 || m[3] != 1 {
		_base.Throw("atomicor8")
	}

	*(*uint64)(unsafe.Pointer(&j)) = ^uint64(0)
	if j == j {
		_base.Throw("float64nan")
	}
	if !(j != j) {
		_base.Throw("float64nan1")
	}

	*(*uint64)(unsafe.Pointer(&j1)) = ^uint64(1)
	if j == j1 {
		_base.Throw("float64nan2")
	}
	if !(j != j1) {
		_base.Throw("float64nan3")
	}

	*(*uint32)(unsafe.Pointer(&i)) = ^uint32(0)
	if i == i {
		_base.Throw("float32nan")
	}
	if i == i {
		_base.Throw("float32nan1")
	}

	*(*uint32)(unsafe.Pointer(&i1)) = ^uint32(1)
	if i == i1 {
		_base.Throw("float32nan2")
	}
	if i == i1 {
		_base.Throw("float32nan3")
	}

	testAtomic64()

	if _base.FixedStack != _base.Round2(_base.FixedStack) {
		_base.Throw("FixedStack is not power-of-2")
	}
}

type dbgVar struct {
	name  string
	value *int32
}

var dbgvars = []dbgVar{
	{"allocfreetrace", &_base.Debug.Allocfreetrace},
	{"efence", &_base.Debug.Efence},
	{"gccheckmark", &_base.Debug.Gccheckmark},
	{"gcpacertrace", &_base.Debug.Gcpacertrace},
	{"gcshrinkstackoff", &_base.Debug.Gcshrinkstackoff},
	{"gcstackbarrieroff", &_base.Debug.Gcstackbarrieroff},
	{"gcstoptheworld", &_base.Debug.Gcstoptheworld},
	{"gctrace", &_base.Debug.Gctrace},
	{"invalidptr", &_base.Debug.Invalidptr},
	{"sbrk", &_base.Debug.Sbrk},
	{"scavenge", &_base.Debug.Scavenge},
	{"scheddetail", &_base.Debug.Scheddetail},
	{"schedtrace", &_base.Debug.Schedtrace},
	{"wbshadow", &_base.Debug.Wbshadow},
}

func parsedebugvars() {
	for p := _gc.Gogetenv("GODEBUG"); p != ""; {
		field := ""
		i := _base.Index(p, ",")
		if i < 0 {
			field, p = p, ""
		} else {
			field, p = p[:i], p[i+1:]
		}
		i = _base.Index(field, "=")
		if i < 0 {
			continue
		}
		key, value := field[:i], field[i+1:]

		// Update MemProfileRate directly here since it
		// is int, not int32, and should only be updated
		// if specified in GODEBUG.
		if key == "memprofilerate" {
			_base.MemProfileRate = _gc.Atoi(value)
		} else {
			for _, v := range dbgvars {
				if v.name == key {
					*v.value = int32(_gc.Atoi(value))
				}
			}
		}
	}

	switch p := _gc.Gogetenv("GOTRACEBACK"); p {
	case "":
		_base.Traceback_cache = 1 << 1
	case "crash":
		_base.Traceback_cache = 2<<1 | 1
	default:
		_base.Traceback_cache = uint32(_gc.Atoi(p)) << 1
	}
	// when C owns the process, simply exit'ing the process on fatal errors
	// and panics is surprising. Be louder and abort instead.
	if _base.Islibrary || _base.Isarchive {
		_base.Traceback_cache |= 1
	}
}

//go:linkname reflect_typelinks reflect.typelinks
//go:nosplit
func reflect_typelinks() [][]*_base.Type {
	ret := [][]*_base.Type{_base.Firstmoduledata.Typelinks}
	for datap := _base.Firstmoduledata.Next; datap != nil; datap = datap.Next {
		ret = append(ret, datap.Typelinks)
	}
	return ret
}
