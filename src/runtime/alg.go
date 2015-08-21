// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	_base "runtime/internal/base"
	_iface "runtime/internal/iface"
	"unsafe"
)

const (
	c0 = uintptr((8-_base.PtrSize)/4*2860486313 + (_base.PtrSize-4)/4*33054211828000289)
	c1 = uintptr((8-_base.PtrSize)/4*3267000013 + (_base.PtrSize-4)/4*23344194077549503)
)

// type algorithms - known to compiler
const (
	alg_MEM = iota
	alg_MEM0
	alg_MEM8
	alg_MEM16
	alg_MEM32
	alg_MEM64
	alg_MEM128
	alg_NOEQ
	alg_NOEQ0
	alg_NOEQ8
	alg_NOEQ16
	alg_NOEQ32
	alg_NOEQ64
	alg_NOEQ128
	alg_STRING
	alg_INTER
	alg_NILINTER
	alg_SLICE
	alg_FLOAT32
	alg_FLOAT64
	alg_CPLX64
	alg_CPLX128
	alg_max
)

func memhash0(p unsafe.Pointer, h uintptr) uintptr {
	return h
}
func memhash8(p unsafe.Pointer, h uintptr) uintptr {
	return _base.Memhash(p, h, 1)
}
func memhash16(p unsafe.Pointer, h uintptr) uintptr {
	return _base.Memhash(p, h, 2)
}
func memhash32(p unsafe.Pointer, h uintptr) uintptr {
	return _base.Memhash(p, h, 4)
}
func memhash64(p unsafe.Pointer, h uintptr) uintptr {
	return _base.Memhash(p, h, 8)
}
func memhash128(p unsafe.Pointer, h uintptr) uintptr {
	return _base.Memhash(p, h, 16)
}

// memhash_varlen is defined in assembly because it needs access
// to the closure.  It appears here to provide an argument
// signature for the assembly routine.
func memhash_varlen(p unsafe.Pointer, h uintptr) uintptr

var algarray = [alg_max]_base.TypeAlg{
	alg_MEM:      {nil, nil}, // not used
	alg_MEM0:     {memhash0, memequal0},
	alg_MEM8:     {memhash8, memequal8},
	alg_MEM16:    {memhash16, memequal16},
	alg_MEM32:    {memhash32, memequal32},
	alg_MEM64:    {memhash64, memequal64},
	alg_MEM128:   {memhash128, memequal128},
	alg_NOEQ:     {nil, nil},
	alg_NOEQ0:    {nil, nil},
	alg_NOEQ8:    {nil, nil},
	alg_NOEQ16:   {nil, nil},
	alg_NOEQ32:   {nil, nil},
	alg_NOEQ64:   {nil, nil},
	alg_NOEQ128:  {nil, nil},
	alg_STRING:   {strhash, strequal},
	alg_INTER:    {interhash, interequal},
	alg_NILINTER: {nilinterhash, nilinterequal},
	alg_SLICE:    {nil, nil},
	alg_FLOAT32:  {f32hash, f32equal},
	alg_FLOAT64:  {f64hash, f64equal},
	alg_CPLX64:   {c64hash, c64equal},
	alg_CPLX128:  {c128hash, c128equal},
}

func aeshash32(p unsafe.Pointer, h uintptr) uintptr
func aeshash64(p unsafe.Pointer, h uintptr) uintptr
func aeshashstr(p unsafe.Pointer, h uintptr) uintptr

func strhash(a unsafe.Pointer, h uintptr) uintptr {
	x := (*_base.StringStruct)(a)
	return _base.Memhash(x.Str, h, uintptr(x.Len))
}

// NOTE: Because NaN != NaN, a map can contain any
// number of (mostly useless) entries keyed with NaNs.
// To avoid long hash chains, we assign a random number
// as the hash value for a NaN.

func f32hash(p unsafe.Pointer, h uintptr) uintptr {
	f := *(*float32)(p)
	switch {
	case f == 0:
		return c1 * (c0 ^ h) // +0, -0
	case f != f:
		return c1 * (c0 ^ h ^ uintptr(_base.Fastrand1())) // any kind of NaN
	default:
		return _base.Memhash(p, h, 4)
	}
}

func f64hash(p unsafe.Pointer, h uintptr) uintptr {
	f := *(*float64)(p)
	switch {
	case f == 0:
		return c1 * (c0 ^ h) // +0, -0
	case f != f:
		return c1 * (c0 ^ h ^ uintptr(_base.Fastrand1())) // any kind of NaN
	default:
		return _base.Memhash(p, h, 8)
	}
}

func c64hash(p unsafe.Pointer, h uintptr) uintptr {
	x := (*[2]float32)(p)
	return f32hash(unsafe.Pointer(&x[1]), f32hash(unsafe.Pointer(&x[0]), h))
}

func c128hash(p unsafe.Pointer, h uintptr) uintptr {
	x := (*[2]float64)(p)
	return f64hash(unsafe.Pointer(&x[1]), f64hash(unsafe.Pointer(&x[0]), h))
}

func interhash(p unsafe.Pointer, h uintptr) uintptr {
	a := (*_iface.Iface)(p)
	tab := a.Tab
	if tab == nil {
		return h
	}
	t := tab.Type
	fn := t.Alg.Hash
	if fn == nil {
		panic(_base.ErrorString("hash of unhashable type " + *t.String))
	}
	if _iface.IsDirectIface(t) {
		return c1 * fn(unsafe.Pointer(&a.Data), h^c0)
	} else {
		return c1 * fn(a.Data, h^c0)
	}
}

func nilinterhash(p unsafe.Pointer, h uintptr) uintptr {
	a := (*_iface.Eface)(p)
	t := a.Type
	if t == nil {
		return h
	}
	fn := t.Alg.Hash
	if fn == nil {
		panic(_base.ErrorString("hash of unhashable type " + *t.String))
	}
	if _iface.IsDirectIface(t) {
		return c1 * fn(unsafe.Pointer(&a.Data), h^c0)
	} else {
		return c1 * fn(a.Data, h^c0)
	}
}

func memequal(p, q unsafe.Pointer, size uintptr) bool {
	if p == q {
		return true
	}
	return memeq(p, q, size)
}

func memequal0(p, q unsafe.Pointer) bool {
	return true
}
func memequal8(p, q unsafe.Pointer) bool {
	return *(*int8)(p) == *(*int8)(q)
}
func memequal16(p, q unsafe.Pointer) bool {
	return *(*int16)(p) == *(*int16)(q)
}
func memequal32(p, q unsafe.Pointer) bool {
	return *(*int32)(p) == *(*int32)(q)
}
func memequal64(p, q unsafe.Pointer) bool {
	return *(*int64)(p) == *(*int64)(q)
}
func memequal128(p, q unsafe.Pointer) bool {
	return *(*[2]int64)(p) == *(*[2]int64)(q)
}
func f32equal(p, q unsafe.Pointer) bool {
	return *(*float32)(p) == *(*float32)(q)
}
func f64equal(p, q unsafe.Pointer) bool {
	return *(*float64)(p) == *(*float64)(q)
}
func c64equal(p, q unsafe.Pointer) bool {
	return *(*complex64)(p) == *(*complex64)(q)
}
func c128equal(p, q unsafe.Pointer) bool {
	return *(*complex128)(p) == *(*complex128)(q)
}
func strequal(p, q unsafe.Pointer) bool {
	return *(*string)(p) == *(*string)(q)
}
func interequal(p, q unsafe.Pointer) bool {
	return ifaceeq(*(*interface {
		f()
	})(p), *(*interface {
		f()
	})(q))
}
func nilinterequal(p, q unsafe.Pointer) bool {
	return efaceeq(*(*interface{})(p), *(*interface{})(q))
}
func efaceeq(p, q interface{}) bool {
	x := (*_iface.Eface)(unsafe.Pointer(&p))
	y := (*_iface.Eface)(unsafe.Pointer(&q))
	t := x.Type
	if t != y.Type {
		return false
	}
	if t == nil {
		return true
	}
	eq := t.Alg.Equal
	if eq == nil {
		panic(_base.ErrorString("comparing uncomparable type " + *t.String))
	}
	if _iface.IsDirectIface(t) {
		return eq(_base.Noescape(unsafe.Pointer(&x.Data)), _base.Noescape(unsafe.Pointer(&y.Data)))
	}
	return eq(x.Data, y.Data)
}
func ifaceeq(p, q interface {
	f()
}) bool {
	x := (*_iface.Iface)(unsafe.Pointer(&p))
	y := (*_iface.Iface)(unsafe.Pointer(&q))
	xtab := x.Tab
	if xtab != y.Tab {
		return false
	}
	if xtab == nil {
		return true
	}
	t := xtab.Type
	eq := t.Alg.Equal
	if eq == nil {
		panic(_base.ErrorString("comparing uncomparable type " + *t.String))
	}
	if _iface.IsDirectIface(t) {
		return eq(_base.Noescape(unsafe.Pointer(&x.Data)), _base.Noescape(unsafe.Pointer(&y.Data)))
	}
	return eq(x.Data, y.Data)
}

// Testing adapters for hash quality tests (see hash_test.go)
func stringHash(s string, seed uintptr) uintptr {
	return algarray[alg_STRING].Hash(_base.Noescape(unsafe.Pointer(&s)), seed)
}

func bytesHash(b []byte, seed uintptr) uintptr {
	s := (*_base.Slice)(unsafe.Pointer(&b))
	return _base.Memhash(s.Array, seed, uintptr(s.Len))
}

func int32Hash(i uint32, seed uintptr) uintptr {
	return algarray[alg_MEM32].Hash(_base.Noescape(unsafe.Pointer(&i)), seed)
}

func int64Hash(i uint64, seed uintptr) uintptr {
	return algarray[alg_MEM64].Hash(_base.Noescape(unsafe.Pointer(&i)), seed)
}

func efaceHash(i interface{}, seed uintptr) uintptr {
	return algarray[alg_NILINTER].Hash(_base.Noescape(unsafe.Pointer(&i)), seed)
}

func ifaceHash(i interface {
	F()
}, seed uintptr) uintptr {
	return algarray[alg_INTER].Hash(_base.Noescape(unsafe.Pointer(&i)), seed)
}

// Testing adapter for memclr
func memclrBytes(b []byte) {
	s := (*_base.Slice)(unsafe.Pointer(&b))
	_base.Memclr(s.Array, uintptr(s.Len))
}

const hashRandomBytes = _base.PtrSize / 4 * 64

// used in asm_{386,amd64}.s to seed the hash function
var aeskeysched [hashRandomBytes]byte

func init() {
	// Install aes hash algorithm if we have the instructions we need
	if (_base.GOARCH == "386" || _base.GOARCH == "amd64") &&
		_base.GOOS != "nacl" &&
		cpuid_ecx&(1<<25) != 0 && // aes (aesenc)
		cpuid_ecx&(1<<9) != 0 && // sse3 (pshufb)
		cpuid_ecx&(1<<19) != 0 { // sse4.1 (pinsr{d,q})
		_base.UseAeshash = true
		algarray[alg_MEM32].Hash = aeshash32
		algarray[alg_MEM64].Hash = aeshash64
		algarray[alg_STRING].Hash = aeshashstr
		// Initialize with random data so hash collisions will be hard to engineer.
		getRandomData(aeskeysched[:])
		return
	}
	getRandomData((*[len(_base.Hashkey) * _base.PtrSize]byte)(unsafe.Pointer(&_base.Hashkey))[:])
	_base.Hashkey[0] |= 1 // make sure this number is odd
}
