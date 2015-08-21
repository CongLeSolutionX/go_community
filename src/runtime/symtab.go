// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	_base "runtime/internal/base"
	"unsafe"
)

// NOTE: Func does not expose the actual unexported fields, because we return *Func
// values to users, and we want to keep them from being able to overwrite the data
// with (say) *f = Func{}.
// All code operating on a *Func must call raw to get the *_func instead.

// A Func represents a Go function in the running binary.
type Func struct {
	opaque struct{} // unexported field to disallow conversions
}

func (f *Func) raw() *_base.Func {
	return (*_base.Func)(unsafe.Pointer(f))
}

var lastmoduledatap *_base.Moduledata // linker symbol

func moduledataverify() {
	for datap := &_base.Firstmoduledata; datap != nil; datap = datap.Next {
		moduledataverify1(datap)
	}
}

const debugPcln = false

func moduledataverify1(datap *_base.Moduledata) {
	// See golang.org/s/go12symtab for header: 0xfffffffb,
	// two zero bytes, a byte giving the PC quantum,
	// and a byte giving the pointer width in bytes.
	pcln := *(**[8]byte)(unsafe.Pointer(&datap.Pclntable))
	pcln32 := *(**[2]uint32)(unsafe.Pointer(&datap.Pclntable))
	if pcln32[0] != 0xfffffffb || pcln[4] != 0 || pcln[5] != 0 || pcln[6] != _base.PCQuantum || pcln[7] != _base.PtrSize {
		println("runtime: function symbol table header:", _base.Hex(pcln32[0]), _base.Hex(pcln[4]), _base.Hex(pcln[5]), _base.Hex(pcln[6]), _base.Hex(pcln[7]))
		_base.Throw("invalid function symbol table\n")
	}

	// ftab is lookup table for function by program counter.
	nftab := len(datap.Ftab) - 1
	for i := 0; i < nftab; i++ {
		// NOTE: ftab[nftab].entry is legal; it is the address beyond the final function.
		if datap.Ftab[i].Entry > datap.Ftab[i+1].Entry {
			f1 := (*_base.Func)(unsafe.Pointer(&datap.Pclntable[datap.Ftab[i].Funcoff]))
			f2 := (*_base.Func)(unsafe.Pointer(&datap.Pclntable[datap.Ftab[i+1].Funcoff]))
			f2name := "end"
			if i+1 < nftab {
				f2name = _base.Funcname(f2)
			}
			println("function symbol table not sorted by program counter:", _base.Hex(datap.Ftab[i].Entry), _base.Funcname(f1), ">", _base.Hex(datap.Ftab[i+1].Entry), f2name)
			for j := 0; j <= i; j++ {
				print("\t", _base.Hex(datap.Ftab[j].Entry), " ", _base.Funcname((*_base.Func)(unsafe.Pointer(&datap.Pclntable[datap.Ftab[j].Funcoff]))), "\n")
			}
			_base.Throw("invalid runtime symbol table")
		}

		if debugPcln || nftab-i < 5 {
			// Check a PC near but not at the very end.
			// The very end might be just padding that is not covered by the tables.
			// No architecture rounds function entries to more than 16 bytes,
			// but if one came along we'd need to subtract more here.
			// But don't use the next PC if it corresponds to a foreign object chunk
			// (no pcln table, f2.pcln == 0). That chunk might have an alignment
			// more than 16 bytes.
			f := (*_base.Func)(unsafe.Pointer(&datap.Pclntable[datap.Ftab[i].Funcoff]))
			end := f.Entry
			if i+1 < nftab {
				f2 := (*_base.Func)(unsafe.Pointer(&datap.Pclntable[datap.Ftab[i+1].Funcoff]))
				if f2.Pcln != 0 {
					end = f2.Entry - 16
					if end < f.Entry {
						end = f.Entry
					}
				}
			}
			_base.Pcvalue(f, f.Pcfile, end, true)
			_base.Pcvalue(f, f.Pcln, end, true)
			_base.Pcvalue(f, f.Pcsp, end, true)
		}
	}

	if datap.Minpc != datap.Ftab[0].Entry ||
		datap.Maxpc != datap.Ftab[nftab].Entry {
		_base.Throw("minpc or maxpc invalid")
	}

	for _, modulehash := range datap.Modulehashes {
		if modulehash.Linktimehash != *modulehash.Runtimehash {
			println("abi mismatch detected between", datap.Modulename, "and", modulehash.Modulename)
			_base.Throw("abi mismatch")
		}
	}
}

// FuncForPC returns a *Func describing the function that contains the
// given program counter address, or else nil.
func FuncForPC(pc uintptr) *Func {
	return (*Func)(unsafe.Pointer(_base.Findfunc(pc)))
}

// Name returns the name of the function.
func (f *Func) Name() string {
	return _base.Funcname(f.raw())
}

// Entry returns the entry address of the function.
func (f *Func) Entry() uintptr {
	return f.raw().Entry
}

// FileLine returns the file name and line number of the
// source code corresponding to the program counter pc.
// The result will not be accurate if pc is not a program
// counter within f.
func (f *Func) FileLine(pc uintptr) (file string, line int) {
	// Pass strict=false here, because anyone can call this function,
	// and they might just be wrong about targetpc belonging to f.
	file, line32 := _base.Funcline1(f.raw(), pc, false)
	return file, int(line32)
}
