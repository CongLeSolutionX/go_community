// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wasm

import (
	"bufio"
	"bytes"
	"cmd/internal/objabi"
	"cmd/link/internal/ld"
	"cmd/link/internal/sym"
	"regexp"
)

const (
	I32 = 0x7F
	I64 = 0x7E
	F32 = 0x7D
	F64 = 0x7C
)

func gentext(ctxt *ld.Link) {
}

type wasmFunc struct {
	Name string
	Type uint32
	Code []byte
}

type wasmSignature struct {
	Params  []byte
	Results []byte
}

var imports = []string{
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
var importMap = makeImportMap()

func makeImportMap() map[string]int32 {
	m := make(map[string]int32)
	for i, name := range imports {
		m[name] = int32(1 + i)
	}
	return m
}

var wasmSignatures = map[string]*wasmSignature{
	"_rt0_wasm_js":           &wasmSignature{Params: []byte{I32, I32}},
	"runtime.wasmmove":       &wasmSignature{Params: []byte{I32, I32, I32}},
	"runtime.wasmzero":       &wasmSignature{Params: []byte{I32, I32}},
	"runtime.wasmdiv":        &wasmSignature{Params: []byte{I64, I64}, Results: []byte{I64}},
	"runtime.wasmtrunc":      &wasmSignature{Params: []byte{F64}, Results: []byte{I64}},
	"runtime.gcWriteBarrier": &wasmSignature{Params: []byte{I32, I64}, Results: []byte{I32}},
	"cmpbody":                &wasmSignature{Params: []byte{I64, I64, I64, I64}, Results: []byte{I64}},
	"memeqbody":              &wasmSignature{Params: []byte{I64, I64, I64}, Results: []byte{I64}},
	"memcmp":                 &wasmSignature{Params: []byte{I32, I32, I32}, Results: []byte{I32}},
	"memchr":                 &wasmSignature{Params: []byte{I32, I32, I32}, Results: []byte{I32}},
}

func assignAddress(ctxt *ld.Link, sect *sym.Section, n int, s *sym.Symbol, va uint64, isTramp bool) (*sym.Section, int, uint64) {
	// WebAssembly functions do not live in the same address space as the linear memory.
	// Instead, they have a 0-based index which get encoded in bits 16-31 of s.Value.
	s.Sect = sect
	s.Value = int64(len(imports)+n) << 16
	n++
	va += uint64(ld.MINFUNC)
	return sect, n, va
}

func asmb(ctxt *ld.Link) {
	if ctxt.Debugvlog != 0 {
		ctxt.Logf("%5.2f asmb\n", ld.Cputime())
	}

	types := []*wasmSignature{
		&wasmSignature{Results: []byte{I32}},
	}
	lookupType := func(sig *wasmSignature) uint32 {
		for i, t := range types {
			if bytes.Equal(sig.Params, t.Params) && bytes.Equal(sig.Results, t.Results) {
				return uint32(i)
			}
		}
		types = append(types, sig)
		return uint32(len(types) - 1)
	}

	var importedFns = []*wasmFunc{
		&wasmFunc{Name: "debug", Type: lookupType(&wasmSignature{
			Params: []byte{I32},
		})},
	}
	for _, name := range imports {
		importedFns = append(importedFns, &wasmFunc{
			Name: name,
			Type: lookupType(&wasmSignature{Params: []byte{I32}}),
		})
	}

	fns := make([]*wasmFunc, len(ctxt.Textp))
	for i, fn := range ctxt.Textp {
		wfn := new(bytes.Buffer)
		if fn.Name == "go.buildid" || fn.Name == "runtime.skipPleaseUseCallersFrames" {
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
					idx = importMap[r.Sym.Name]
				default:
					panic("bad reloc type")
				}
				if idx == 0 {
					panic("bad reloc")
				}
				writeSleb128(wfn, idx)
			}
			wfn.Write(fn.P[off:])
		}

		typ := uint32(0)
		if sig, ok := wasmSignatures[fn.Name]; ok {
			typ = lookupType(sig)
		}

		name := nameRegexp.ReplaceAllString(fn.Name, "_")
		fns[i] = &wasmFunc{Name: name, Type: typ, Code: wfn.Bytes()}
	}

	fns = append(fns, &wasmFunc{Name: "unreachable", Type: 0, Code: []byte{0, 0x00, 0x0b}})

	allFns := append(importedFns, fns...)

	ctxt.Out.Write([]byte{0x00, 0x61, 0x73, 0x6d}) // magic
	ctxt.Out.Write([]byte{0x01, 0x00, 0x00, 0x00}) // version

	writeSec(ctxt, 1, genTypeSec(types))
	writeSec(ctxt, 2, genImportSec(importedFns))
	writeSec(ctxt, 3, genFuncSec(fns))
	writeSec(ctxt, 4, genTableSec(allFns))
	writeSec(ctxt, 5, genMemSec())
	writeSec(ctxt, 6, genGlobalSec())
	writeSec(ctxt, 7, genExportSec(ctxt))
	writeSec(ctxt, 9, genElementSec(allFns))
	writeSec(ctxt, 10, genCodeSec(fns))
	writeSec(ctxt, 11, genDataSec(ctxt))
	writeSec(ctxt, 0, genNameSec(allFns))

	ctxt.Out.Flush()
}

func writeSec(ctxt *ld.Link, id uint8, b []byte) {
	ctxt.Out.Write8(id)
	w := new(bytes.Buffer)
	writeUleb128(w, uint32(len(b)))
	w.Write(b)
	ctxt.Out.Write(w.Bytes())
}

func genTypeSec(types []*wasmSignature) []byte {
	w := new(bytes.Buffer)
	writeUleb128(w, uint32(len(types)))

	for _, t := range types {
		w.WriteByte(0x60) // functype
		writeUleb128(w, uint32(len(t.Params)))
		for _, v := range t.Params {
			w.WriteByte(byte(v))
		}
		writeUleb128(w, uint32(len(t.Results)))
		for _, v := range t.Results {
			w.WriteByte(byte(v))
		}
	}

	return w.Bytes()
}

func genImportSec(importedFns []*wasmFunc) []byte {
	w := new(bytes.Buffer)
	writeUleb128(w, uint32(len(importedFns))) // number of imports
	for _, fn := range importedFns {
		writeName(w, "js")
		writeName(w, fn.Name)
		w.WriteByte(0x00) // func import
		writeUleb128(w, fn.Type)
	}
	return w.Bytes()
}

func genFuncSec(fns []*wasmFunc) []byte {
	w := new(bytes.Buffer)
	writeUleb128(w, uint32(len(fns)))
	for _, fn := range fns {
		writeUleb128(w, fn.Type)
	}
	return w.Bytes()
}

func genTableSec(fns []*wasmFunc) []byte {
	w := new(bytes.Buffer)
	writeUleb128(w, 1)                // number of tables
	w.WriteByte(0x70)                 // type: anyfunc
	w.WriteByte(0x00)                 // no max
	writeUleb128(w, uint32(len(fns))) // min
	return w.Bytes()
}

func genMemSec() []byte {
	w := new(bytes.Buffer)
	writeUleb128(w, 1)       // number of memories
	w.WriteByte(0x00)        // no max
	writeUleb128(w, 1024*16) // min
	return w.Bytes()
}

func genGlobalSec() []byte {
	globalRegs := []byte{
		I32, // 0: PC_F
		I32, // 1: PC_B
		I32, // 2: SP
		I64, // 3: CTX
		I64, // 4: g
		I64, // 5: RET0
		I64, // 6: RET1
		I64, // 7: RET2
		I64, // 8: RET3
	}

	w := new(bytes.Buffer)
	writeUleb128(w, uint32(len(globalRegs))) // number of globals

	for _, typ := range globalRegs {
		w.WriteByte(typ)
		w.WriteByte(0x01) // var
		switch typ {
		case I32:
			w.WriteByte(0x41) // i32.const
		case I64:
			w.WriteByte(0x42) // i64.const
		}
		writeSleb128(w, 0)
		w.WriteByte(0x0b) // end
	}

	return w.Bytes()
}

func genExportSec(ctxt *ld.Link) []byte {
	w := new(bytes.Buffer)
	writeUleb128(w, 2) // number of exports

	writeName(w, "run")
	w.WriteByte(0x00)                                                        // func export
	writeUleb128(w, uint32(ctxt.Syms.ROLookup("_rt0_wasm_js", 0).Value>>16)) // funcidx

	writeName(w, "mem")
	w.WriteByte(0x02)  // mem export
	writeUleb128(w, 0) // memidx

	return w.Bytes()
}

func genElementSec(fns []*wasmFunc) []byte {
	w := new(bytes.Buffer)
	writeUleb128(w, 1) // number of element segments

	writeUleb128(w, 0) // tableidx
	w.WriteByte(0x41)  // i32.const
	writeSleb128(w, 0) // offset
	w.WriteByte(0x0b)  // end

	writeUleb128(w, uint32(len(fns))) // number of entries
	for i, fn := range fns {
		if fn.Type != 0 {
			writeUleb128(w, uint32(len(fns)-1)) // "unreachable"
			continue
		}
		writeUleb128(w, uint32(i))
	}

	return w.Bytes()
}

func genCodeSec(fns []*wasmFunc) []byte {
	w := new(bytes.Buffer)
	writeUleb128(w, uint32(len(fns))) // number of code entries
	for _, fn := range fns {
		writeUleb128(w, uint32(len(fn.Code)))
		w.Write(fn.Code)
	}
	return w.Bytes()
}

func genDataSec(ctxt *ld.Link) []byte {
	sections := []*sym.Section{
		findSymSection(ld.Segtext, ".rodata"),
		findSymSection(ld.Segtext, ".typelink"),
		findSymSection(ld.Segtext, ".itablink"),
		findSymSection(ld.Segtext, ".gosymtab"),
		findSymSection(ld.Segtext, ".gopclntab"),
		findSymSection(ld.Segdata, ".noptrdata"),
		findSymSection(ld.Segdata, ".data"),
	}

	w := new(bytes.Buffer)
	writeUleb128(w, uint32(len(sections))) // number of data entries

	for _, sec := range sections {
		writeUleb128(w, 0) // memidx
		w.WriteByte(0x41)  // i32.const
		writeSleb128(w, int32(sec.Vaddr))
		w.WriteByte(0x0b) // end
		writeUleb128(w, uint32(sec.Length))

		mainOut := ctxt.Out
		ctxt.Out = ld.NewOutBuf(bufio.NewWriter(w))
		ld.Datblk(ctxt, int64(sec.Vaddr), int64(sec.Length))
		ctxt.Out.Flush()
		ctxt.Out = mainOut
	}

	return w.Bytes()
}

func findSymSection(segment sym.Segment, name string) *sym.Section {
	for _, sec := range segment.Sections {
		if sec.Name == name {
			return sec
		}
	}
	panic("section not found")
}

var nameRegexp = regexp.MustCompile(`[^\w\.]`)

func genNameSec(fns []*wasmFunc) []byte {
	w := new(bytes.Buffer)
	writeName(w, "name")

	w2 := new(bytes.Buffer)
	writeUleb128(w2, uint32(len(fns)))
	for i, fn := range fns {
		writeUleb128(w2, uint32(i))
		writeName(w2, fn.Name)
	}

	w.WriteByte(0x01) // function names
	writeUleb128(w, uint32(w2.Len()))
	w.Write(w2.Bytes())

	return w.Bytes()
}

func writeName(w *bytes.Buffer, name string) {
	writeUleb128(w, uint32(len(name)))
	w.Write([]byte(name))
}

func writeUleb128(w *bytes.Buffer, v uint32) {
	for {
		c := uint8(v & 0x7f)
		v >>= 7
		if v != 0 {
			c |= 0x80
		}
		w.WriteByte(c)
		if c&0x80 == 0 {
			break
		}
	}
}

func writeSleb128(w *bytes.Buffer, v int32) {
	for {
		c := uint8(v & 0x7f)
		s := uint8(v & 0x40)
		v >>= 7
		if (v != -1 || s == 0) && (v != 0 || s != 0) {
			c |= 0x80
		}
		w.WriteByte(c)
		if c&0x80 == 0 {
			break
		}
	}
}
