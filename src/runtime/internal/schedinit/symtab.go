// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schedinit

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
	"unsafe"
)

func symtabinit() {
	// See golang.org/s/go12symtab for header: 0xfffffffb,
	// two zero bytes, a byte giving the PC quantum,
	// and a byte giving the pointer width in bytes.
	pcln := (*[8]byte)(unsafe.Pointer(&_lock.Pclntab))
	pcln32 := (*[2]uint32)(unsafe.Pointer(&_lock.Pclntab))
	if pcln32[0] != 0xfffffffb || pcln[4] != 0 || pcln[5] != 0 || pcln[6] != _lock.PCQuantum || pcln[7] != _core.PtrSize {
		println("runtime: function symbol table header:", _core.Hex(pcln32[0]), _core.Hex(pcln[4]), _core.Hex(pcln[5]), _core.Hex(pcln[6]), _core.Hex(pcln[7]))
		_lock.Throw("invalid function symbol table\n")
	}

	// pclntable is all bytes of pclntab symbol.
	sp := (*_sched.SliceStruct)(unsafe.Pointer(&_lock.Pclntable))
	sp.Array = unsafe.Pointer(&_lock.Pclntab)
	sp.Len = int(uintptr(unsafe.Pointer(&_lock.Epclntab)) - uintptr(unsafe.Pointer(&_lock.Pclntab)))
	sp.Cap = sp.Len

	// ftab is lookup table for function by program counter.
	nftab := int(*(*uintptr)(_core.Add(unsafe.Pointer(pcln), 8)))
	p := _core.Add(unsafe.Pointer(pcln), 8+_core.PtrSize)
	sp = (*_sched.SliceStruct)(unsafe.Pointer(&_lock.Ftab))
	sp.Array = p
	sp.Len = nftab + 1
	sp.Cap = sp.Len
	for i := 0; i < nftab; i++ {
		// NOTE: ftab[nftab].entry is legal; it is the address beyond the final function.
		if _lock.Ftab[i].Entry > _lock.Ftab[i+1].Entry {
			f1 := (*_lock.Func)(unsafe.Pointer(&_lock.Pclntable[_lock.Ftab[i].Funcoff]))
			f2 := (*_lock.Func)(unsafe.Pointer(&_lock.Pclntable[_lock.Ftab[i+1].Funcoff]))
			f2name := "end"
			if i+1 < nftab {
				f2name = _lock.Funcname(f2)
			}
			println("function symbol table not sorted by program counter:", _core.Hex(_lock.Ftab[i].Entry), _lock.Funcname(f1), ">", _core.Hex(_lock.Ftab[i+1].Entry), f2name)
			for j := 0; j <= i; j++ {
				print("\t", _core.Hex(_lock.Ftab[j].Entry), " ", _lock.Funcname((*_lock.Func)(unsafe.Pointer(&_lock.Pclntable[_lock.Ftab[j].Funcoff]))), "\n")
			}
			_lock.Throw("invalid runtime symbol table")
		}
	}

	// The ftab ends with a half functab consisting only of
	// 'entry', followed by a uint32 giving the pcln-relative
	// offset of the file table.
	sp = (*_sched.SliceStruct)(unsafe.Pointer(&_lock.Filetab))
	end := unsafe.Pointer(&_lock.Ftab[nftab].Funcoff) // just beyond ftab
	fileoffset := *(*uint32)(end)
	sp.Array = unsafe.Pointer(&_lock.Pclntable[fileoffset])
	// length is in first element of array.
	// set len to 1 so we can get first element.
	sp.Len = 1
	sp.Cap = 1
	sp.Len = int(_lock.Filetab[0])
	sp.Cap = sp.Len

	_lock.Minpc = _lock.Ftab[0].Entry
	_lock.Maxpc = _lock.Ftab[nftab].Entry
}
