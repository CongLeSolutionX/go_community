// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package base

import (
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

// moduledata records information about the layout of the executable
// image. It is written by the linker. Any changes here must be
// matched changes to the code in cmd/internal/ld/symtab.go:symtab.
// moduledata is stored in read-only memory; none of the pointers here
// are visible to the garbage collector.
type Moduledata struct {
	Pclntable    []byte
	Ftab         []Functab
	filetab      []uint32
	findfunctab  uintptr
	Minpc, Maxpc uintptr

	text, etext           uintptr
	Noptrdata, Enoptrdata uintptr
	Data, Edata           uintptr
	Bss, Ebss             uintptr
	Noptrbss, Enoptrbss   uintptr
	End, Gcdata, Gcbss    uintptr

	Typelinks []*Type

	Modulename   string
	Modulehashes []Modulehash

	Gcdatamask, Gcbssmask Bitvector

	Next *Moduledata
}

// For each shared library a module links against, the linker creates an entry in the
// moduledata.modulehashes slice containing the name of the module, the abi hash seen
// at link time and a pointer to the runtime abi hash. These are checked in
// moduledataverify1 below.
type Modulehash struct {
	Modulename   string
	Linktimehash string
	Runtimehash  *string
}

var Firstmoduledata Moduledata // linker symbol

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

func findmoduledatap(pc uintptr) *Moduledata {
	for datap := &Firstmoduledata; datap != nil; datap = datap.Next {
		if datap.Minpc <= pc && pc <= datap.Maxpc {
			return datap
		}
	}
	return nil
}

func Findfunc(pc uintptr) *Func {
	datap := findmoduledatap(pc)
	if datap == nil {
		return nil
	}
	const nsub = uintptr(len(findfuncbucket{}.subbuckets))

	x := pc - datap.Minpc
	b := x / pcbucketsize
	i := x % pcbucketsize / (pcbucketsize / nsub)

	ffb := (*findfuncbucket)(Add(unsafe.Pointer(datap.findfunctab), b*unsafe.Sizeof(findfuncbucket{})))
	idx := ffb.idx + uint32(ffb.subbuckets[i])
	if pc < datap.Ftab[idx].Entry {
		Throw("findfunc: bad findfunctab entry")
	}

	// linear search to find func with pc >= entry.
	for datap.Ftab[idx+1].Entry <= pc {
		idx++
	}
	return (*Func)(unsafe.Pointer(&datap.Pclntable[datap.Ftab[idx].Funcoff]))
}

func Pcvalue(f *Func, off int32, targetpc uintptr, strict bool) int32 {
	if off == 0 {
		return -1
	}
	datap := findmoduledatap(f.Entry) // inefficient
	if datap == nil {
		if strict && Panicking == 0 {
			print("runtime: no module data for ", Hex(f.Entry), "\n")
			Throw("no module data")
		}
		return -1
	}
	p := datap.Pclntable[off:]
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

	print("runtime: invalid pc-encoded table f=", Funcname(f), " pc=", Hex(pc), " targetpc=", Hex(targetpc), " tab=", p, "\n")

	p = datap.Pclntable[off:]
	pc = f.Entry
	val = -1
	for {
		var ok bool
		p, ok = step(p, &pc, &val, pc == f.Entry)
		if !ok {
			break
		}
		print("\tvalue=", val, " until pc=", Hex(pc), "\n")
	}

	Throw("invalid runtime symbol table")
	return -1
}

func cfuncname(f *Func) *byte {
	if f == nil || f.nameoff == 0 {
		return nil
	}
	datap := findmoduledatap(f.Entry) // inefficient
	if datap == nil {
		return nil
	}
	return (*byte)(unsafe.Pointer(&datap.Pclntable[f.nameoff]))
}

func Funcname(f *Func) string {
	return Gostringnocopy(cfuncname(f))
}

func Funcline1(f *Func, targetpc uintptr, strict bool) (file string, line int32) {
	datap := findmoduledatap(f.Entry) // inefficient
	if datap == nil {
		return "?", 0
	}
	fileno := int(Pcvalue(f, f.Pcfile, targetpc, strict))
	line = Pcvalue(f, f.Pcln, targetpc, strict)
	if fileno == -1 || line == -1 || fileno >= len(datap.filetab) {
		// print("looking for ", hex(targetpc), " in ", funcname(f), " got file=", fileno, " line=", lineno, "\n")
		return "?", 0
	}
	file = Gostringnocopy(&datap.Pclntable[datap.filetab[fileno]])
	return
}

func Funcline(f *Func, targetpc uintptr) (file string, line int32) {
	return Funcline1(f, targetpc, true)
}

func funcspdelta(f *Func, targetpc uintptr) int32 {
	x := Pcvalue(f, f.Pcsp, targetpc, true)
	if x&(PtrSize-1) != 0 {
		print("invalid spdelta ", Funcname(f), " ", Hex(f.Entry), " ", Hex(targetpc), " ", Hex(f.Pcsp), " ", x, "\n")
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
