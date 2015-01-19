// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hash

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
	"unsafe"
)

const (
	c0 = uintptr((8-_core.PtrSize)/4*2860486313 + (_core.PtrSize-4)/4*33054211828000289)
	c1 = uintptr((8-_core.PtrSize)/4*3267000013 + (_core.PtrSize-4)/4*23344194077549503)
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

var algarray = [alg_max]_core.TypeAlg{
	alg_MEM:      {memhash, memequal},
	alg_MEM0:     {memhash, memequal0},
	alg_MEM8:     {memhash, memequal8},
	alg_MEM16:    {memhash, memequal16},
	alg_MEM32:    {memhash, memequal32},
	alg_MEM64:    {memhash, memequal64},
	alg_MEM128:   {memhash, memequal128},
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

var useAeshash bool

// in asm_*.s
func aeshash(p unsafe.Pointer, s, h uintptr) uintptr
func aeshash32(p unsafe.Pointer, s, h uintptr) uintptr
func aeshash64(p unsafe.Pointer, s, h uintptr) uintptr
func aeshashstr(p unsafe.Pointer, s, h uintptr) uintptr

func strhash(a unsafe.Pointer, s, h uintptr) uintptr {
	x := (*_lock.StringStruct)(a)
	return memhash(x.Str, uintptr(x.Len), h)
}

// NOTE: Because NaN != NaN, a map can contain any
// number of (mostly useless) entries keyed with NaNs.
// To avoid long hash chains, we assign a random number
// as the hash value for a NaN.

func f32hash(p unsafe.Pointer, s, h uintptr) uintptr {
	f := *(*float32)(p)
	switch {
	case f == 0:
		return c1 * (c0 ^ h) // +0, -0
	case f != f:
		return c1 * (c0 ^ h ^ uintptr(_lock.Fastrand1())) // any kind of NaN
	default:
		return memhash(p, 4, h)
	}
}

func f64hash(p unsafe.Pointer, s, h uintptr) uintptr {
	f := *(*float64)(p)
	switch {
	case f == 0:
		return c1 * (c0 ^ h) // +0, -0
	case f != f:
		return c1 * (c0 ^ h ^ uintptr(_lock.Fastrand1())) // any kind of NaN
	default:
		return memhash(p, 8, h)
	}
}

func c64hash(p unsafe.Pointer, s, h uintptr) uintptr {
	x := (*[2]float32)(p)
	return f32hash(unsafe.Pointer(&x[1]), 4, f32hash(unsafe.Pointer(&x[0]), 4, h))
}

func c128hash(p unsafe.Pointer, s, h uintptr) uintptr {
	x := (*[2]float64)(p)
	return f64hash(unsafe.Pointer(&x[1]), 8, f64hash(unsafe.Pointer(&x[0]), 8, h))
}

func interhash(p unsafe.Pointer, s, h uintptr) uintptr {
	a := (*Iface)(p)
	tab := a.Tab
	if tab == nil {
		return h
	}
	t := tab.Type
	fn := Goalg(t.Alg).Hash
	if fn == nil {
		panic(_sched.ErrorString("hash of unhashable type " + *t.String))
	}
	if IsDirectIface(t) {
		return c1 * fn(unsafe.Pointer(&a.Data), uintptr(t.Size), h^c0)
	} else {
		return c1 * fn(a.Data, uintptr(t.Size), h^c0)
	}
}

func nilinterhash(p unsafe.Pointer, s, h uintptr) uintptr {
	a := (*_core.Eface)(p)
	t := a.Type
	if t == nil {
		return h
	}
	fn := Goalg(t.Alg).Hash
	if fn == nil {
		panic(_sched.ErrorString("hash of unhashable type " + *t.String))
	}
	if IsDirectIface(t) {
		return c1 * fn(unsafe.Pointer(&a.Data), uintptr(t.Size), h^c0)
	} else {
		return c1 * fn(a.Data, uintptr(t.Size), h^c0)
	}
}

func memequal(p, q unsafe.Pointer, size uintptr) bool {
	if p == q {
		return true
	}
	return Memeq(p, q, size)
}

func memequal0(p, q unsafe.Pointer, size uintptr) bool {
	return true
}
func memequal8(p, q unsafe.Pointer, size uintptr) bool {
	return *(*int8)(p) == *(*int8)(q)
}
func memequal16(p, q unsafe.Pointer, size uintptr) bool {
	return *(*int16)(p) == *(*int16)(q)
}
func memequal32(p, q unsafe.Pointer, size uintptr) bool {
	return *(*int32)(p) == *(*int32)(q)
}
func memequal64(p, q unsafe.Pointer, size uintptr) bool {
	return *(*int64)(p) == *(*int64)(q)
}
func memequal128(p, q unsafe.Pointer, size uintptr) bool {
	return *(*[2]int64)(p) == *(*[2]int64)(q)
}
func f32equal(p, q unsafe.Pointer, size uintptr) bool {
	return *(*float32)(p) == *(*float32)(q)
}
func f64equal(p, q unsafe.Pointer, size uintptr) bool {
	return *(*float64)(p) == *(*float64)(q)
}
func c64equal(p, q unsafe.Pointer, size uintptr) bool {
	return *(*complex64)(p) == *(*complex64)(q)
}
func c128equal(p, q unsafe.Pointer, size uintptr) bool {
	return *(*complex128)(p) == *(*complex128)(q)
}
func strequal(p, q unsafe.Pointer, size uintptr) bool {
	return *(*string)(p) == *(*string)(q)
}
func interequal(p, q unsafe.Pointer, size uintptr) bool {
	return ifaceeq(*(*interface {
		f()
	})(p), *(*interface {
		f()
	})(q))
}
func nilinterequal(p, q unsafe.Pointer, size uintptr) bool {
	return efaceeq(*(*interface{})(p), *(*interface{})(q))
}
func efaceeq(p, q interface{}) bool {
	x := (*_core.Eface)(unsafe.Pointer(&p))
	y := (*_core.Eface)(unsafe.Pointer(&q))
	t := x.Type
	if t != y.Type {
		return false
	}
	if t == nil {
		return true
	}
	eq := Goalg(t.Alg).Equal
	if eq == nil {
		panic(_sched.ErrorString("comparing uncomparable type " + *t.String))
	}
	if IsDirectIface(t) {
		return eq(_core.Noescape(unsafe.Pointer(&x.Data)), _core.Noescape(unsafe.Pointer(&y.Data)), uintptr(t.Size))
	}
	return eq(x.Data, y.Data, uintptr(t.Size))
}
func ifaceeq(p, q interface {
	f()
}) bool {
	x := (*Iface)(unsafe.Pointer(&p))
	y := (*Iface)(unsafe.Pointer(&q))
	xtab := x.Tab
	if xtab != y.Tab {
		return false
	}
	if xtab == nil {
		return true
	}
	t := xtab.Type
	eq := Goalg(t.Alg).Equal
	if eq == nil {
		panic(_sched.ErrorString("comparing uncomparable type " + *t.String))
	}
	if IsDirectIface(t) {
		return eq(_core.Noescape(unsafe.Pointer(&x.Data)), _core.Noescape(unsafe.Pointer(&y.Data)), uintptr(t.Size))
	}
	return eq(x.Data, y.Data, uintptr(t.Size))
}

// Testing adapters for hash quality tests (see hash_test.go)
func stringHash(s string, seed uintptr) uintptr {
	return algarray[alg_STRING].Hash(_core.Noescape(unsafe.Pointer(&s)), unsafe.Sizeof(s), seed)
}

func bytesHash(b []byte, seed uintptr) uintptr {
	s := (*_sched.SliceStruct)(unsafe.Pointer(&b))
	return algarray[alg_MEM].Hash(s.Array, uintptr(s.Len), seed)
}

func int32Hash(i uint32, seed uintptr) uintptr {
	return algarray[alg_MEM32].Hash(_core.Noescape(unsafe.Pointer(&i)), 4, seed)
}

func int64Hash(i uint64, seed uintptr) uintptr {
	return algarray[alg_MEM64].Hash(_core.Noescape(unsafe.Pointer(&i)), 8, seed)
}

func efaceHash(i interface{}, seed uintptr) uintptr {
	return algarray[alg_NILINTER].Hash(_core.Noescape(unsafe.Pointer(&i)), unsafe.Sizeof(i), seed)
}

func ifaceHash(i interface {
	F()
}, seed uintptr) uintptr {
	return algarray[alg_INTER].Hash(_core.Noescape(unsafe.Pointer(&i)), unsafe.Sizeof(i), seed)
}

// TODO(dvyukov): remove when Type is converted to Go and contains *typeAlg.
func Goalg(a unsafe.Pointer) *_core.TypeAlg {
	return (*_core.TypeAlg)(a)
}

// used in asm_{386,amd64}.s
const hashRandomBytes = _core.PtrSize / 4 * 64

var aeskeysched [hashRandomBytes]byte

func init() {
	if _lock.GOOS == "nacl" {
		return
	}

	// Install aes hash algorithm if we have the instructions we need
	if (cpuid_ecx&(1<<25)) != 0 && // aes (aesenc)
		(cpuid_ecx&(1<<9)) != 0 && // sse3 (pshufb)
		(cpuid_ecx&(1<<19)) != 0 { // sse4.1 (pinsr{d,q})
		useAeshash = true
		algarray[alg_MEM].Hash = aeshash
		algarray[alg_MEM8].Hash = aeshash
		algarray[alg_MEM16].Hash = aeshash
		algarray[alg_MEM32].Hash = aeshash32
		algarray[alg_MEM64].Hash = aeshash64
		algarray[alg_MEM128].Hash = aeshash
		algarray[alg_STRING].Hash = aeshashstr
		// Initialize with random data so hash collisions will be hard to engineer.
		getRandomData(aeskeysched[:])
	}
}
