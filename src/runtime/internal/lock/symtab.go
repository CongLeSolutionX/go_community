// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lock

import (
	_core "runtime/internal/core"
	"unsafe"
)

// funcdata.h
const (
	PCDATA_StackMapIndex       = 0
	FUNCDATA_ArgsPointerMaps   = 0
	FUNCDATA_LocalsPointerMaps = 1
	FUNCDATA_DeadValueMaps     = 2
	ArgsSizeUnknown            = -0x80000000
)

var (
	Pclntable []byte
	Ftab      []Functab
	Filetab   []uint32

	Pclntab, Epclntab, Findfunctab struct{} // linker symbols

	Minpc, Maxpc uintptr
)

type Functab struct {
	Entry   uintptr
	Funcoff uintptr
}

const minfunc = 16                 // minimum function size
const pcbucketsize = 256 * minfunc // size of bucket in the pc->func lookup table

// findfunctab is an array of these structures.
// Each bucket represents 4096 bytes of the text segment.
// Each subbucket represents 256 bytes of the text segment.
// To find a function given a pc, locate the bucket and subbucket for
// that pc.  Add together the idx and subbucket value to obtain a
// function index.  Then scan the functab array starting at that
// index to find the target function.
// This table uses 20 bytes for every 4096 bytes of code, or ~0.5% overhead.
type findfuncbucket struct {
	idx        uint32
	subbuckets [16]byte
}

func Findfunc(pc uintptr) *Func {
	if pc < Minpc || pc >= Maxpc {
		return nil
	}
	const nsub = uintptr(len(findfuncbucket{}.subbuckets))

	x := pc - Minpc
	b := x / pcbucketsize
	i := x % pcbucketsize / (pcbucketsize / nsub)

	ffb := (*findfuncbucket)(_core.Add(unsafe.Pointer(&Findfunctab), b*unsafe.Sizeof(findfuncbucket{})))
	idx := ffb.idx + uint32(ffb.subbuckets[i])
	if pc < Ftab[idx].Entry {
		Throw("findfunc: bad findfunctab entry")
	}

	// linear search to find func with pc >= entry.
	for Ftab[idx+1].Entry <= pc {
		idx++
	}
	return (*Func)(unsafe.Pointer(&Pclntable[Ftab[idx].Funcoff]))
}

func Pcvalue(f *Func, off int32, targetpc uintptr, strict bool) int32 {
	if off == 0 {
		return -1
	}
	p := Pclntable[off:]
	pc := f.Entry
	val := int32(-1)
	for {
		var ok bool
		p, ok = step(p, &pc, &val, pc == f.Entry)
		if !ok {
			break
		}
		if targetpc < pc {
			return val
		}
	}

	// If there was a table, it should have covered all program counters.
	// If not, something is wrong.
	if Panicking != 0 || !strict {
		return -1
	}

	print("runtime: invalid pc-encoded table f=", Funcname(f), " pc=", _core.Hex(pc), " targetpc=", _core.Hex(targetpc), " tab=", p, "\n")

	p = Pclntable[off:]
	pc = f.Entry
	val = -1
	for {
		var ok bool
		p, ok = step(p, &pc, &val, pc == f.Entry)
		if !ok {
			break
		}
		print("\tvalue=", val, " until pc=", _core.Hex(pc), "\n")
	}

	Throw("invalid runtime symbol table")
	return -1
}

func cfuncname(f *Func) *byte {
	if f == nil || f.nameoff == 0 {
		return nil
	}
	return (*byte)(unsafe.Pointer(&Pclntable[f.nameoff]))
}

func Funcname(f *Func) string {
	return Gostringnocopy(cfuncname(f))
}

func Funcline1(f *Func, targetpc uintptr, strict bool) (file string, line int32) {
	fileno := int(Pcvalue(f, f.pcfile, targetpc, strict))
	line = Pcvalue(f, f.pcln, targetpc, strict)
	if fileno == -1 || line == -1 || fileno >= len(Filetab) {
		// print("looking for ", hex(targetpc), " in ", funcname(f), " got file=", fileno, " line=", lineno, "\n")
		return "?", 0
	}
	file = Gostringnocopy(&Pclntable[Filetab[fileno]])
	return
}

func Funcline(f *Func, targetpc uintptr) (file string, line int32) {
	return Funcline1(f, targetpc, true)
}

func funcspdelta(f *Func, targetpc uintptr) int32 {
	x := Pcvalue(f, f.pcsp, targetpc, true)
	if x&(_core.PtrSize-1) != 0 {
		print("invalid spdelta ", _core.Hex(f.Entry), " ", _core.Hex(targetpc), " ", _core.Hex(f.pcsp), " ", x, "\n")
	}
	return x
}

// step advances to the next pc, value pair in the encoded table.
func step(p []byte, pc *uintptr, val *int32, first bool) (newp []byte, ok bool) {
	p, uvdelta := readvarint(p)
	if uvdelta == 0 && !first {
		return nil, false
	}
	if uvdelta&1 != 0 {
		uvdelta = ^(uvdelta >> 1)
	} else {
		uvdelta >>= 1
	}
	vdelta := int32(uvdelta)
	p, pcdelta := readvarint(p)
	*pc += uintptr(pcdelta * PCQuantum)
	*val += vdelta
	return p, true
}

// readvarint reads a varint from p.
func readvarint(p []byte) (newp []byte, val uint32) {
	var v, shift uint32
	for {
		b := p[0]
		p = p[1:]
		v |= (uint32(b) & 0x7F) << shift
		if b&0x80 == 0 {
			break
		}
		shift += 7
	}
	return p, v
}
