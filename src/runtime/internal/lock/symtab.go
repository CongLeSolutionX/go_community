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
)

type Functab struct {
	Entry   uintptr
	Funcoff uintptr
}

func Findfunc(pc uintptr) *Func {
	if len(Ftab) == 0 {
		return nil
	}

	if pc < Ftab[0].Entry || pc >= Ftab[len(Ftab)-1].Entry {
		return nil
	}

	// binary search to find func with entry <= pc.
	lo := 0
	nf := len(Ftab) - 1 // last entry is sentinel
	for nf > 0 {
		n := nf / 2
		f := &Ftab[lo+n]
		if f.Entry <= pc && pc < Ftab[lo+n+1].Entry {
			return (*Func)(unsafe.Pointer(&Pclntable[f.Funcoff]))
		} else if pc < f.Entry {
			nf = n
		} else {
			lo += n + 1
			nf -= n + 1
		}
	}

	Throw("findfunc: binary search failed")
	return nil
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

	print("runtime: invalid pc-encoded table f=", Gofuncname(f), " pc=", _core.Hex(pc), " targetpc=", _core.Hex(targetpc), " tab=", p, "\n")

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

func Funcname(f *Func) *byte {
	if f == nil || f.nameoff == 0 {
		return nil
	}
	return (*byte)(unsafe.Pointer(&Pclntable[f.nameoff]))
}

func Gofuncname(f *Func) string {
	return Gostringnocopy(Funcname(f))
}

func Funcline1(f *Func, targetpc uintptr, strict bool) (file string, line int32) {
	fileno := int(Pcvalue(f, f.pcfile, targetpc, strict))
	line = Pcvalue(f, f.pcln, targetpc, strict)
	if fileno == -1 || line == -1 || fileno >= len(Filetab) {
		// print("looking for ", hex(targetpc), " in ", gofuncname(f), " got file=", fileno, " line=", lineno, "\n")
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
