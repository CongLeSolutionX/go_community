// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wasm

import (
	"bytes"
	"cmd/internal/objabi"
	"cmd/link/internal/ld"
	"cmd/link/internal/sym"
	"io"
	"regexp"
)

const (
	I32 = 0x7F
	I64 = 0x7E
	F32 = 0x7D
	F64 = 0x7C
)

const (
	sectionCustom   = 0
	sectionType     = 1
	sectionImport   = 2
	sectionFunction = 3
	sectionTable    = 4
	sectionMemory   = 5
	sectionGlobal   = 6
	sectionExport   = 7
	sectionStart    = 8
	sectionElement  = 9
	sectionCode     = 10
	sectionData     = 11
)

func gentext(ctxt *ld.Link) {
}

type wasmFunc struct {
	Name string
	Type uint32
	Code []byte
}

type wasmFuncType struct {
	Params  []byte
	Results []byte
}

// Functions that get imported from the WebAssembly host (usually JavaScript).
var hostImports = []string{
	"runtime.wasmexit",
	"runtime.wasmwrite",
	"runtime.nanotime",
	"runtime.walltime",
	"runtime/js.boolVal",
	"runtime/js.intVal",
	"runtime/js.floatVal",
	"runtime/js.stringVal",
	"runtime/js.Value.Get",
	"runtime/js.Value.set",
	"runtime/js.Value.Index",
	"runtime/js.Value.setIndex",
	"runtime/js.Value.call",
	"runtime/js.Value.invoke",
	"runtime/js.Value.wasmnew",
	"runtime/js.Value.Float",
	"runtime/js.Value.Int",
	"runtime/js.Value.Bool",
	"runtime/js.Value.Length",
	"runtime/js.Value.prepareString",
	"runtime/js.Value.loadString",
}
var hostImportMap = makeHostImportMap()

func makeHostImportMap() map[string]int32 {
	m := make(map[string]int32)
	for i, name := range hostImports {
		m[name] = int32(1 + i)
	}
	return m
}

var wasmFuncTypes = map[string]*wasmFuncType{
	"_rt0_wasm_js":           &wasmFuncType{Params: []byte{I32, I32}},
	"runtime.wasmmove":       &wasmFuncType{Params: []byte{I32, I32, I32}},
	"runtime.wasmzero":       &wasmFuncType{Params: []byte{I32, I32}},
	"runtime.wasmdiv":        &wasmFuncType{Params: []byte{I64, I64}, Results: []byte{I64}},
	"runtime.wasmtrunc":      &wasmFuncType{Params: []byte{F64}, Results: []byte{I64}},
	"runtime.gcWriteBarrier": &wasmFuncType{Params: []byte{I32, I64}, Results: []byte{I32}},
	"cmpbody":                &wasmFuncType{Params: []byte{I64, I64, I64, I64}, Results: []byte{I64}},
	"memeqbody":              &wasmFuncType{Params: []byte{I64, I64, I64}, Results: []byte{I64}},
	"memcmp":                 &wasmFuncType{Params: []byte{I32, I32, I32}, Results: []byte{I32}},
	"memchr":                 &wasmFuncType{Params: []byte{I32, I32, I32}, Results: []byte{I32}},
}

func assignAddress(ctxt *ld.Link, sect *sym.Section, n int, s *sym.Symbol, va uint64, isTramp bool) (*sym.Section, int, uint64) {
	// WebAssembly functions do not live in the same address space as the linear memory.
	// Instead, they have a 0-based index which get encoded in bits 16-31 of the PC (s.Value).
	// Bits 0-15 are the index of the resume point inside of the function, which is always zero
	// for the entry of the function. Bits 32-63 are unused.
	s.Sect = sect
	s.Value = int64(len(hostImports)+n) << 16
	n++
	va += uint64(ld.MINFUNC)
	return sect, n, va
}

// asmb writes the final WebAssembly module binary.
// Spec: http://webassembly.github.io/spec/core/binary/modules.html
func asmb(ctxt *ld.Link) {
	if ctxt.Debugvlog != 0 {
		ctxt.Logf("%5.2f asmb\n", ld.Cputime())
	}

	types := []*wasmFuncType{
		&wasmFuncType{Results: []byte{I32}},
	}

	var importedFns = []*wasmFunc{
		&wasmFunc{Name: "debug", Type: lookupType(&wasmFuncType{
			Params: []byte{I32},
		}, &types)},
	}
	for _, name := range hostImports {
		importedFns = append(importedFns, &wasmFunc{
			Name: name,
			Type: lookupType(&wasmFuncType{Params: []byte{I32}}, &types),
		})
	}

	fns := make([]*wasmFunc, len(ctxt.Textp))
	for i, fn := range ctxt.Textp {
		wfn := new(bytes.Buffer)
		if fn.Name == "go.buildid" {
			writeUleb128(wfn, 0) // number of sets of locals
			wfn.WriteByte(0x41)  // i32.const
			writeSleb128(wfn, 0) // offset
			wfn.WriteByte(0x0b)  // end

		} else {
			// Relocations have variable length, handle them here.
			off := int32(0)
			for _, r := range fn.R {
				wfn.Write(fn.P[off:r.Off])
				off = r.Off
				var idx int32
				switch r.Type {
				case objabi.R_ADDR:
					idx = int32(r.Sym.Value)
				case objabi.R_CALL:
					idx = int32(r.Sym.Value >> 16)
				case objabi.R_WASMIMPORT:
					idx = hostImportMap[r.Sym.Name]
				default:
					ld.Errorf(fn, "bad reloc type %d (%s)", r.Type, sym.RelocName(ctxt.Arch, r.Type))
					continue
				}
				if idx == 0 {
					ld.Errorf(fn, "bad reloc")
					continue
				}
				writeSleb128(wfn, idx)
			}
			wfn.Write(fn.P[off:])
		}

		typ := uint32(0)
		if sig, ok := wasmFuncTypes[fn.Name]; ok {
			typ = lookupType(sig, &types)
		}

		name := nameRegexp.ReplaceAllString(fn.Name, "_")
		fns[i] = &wasmFunc{Name: name, Type: typ, Code: wfn.Bytes()}
	}

	fns = append(fns, &wasmFunc{Name: "unreachable", Type: 0, Code: []byte{0, 0x00, 0x0b}})

	allFns := append(importedFns, fns...)

	ctxt.Out.Write([]byte{0x00, 0x61, 0x73, 0x6d}) // magic
	ctxt.Out.Write([]byte{0x01, 0x00, 0x00, 0x00}) // version

	writeSec(ctxt, sectionType, func() { writeTypeSec(ctxt, types) })
	writeSec(ctxt, sectionImport, func() { writeImportSec(ctxt, importedFns) })
	writeSec(ctxt, sectionFunction, func() { writeFuncSec(ctxt, fns) })
	writeSec(ctxt, sectionTable, func() { writeTableSec(ctxt, allFns) })
	writeSec(ctxt, sectionMemory, func() { writeMemSec(ctxt) })
	writeSec(ctxt, sectionGlobal, func() { writeGlobalSec(ctxt) })
	writeSec(ctxt, sectionExport, func() { writeExportSec(ctxt) })
	writeSec(ctxt, sectionElement, func() { writeElementSec(ctxt, allFns) })
	writeSec(ctxt, sectionCode, func() { writeCodeSec(ctxt, fns) })
	writeSec(ctxt, sectionData, func() { writeDataSec(ctxt) })
	writeSec(ctxt, sectionCustom, func() { writeNameSec(ctxt, allFns) })

	ctxt.Out.Flush()
}

func lookupType(sig *wasmFuncType, types *[]*wasmFuncType) uint32 {
	for i, t := range *types {
		if bytes.Equal(sig.Params, t.Params) && bytes.Equal(sig.Results, t.Results) {
			return uint32(i)
		}
	}
	*types = append(*types, sig)
	return uint32(len(*types) - 1)
}

func writeSec(ctxt *ld.Link, id uint8, writeSecFn func()) {
	ctxt.Out.WriteByte(id)
	offsetLength := ctxt.Out.Offset()
	ctxt.Out.Write(make([]byte, 5)) // placeholder for length
	writeSecFn()
	offsetEnd := ctxt.Out.Offset()

	ctxt.Out.SeekSet(offsetLength)
	writeUleb128FixedLength(ctxt.Out, uint32(offsetEnd-offsetLength-5), 5)
	ctxt.Out.SeekSet(offsetEnd)
}

// writeTypeSec writes the section that declares all function types
// so they can be referenced by index.
func writeTypeSec(ctxt *ld.Link, types []*wasmFuncType) {
	writeUleb128(ctxt.Out, uint32(len(types)))

	for _, t := range types {
		ctxt.Out.WriteByte(0x60) // functype
		writeUleb128(ctxt.Out, uint32(len(t.Params)))
		for _, v := range t.Params {
			ctxt.Out.WriteByte(byte(v))
		}
		writeUleb128(ctxt.Out, uint32(len(t.Results)))
		for _, v := range t.Results {
			ctxt.Out.WriteByte(byte(v))
		}
	}
}

// writeImportSec writes the section that lists the functions that get
// imported from the WebAssembly host, usually JavaScript.
func writeImportSec(ctxt *ld.Link, importedFns []*wasmFunc) {
	writeUleb128(ctxt.Out, uint32(len(importedFns))) // number of imports
	for _, fn := range importedFns {
		writeName(ctxt.Out, "js") // provided by the import object in wasm_exec.js
		writeName(ctxt.Out, fn.Name)
		ctxt.Out.WriteByte(0x00) // func import
		writeUleb128(ctxt.Out, fn.Type)
	}
}

// writeFuncSec writes the section that declares the types of functions later provided in the "code" section.
func writeFuncSec(ctxt *ld.Link, fns []*wasmFunc) {
	writeUleb128(ctxt.Out, uint32(len(fns)))
	for _, fn := range fns {
		writeUleb128(ctxt.Out, fn.Type)
	}
}

// writeTableSec writes the section that declares tables. Currently there is only a single table
// that is used by the CallIndirect operation to dynamically call any function.
// The contents of the table get initialized by the "element" section.
func writeTableSec(ctxt *ld.Link, fns []*wasmFunc) {
	writeUleb128(ctxt.Out, 1)                // number of tables
	ctxt.Out.WriteByte(0x70)                 // type: anyfunc
	ctxt.Out.WriteByte(0x00)                 // no max
	writeUleb128(ctxt.Out, uint32(len(fns))) // min
}

// writeMemSec writes the section that declares linear memories. Currently one linear memory is being used.
func writeMemSec(ctxt *ld.Link) {
	writeUleb128(ctxt.Out, 1)       // number of memories
	ctxt.Out.WriteByte(0x00)        // no maximum memory size
	writeUleb128(ctxt.Out, 1024*16) // minimum (initial) memory size, linear memory always starts at address zero
}

// writeGlobalSec writes the section that declares global variables.
func writeGlobalSec(ctxt *ld.Link) {
	globalRegs := []byte{
		I32, // 0: PC_F
		I32, // 1: PC_B
		I32, // 2: SP
		I64, // 3: CTXT
		I64, // 4: g
		I64, // 5: RET0
		I64, // 6: RET1
		I64, // 7: RET2
		I64, // 8: RET3
	}

	writeUleb128(ctxt.Out, uint32(len(globalRegs))) // number of globals

	for _, typ := range globalRegs {
		ctxt.Out.WriteByte(typ)
		ctxt.Out.WriteByte(0x01) // var
		switch typ {
		case I32:
			ctxt.Out.WriteByte(0x41) // i32.const
		case I64:
			ctxt.Out.WriteByte(0x42) // i64.const
		}
		writeSleb128(ctxt.Out, 0)
		ctxt.Out.WriteByte(0x0b) // end
	}
}

// writeExportSec writes the section that declares exports.
// Exports can be accessed by the WebAssembly host, usually JavaScript.
// Currently _rt0_wasm_js (program entry point) and the linear memory get exported.
func writeExportSec(ctxt *ld.Link) {
	writeUleb128(ctxt.Out, 2) // number of exports

	rt0 := uint32(ctxt.Syms.ROLookup("_rt0_wasm_js", 0).Value >> 16)
	writeName(ctxt.Out, "run")  // inst.exports.run in wasm_exec.js
	ctxt.Out.WriteByte(0x00)    // func export
	writeUleb128(ctxt.Out, rt0) // funcidx

	writeName(ctxt.Out, "mem") // inst.exports.mem in wasm_exec.js
	ctxt.Out.WriteByte(0x02)   // mem export
	writeUleb128(ctxt.Out, 0)  // memidx
}

// writeElementSec writes the section that initializes the tables declared by the "table" section.
// The table for CallIndirect gets initialized in a very simple way so that each table index
// is equal to the function index.
func writeElementSec(ctxt *ld.Link, fns []*wasmFunc) {
	writeUleb128(ctxt.Out, 1) // number of element segments

	writeUleb128(ctxt.Out, 0) // tableidx
	ctxt.Out.WriteByte(0x41)  // i32.const
	writeSleb128(ctxt.Out, 0) // offset
	ctxt.Out.WriteByte(0x0b)  // end

	writeUleb128(ctxt.Out, uint32(len(fns))) // number of entries
	for i, fn := range fns {
		if fn.Type != 0 {
			// This function is not a normal Go function can thus will not be called via CallIndirect.
			// Set the function index to the special "unreachable" function just as a favor to the
			// WebAssembly compiler so it does not need to consider this function type as a potential
			// target for CallIndirect.
			writeUleb128(ctxt.Out, uint32(len(fns)-1)) // "unreachable"
			continue
		}
		writeUleb128(ctxt.Out, uint32(i))
	}
}

// writeElementSec writes the section that provides the function bodies for the functions
// declared by the "func" section.
func writeCodeSec(ctxt *ld.Link, fns []*wasmFunc) {
	writeUleb128(ctxt.Out, uint32(len(fns))) // number of code entries
	for _, fn := range fns {
		writeUleb128(ctxt.Out, uint32(len(fn.Code)))
		ctxt.Out.Write(fn.Code)
	}
}

// writeDataSec writes the section that provides data that will be used to initialize the linear memory.
func writeDataSec(ctxt *ld.Link) {
	sections := []*sym.Section{
		ctxt.Syms.Lookup("runtime.rodata", 0).Sect,
		ctxt.Syms.Lookup("runtime.typelink", 0).Sect,
		ctxt.Syms.Lookup("runtime.itablink", 0).Sect,
		ctxt.Syms.Lookup("runtime.symtab", 0).Sect,
		ctxt.Syms.Lookup("runtime.pclntab", 0).Sect,
		ctxt.Syms.Lookup("runtime.noptrdata", 0).Sect,
		ctxt.Syms.Lookup("runtime.data", 0).Sect,
	}

	writeUleb128(ctxt.Out, uint32(len(sections))) // number of data entries

	for _, sec := range sections {
		writeUleb128(ctxt.Out, 0) // memidx
		ctxt.Out.WriteByte(0x41)  // i32.const
		writeSleb128(ctxt.Out, int32(sec.Vaddr))
		ctxt.Out.WriteByte(0x0b) // end
		writeUleb128(ctxt.Out, uint32(sec.Length))
		ld.Datblk(ctxt, int64(sec.Vaddr), int64(sec.Length))
	}
}

var nameRegexp = regexp.MustCompile(`[^\w\.]`)

// writeNameSec writes an optional section that assigns names to the functions declared by the "func" section.
// The names are only used by WebAssembly stack traces, debuggers and decompilers.
func writeNameSec(ctxt *ld.Link, fns []*wasmFunc) {
	writeName(ctxt.Out, "name")

	ctxt.Out.WriteByte(0x01) // function names

	offsetLength := ctxt.Out.Offset()
	ctxt.Out.Write(make([]byte, 5)) // placeholder for length
	writeUleb128(ctxt.Out, uint32(len(fns)))
	for i, fn := range fns {
		writeUleb128(ctxt.Out, uint32(i))
		writeName(ctxt.Out, fn.Name)
	}
	offsetEnd := ctxt.Out.Offset()

	ctxt.Out.SeekSet(offsetLength)
	writeUleb128FixedLength(ctxt.Out, uint32(offsetEnd-offsetLength-5), 5)
	ctxt.Out.SeekSet(offsetEnd)
}

type nameWriter interface {
	io.ByteWriter
	io.Writer
}

func writeName(w nameWriter, name string) {
	writeUleb128(w, uint32(len(name)))
	w.Write([]byte(name))
}

func writeUleb128(w io.ByteWriter, v uint32) {
	more := true
	for more {
		c := uint8(v & 0x7f)
		v >>= 7
		more = v != 0
		if more {
			c |= 0x80
		}
		w.WriteByte(c)
	}
}

func writeUleb128FixedLength(w *ld.OutBuf, v uint32, length int) {
	for i := 0; i < length; i++ {
		c := uint8(v & 0x7f)
		v >>= 7
		if i < length-1 {
			c |= 0x80
		}
		w.WriteByte(c)
	}
	if v != 0 {
		panic("writeUleb128FixedLength: length too small")
	}
}

func writeSleb128(w io.ByteWriter, v int32) {
	more := true
	for more {
		c := uint8(v & 0x7f)
		s := uint8(v & 0x40)
		v >>= 7
		more = !((v == 0 && s == 0) || (v == -1 && s != 0))
		if more {
			c |= 0x80
		}
		w.WriteByte(c)
	}
}
