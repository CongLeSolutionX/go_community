// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Writing of Go object files.

// TODO: it may make sense to split this to a separate
// package.

package obj

import (
	"bytes"
	"cmd/internal/bio"
	"cmd/internal/objabi"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"unsafe"
)

// New object file format.
//
//    Header struct {
//       Magic   [...]byte   // "\x00go114ld"
//       // TODO: Fingerprint
//       Offsets [...]uint32 // byte offset of each block below
//    }
//
//    Strings [...]struct {
//       Len  uint32
//       Data [...]byte
//    }
//
//    PkgIndex [...]stringOff // TODO: add fingerprints
//
//    SymbolDefs [...]struct {
//       Name stringOff
//       ABI  uint16
//       Type uint8
//       Flag uint8
//       Size uint32
//    }
//    NonPkgDefs [...]struct { // non-pkg symbol definitions
//       ... // same as SymbolDefs
//    }
//    NonPkgRefs [...]struct { // non-pkg symbol references
//       ... // same as SymbolDefs
//    }
//
//    RelocIndex [...]uint32 // index to Relocs
//    AuxIndex   [...]uint32 // index to Aux
//    DataIndex  [...]uint32 // offset to Data
//
//    Relocs [...]struct {
//       Off  int32
//       Size uint8
//       Type uint8
//       Add  int64
//       Sym  symRef
//    }
//
//    Aux [...]struct {
//       Type uint8
//       Sym  symRef
//    }
//
//    Data   [...]byte
//    Pcdata [...]byte
//
// stringOff is a uint32 (?) offset that points to the corresponding
// string, which is a uint32 length followed by that number of bytes.
//
// symRef is struct { PkgIdx, SymIdx uint32 }.
//
// Slice type (e.g. []symRef) is encoded as a length prefix (uint32)
// followed by that number of elements.
//
// The O types below correspond to the encoded data structure in the
// object file. (If we split this to a separate package, the names
// don't need to have the O.)

// Symbol indexing.
//
// Each symbol is referenced with a pair of indices, { PkgIdx, SymIdx },
// as the symRef struct above.
//
// PkgIdx is either a predeclared index (see link.go:/PkgIdxInvalid) or
// an index of an imported package. For the latter case, PkgIdx-PkgIdxRefBase
// is the index of the package in the PkgIndex array.
//
// SymIdx is the index of the symbol in the given package.
// - If PkgIdx is PkgIdxSelf, SymIdx-1 is the index of the symbol in the
//   SymbolDefs array.
// - If PkgIdx is PkgIdxNone, SymIdx-1 is the index of the symbol in the
//   NonPkgDefs array (could natually overflow to NonPkgRefs array).
// - Otherwise, SymIdx-1 is the index of the symbol in some other package's
//   SymbolDefs array.
//
// {0, 0} represents a nil symbol. Otherwise neither index should be 0.
//
// RelocIndex, AuxIndex, and DataIndex contains indices/offsets to
// Relocs/Aux/Data blocks, one element per symbol, first for all the
// defined symbols, then all the defined non-package symbols, in the
// same order of SymbolDefs/NonPkgDefs arrays. So they can be accessed
// by index. For the i-th symbol, its relocations are the RelocIndex[i]-th
// (inclusive) to RelocIndex[i+1]-th (exclusive) elements in the Relocs
// array. Aux/Data are likewise. (The index is 0-based.)

// Auxiliary symbols.
//
// Each symbol may (or may not) be associated with a number of auxiliary
// symbols. They are described in the Aux block. See OAux struct below.
// Currently a symbol's Gotype and FuncInfo are auxiliary symbols. We
// may make use of aux symbols in more cases, e.g. DWARF symbols.

// Blocks
const (
	BlkPkgIdx = iota
	BlkSymdef
	BlkNonpkgdef
	BlkNonpkgref
	BlkRelocIdx
	BlkAuxIdx
	BlkDataIdx
	BlkReloc
	BlkAux
	BlkData
	BlkPcdata
	NBlk
)

// File header.
// TODO: probably no need to export this.
type OHeader struct {
	Magic   string
	Offsets [NBlk]uint32
}

const magic = "\x00go114LD"

func (h *OHeader) Write(w *Writer) {
	w.RawString(h.Magic)
	for _, x := range h.Offsets {
		w.Uint32(x)
	}
}

func (h *OHeader) Read(r *Reader) error {
	b := r.BytesAt(0, len(magic))
	h.Magic = string(b)
	if h.Magic != magic {
		return errors.New("wrong magic, not a Go object file")
	}
	off := uint32(len(h.Magic))
	for i := range h.Offsets {
		h.Offsets[i] = r.Uint32At(off)
		off += 4
	}
	return nil
}

func (h *OHeader) Size() int {
	return len(h.Magic) + 4*len(h.Offsets)
}

// Symbol definition.
type OSym struct {
	Name string
	ABI  uint16
	Type uint8
	Flag uint8
	Siz  uint32
}

const SymABIstatic = ^uint16(0)

const (
	SymFlagDupok = 1 << iota
	SymFlagLocal
	SymFlagTypelink
)

func (s *OSym) Write(w *Writer) {
	w.StringRef(s.Name)
	w.Uint16(s.ABI)
	w.Uint8(s.Type)
	w.Uint8(s.Flag)
	w.Uint32(s.Siz)
}

func (s *OSym) Read(r *Reader, off uint32) {
	s.Name = r.StringRef(off)
	s.ABI = r.Uint16At(off + 4)
	s.Type = r.Uint8At(off + 6)
	s.Flag = r.Uint8At(off + 7)
	s.Siz = r.Uint32At(off + 8)
}

func (s *OSym) Size() int {
	return 4 + 2 + 1 + 1 + 4
}

// Symbol reference.
type OSymRef struct {
	PkgIdx uint32
	SymIdx uint32
}

func (s *OSymRef) Write(w *Writer) {
	w.Uint32(s.PkgIdx)
	w.Uint32(s.SymIdx)
}

func (s *OSymRef) Read(r *Reader, off uint32) {
	s.PkgIdx = r.Uint32At(off)
	s.SymIdx = r.Uint32At(off + 4)
}

func (s *OSymRef) Size() int {
	return 4 + 4
}

// Relocation.
type OReloc struct {
	Off  int32
	Siz  uint8
	Type uint8
	Add  int64
	Sym  OSymRef
}

func (r *OReloc) Write(w *Writer) {
	w.Uint32(uint32(r.Off))
	w.Uint8(r.Siz)
	w.Uint8(r.Type)
	w.Uint64(uint64(r.Add))
	r.Sym.Write(w)
}

func (o *OReloc) Read(r *Reader, off uint32) {
	o.Off = r.Int32At(off)
	o.Siz = r.Uint8At(off + 4)
	o.Type = r.Uint8At(off + 5)
	o.Add = r.Int64At(off + 6)
	o.Sym.Read(r, off+14)
}

func (r *OReloc) Size() int {
	return 4 + 1 + 1 + 8 + r.Sym.Size()
}

// Aux symbol info.
type OAux struct {
	Type uint8
	Sym  OSymRef
}

// Aux Type
const (
	AuxGotype = iota
	AuxFuncInfo
	AuxFuncdata

	// TODO: more. DWARF? Pcdata?
)

func (a *OAux) Write(w *Writer) {
	w.Uint8(a.Type)
	a.Sym.Write(w)
}

func (a *OAux) Read(r *Reader, off uint32) {
	a.Type = r.Uint8At(off)
	a.Sym.Read(r, off+1)
}

func (a *OAux) Size() int {
	return 1 + a.Sym.Size()
}

// Entry point of writing new object file.
func WriteObjFile2(ctxt *Link, b *bio.Writer, pkgpath string) {
	genFuncInfoSyms(ctxt)

	w := Writer{
		ctxt:    ctxt,
		wr:      b,
		pkgpath: objabi.PathToPrefix(pkgpath),
	}

	start := w.wr.Offset()
	w.init()

	// Header
	// We just reserve the space. We'll fill in the offsets later.
	h := OHeader{Magic: magic}
	h.Write(&w)

	// String table
	w.StringTable()

	// Package references
	h.Offsets[BlkPkgIdx] = w.off
	for _, pkg := range w.pkglist {
		w.StringRef(pkg)
	}

	// Symbol definitions
	h.Offsets[BlkSymdef] = w.off
	for _, s := range ctxt.defs[1:] {
		w.Sym(s)
	}

	// Non-pkg symbol definitions
	h.Offsets[BlkNonpkgdef] = w.off
	for _, s := range ctxt.nonpkgdefs[1:] {
		w.Sym(s)
	}

	// Non-pkg symbol references
	h.Offsets[BlkNonpkgref] = w.off
	for _, s := range ctxt.nonpkgrefs {
		w.Sym(s)
	}

	// Reloc indexes
	h.Offsets[BlkRelocIdx] = w.off
	nreloc := uint32(0)
	lists := [][]*LSym{ctxt.defs[1:], ctxt.nonpkgdefs[1:]}
	for _, list := range lists {
		for _, s := range list {
			w.Uint32(nreloc)
			nreloc += uint32(len(s.R))
		}
	}
	w.Uint32(nreloc)

	// Symbol Info indexes
	h.Offsets[BlkAuxIdx] = w.off
	naux := uint32(0)
	for _, list := range lists {
		for _, s := range list {
			w.Uint32(naux)
			if s.Gotype != nil {
				naux++
			}
			if s.Func != nil {
				// FuncInfo is an aux symbol, each Funcdata is an aux symbol
				naux += 1 + uint32(len(s.Func.Pcln.Funcdata))
			}
		}
	}
	w.Uint32(naux)

	// Data indexes
	h.Offsets[BlkDataIdx] = w.off
	dataOff := uint32(0)
	for _, list := range lists {
		for _, s := range list {
			w.Uint32(dataOff)
			dataOff += uint32(len(s.P))
		}
	}
	w.Uint32(dataOff)

	// Relocs
	h.Offsets[BlkReloc] = w.off
	for _, list := range lists {
		for _, s := range list {
			for i := range s.R {
				w.Reloc(&s.R[i])
			}
		}
	}

	// Aux symbol info
	h.Offsets[BlkAux] = w.off
	for _, list := range lists {
		for _, s := range list {
			w.Aux(s)
		}
	}

	// Data
	h.Offsets[BlkData] = w.off
	for _, list := range lists {
		for _, s := range list {
			w.Bytes(s.P)
		}
	}

	// Pcdata
	h.Offsets[BlkPcdata] = w.off
	for _, s := range ctxt.Text { // iteration order must match genFuncInfoSyms
		if s.Func != nil {
			pc := &s.Func.Pcln
			w.Bytes(pc.Pcsp.P)
			w.Bytes(pc.Pcfile.P)
			w.Bytes(pc.Pcline.P)
			w.Bytes(pc.Pcinline.P)
			for i := range pc.Pcdata {
				w.Bytes(pc.Pcdata[i].P)
			}
		}
	}

	// Fix up block offsets in the header
	end := start + int64(w.off)
	w.wr.MustSeek(start, 0)
	w.off = 0
	h.Write(&w)
	w.wr.MustSeek(end, 0)
}

type Writer struct {
	wr      *bio.Writer
	ctxt    *Link
	pkgpath string // the package import path (escaped), "" if unknown

	pkglist []string // list of packages referenced, indexed by ctxt.pkgIdx

	stringMap map[string]uint32

	off uint32 // running offset

	auxDataOff uint32 // running offset for tracking auxdata
}

// prepare package index list
func (w *Writer) init() {
	w.pkglist = make([]string, len(w.ctxt.pkgIdx))
	for pkg, i := range w.ctxt.pkgIdx {
		w.pkglist[i-PkgIdxRefBase] = pkg
	}

	// Also make sure imported packages appear in the list (even if no symbol is referenced).
	for _, pkg := range w.ctxt.Imports {
		if _, ok := w.ctxt.pkgIdx[pkg]; !ok {
			w.pkglist = append(w.pkglist, pkg)
		}
	}
}

func (w *Writer) StringTable() {
	w.stringMap = make(map[string]uint32)
	w.addString("")
	for _, pkg := range w.ctxt.Imports {
		w.addString(pkg)
	}
	for _, pkg := range w.pkglist {
		w.addString(pkg)
	}
	w.ctxt.traverseSyms(traverseAll, func(s *LSym) {
		w.addString(s.Name)
	})
	w.ctxt.traverseSyms(traverseDefs, func(s *LSym) {
		if s.Type != objabi.STEXT {
			return
		}
		pc := &s.Func.Pcln
		for _, f := range pc.File {
			w.addString(f)
		}
		for _, call := range pc.InlTree.nodes {
			f, _ := linkgetlineFromPos(w.ctxt, call.Pos)
			w.addString(f)
		}
	})
}

func (w *Writer) addString(s string) {
	if _, ok := w.stringMap[s]; ok {
		return
	}
	w.stringMap[s] = w.off
	w.Uint32(uint32(len(s)))
	w.RawString(s)
}

func (w *Writer) StringRef(s string) {
	off, ok := w.stringMap[s]
	if !ok {
		panic(fmt.Sprintf("writeStringRef: string not added: %q", s))
	}
	w.Uint32(off)
}

func (w *Writer) RawString(s string) {
	w.wr.WriteString(s)
	w.off += uint32(len(s))
}

func (w *Writer) Bytes(s []byte) {
	w.wr.Write(s)
	w.off += uint32(len(s))
}

func (w *Writer) Uint64(x uint64) {
	var b [8]byte
	binary.LittleEndian.PutUint64(b[:], x)
	w.wr.Write(b[:])
	w.off += 8
}

func (w *Writer) Uint32(x uint32) {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], x)
	w.wr.Write(b[:])
	w.off += 4
}

func (w *Writer) Uint16(x uint16) {
	var b [2]byte
	binary.LittleEndian.PutUint16(b[:], x)
	w.wr.Write(b[:])
	w.off += 2
}

func (w *Writer) Uint8(x uint8) {
	w.wr.WriteByte(x)
	w.off++
}

func (w *Writer) Sym(s *LSym) {
	abi := uint16(s.ABI())
	if s.Static() {
		abi = SymABIstatic
	}
	flag := uint8(0)
	if s.DuplicateOK() {
		flag |= SymFlagDupok
	}
	if s.Local() {
		flag |= SymFlagLocal
	}
	if s.MakeTypelink() {
		flag |= SymFlagTypelink
	}
	o := OSym{
		Name: s.Name,
		ABI:  abi,
		Type: uint8(s.Type),
		Flag: flag,
		Siz:  uint32(s.Size),
	}
	o.Write(w)
}

func makeSymRef(s *LSym) OSymRef {
	if s == nil {
		return OSymRef{0, 0}
	}
	if s.PkgIdx == 0 || s.SymIdx == 0 {
		fmt.Printf("unindexed symbol reference: %v\n", s)
		panic("unindexed symbol reference")
	}
	return OSymRef{uint32(s.PkgIdx), uint32(s.SymIdx)}
}

func (w *Writer) Reloc(r *Reloc) {
	o := OReloc{
		Off:  r.Off,
		Siz:  r.Siz,
		Type: uint8(r.Type),
		Add:  r.Add,
		Sym:  makeSymRef(r.Sym),
	}
	o.Write(w)
}

func (w *Writer) Aux(s *LSym) {
	if s.Gotype != nil {
		o := OAux{
			Type: AuxGotype,
			Sym:  makeSymRef(s.Gotype),
		}
		o.Write(w)
	}
	if s.Func != nil {
		o := OAux{
			Type: AuxFuncInfo,
			Sym:  makeSymRef(s.Func.FuncInfoSym),
		}
		o.Write(w)

		for _, d := range s.Func.Pcln.Funcdata {
			o := OAux{
				Type: AuxFuncdata,
				Sym:  makeSymRef(d),
			}
			o.Write(w)
		}
	}
}

type Reader struct {
	b        []byte // mmapped bytes, if not nil
	readonly bool   // whether b is backed with read-only memory

	rd    io.ReaderAt
	start uint32
	h     OHeader // keep block offsets
}

func NewReader(rd io.ReaderAt, off uint32) *Reader {
	r := &Reader{rd: rd, start: off}
	err := r.h.Read(r)
	if err != nil {
		return nil
	}
	return r
}

func NewReaderFromBytes(b []byte, readonly bool) *Reader {
	r := &Reader{b: b, readonly: readonly, rd: bytes.NewReader(b), start: 0}
	err := r.h.Read(r)
	if err != nil {
		return nil
	}
	return r
}

func (r *Reader) BytesAt(off uint32, len int) []byte {
	if r.b != nil {
		return r.b[int(off) : int(off)+len]
	}
	b := make([]byte, len)
	_, err := r.rd.ReadAt(b[:], int64(r.start+off))
	if err != nil {
		panic("corrupted input")
	}
	return b
}

func (r *Reader) Uint64At(off uint32) uint64 {
	var b [8]byte
	n, err := r.rd.ReadAt(b[:], int64(r.start+off))
	if n != 8 || err != nil {
		panic("corrupted input")
	}
	return binary.LittleEndian.Uint64(b[:])
}

func (r *Reader) Int64At(off uint32) int64 {
	return int64(r.Uint64At(off))
}

func (r *Reader) Uint32At(off uint32) uint32 {
	var b [4]byte
	n, err := r.rd.ReadAt(b[:], int64(r.start+off))
	if n != 4 || err != nil {
		panic("corrupted input")
	}
	return binary.LittleEndian.Uint32(b[:])
}

func (r *Reader) Int32At(off uint32) int32 {
	return int32(r.Uint32At(off))
}

func (r *Reader) Uint16At(off uint32) uint16 {
	var b [2]byte
	n, err := r.rd.ReadAt(b[:], int64(r.start+off))
	if n != 2 || err != nil {
		panic("corrupted input")
	}
	return binary.LittleEndian.Uint16(b[:])
}

func (r *Reader) Uint8At(off uint32) uint8 {
	var b [1]byte
	n, err := r.rd.ReadAt(b[:], int64(r.start+off))
	if n != 1 || err != nil {
		panic("corrupted input")
	}
	return b[0]
}

func (r *Reader) StringAt(off uint32) string {
	l := r.Uint32At(off)
	if r.b != nil {
		return toString(r.b[int(off+4):int(off+4+l)])
	}
	b := make([]byte, l)
	n, err := r.rd.ReadAt(b, int64(r.start+off+4))
	if n != int(l) || err != nil {
		panic("corrupted input")
	}
	return string(b)
}

type stringHeader struct {
	str unsafe.Pointer
	len int
}

func toString(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	ss := stringHeader{str: unsafe.Pointer(&b[0]), len: len(b)}
	s := *(*string)(unsafe.Pointer(&ss))
	return s
}

func (r *Reader) StringRef(off uint32) string {
	return r.StringAt(r.Uint32At(off))
}

func (r *Reader) Pkglist() []string {
	n := (r.h.Offsets[BlkPkgIdx+1] - r.h.Offsets[BlkPkgIdx]) / 4
	s := make([]string, n)
	for i := range s {
		off := r.h.Offsets[BlkPkgIdx] + uint32(i)*4
		s[i] = r.StringRef(off)
	}
	return s
}

func (r *Reader) NSym() int {
	symsiz := (&OSym{}).Size()
	return int(r.h.Offsets[BlkSymdef+1]-r.h.Offsets[BlkSymdef]) / symsiz
}

func (r *Reader) NNonpkgdef() int {
	symsiz := (&OSym{}).Size()
	return int(r.h.Offsets[BlkNonpkgdef+1]-r.h.Offsets[BlkNonpkgdef]) / symsiz
}

func (r *Reader) NNonpkgref() int {
	symsiz := (&OSym{}).Size()
	return int(r.h.Offsets[BlkNonpkgref+1]-r.h.Offsets[BlkNonpkgref]) / symsiz
}

// SymOff returns the offset of the i-th symbol.
func (r *Reader) SymOff(i int) uint32 {
	symsiz := (&OSym{}).Size()
	return r.h.Offsets[BlkSymdef] + uint32(i*symsiz)
}

// NReloc returns the number of relocations of the i-th symbol.
func (r *Reader) NReloc(i int) int {
	relocIdxOff := r.h.Offsets[BlkRelocIdx] + uint32(i*4)
	return int(r.Uint32At(relocIdxOff+4) - r.Uint32At(relocIdxOff))
}

// RelocOff returns the offset of the j-th relocation of the i-th symbol.
func (r *Reader) RelocOff(i int, j int) uint32 {
	relocIdxOff := r.h.Offsets[BlkRelocIdx] + uint32(i*4)
	relocIdx := r.Uint32At(relocIdxOff)
	relocsiz := (&OReloc{}).Size()
	return r.h.Offsets[BlkReloc] + (relocIdx+uint32(j))*uint32(relocsiz)
}

// NAux returns the number of aux symbols of the i-th symbol.
func (r *Reader) NAux(i int) int {
	auxIdxOff := r.h.Offsets[BlkAuxIdx] + uint32(i*4)
	return int(r.Uint32At(auxIdxOff+4) - r.Uint32At(auxIdxOff))
}

// AuxOff returns the offset of the j-th aux symbol of the i-th symbol.
func (r *Reader) AuxOff(i int, j int) uint32 {
	auxIdxOff := r.h.Offsets[BlkAuxIdx] + uint32(i*4)
	auxIdx := r.Uint32At(auxIdxOff)
	auxsiz := (&OAux{}).Size()
	return r.h.Offsets[BlkAux] + (auxIdx+uint32(j))*uint32(auxsiz)
}

// DataOff returns the offset of the i-th symbol's data.
func (r *Reader) DataOff(i int) uint32 {
	dataIdxOff := r.h.Offsets[BlkDataIdx] + uint32(i*4)
	return r.h.Offsets[BlkData] + r.Uint32At(dataIdxOff)
}

// DataSize returns the size of the i-th symbol's data.
func (r *Reader) DataSize(i int) int {
	return int(r.DataOff(i+1) - r.DataOff(i))
}

// AuxDataBase returns the base offset of the aux data block.
func (r *Reader) PcdataBase() uint32 {
	return r.h.Offsets[BlkPcdata]
}

// ReadOnly returns whether r.BytesAt returns read-only bytes.
func (r *Reader) ReadOnly() bool {
	return r.readonly
}

// FuncInfo is serialized as a symbol (aux symbol). The symbol data is
// the binary encoding of the struct below.
//
// TODO: make each pcdata a separate symbol?
//
//   FuncInfo struct {
//      NoSplit uint8
//      Flags   uint8 // TODO: compact flags: Nosplit should be a bit, Shared shouldn't be per symbol
//
//      Args   uint32
//      Locals uint32
//
//      Pcsp        uint32 // offset to auxdata // TODO: have a better way to organize them...
//      Pcfile      uint32
//      Pcline      uint32
//      Pcinline    uint32
//      Pcdata      []uint32
//      PcdataEnd   uint32 // offset of the end of the last Pcdata element, so we know where it ends
//      Funcdataoff []uint32 // What is this?
//      File        []stringOff
//
//      // TODO:
//      //InlTree
//      //Autom []symRef // annoying
//   }
//
type OFuncInfo struct {
	NoSplit uint8
	Flags   uint8

	Args   uint32
	Locals uint32

	Pcsp        uint32
	Pcfile      uint32
	Pcline      uint32
	Pcinline    uint32
	Pcdata      []uint32
	PcdataEnd   uint32
	Funcdataoff []uint32
	File        []OSymRef

	// TODO:
	//InlTree
	//Autom []OSymRef // annoying
}

const (
	FuncFlagLeaf = 1 << iota
	FuncFlagCFunc
	FuncFlagReflectMethod
	FuncFlagShared // This is really silly
	FuncFlagTopFrame
)

func (a *OFuncInfo) Write(w *bytes.Buffer) {
	w.WriteByte(a.NoSplit)
	w.WriteByte(a.Flags)

	var b [4]byte
	writeUint32 := func(x uint32) {
		binary.LittleEndian.PutUint32(b[:], x)
		w.Write(b[:])
	}

	writeUint32(a.Args)
	writeUint32(a.Locals)

	writeUint32(a.Pcsp)
	writeUint32(a.Pcfile)
	writeUint32(a.Pcline)
	writeUint32(a.Pcinline)
	writeUint32(uint32(len(a.Pcdata)))
	for _, x := range a.Pcdata {
		writeUint32(x)
	}
	writeUint32(a.PcdataEnd)
	writeUint32(uint32(len(a.Funcdataoff)))
	for _, x := range a.Funcdataoff {
		writeUint32(x)
	}
	writeUint32(uint32(len(a.File)))
	for _, f := range a.File {
		writeUint32(f.PkgIdx)
		writeUint32(f.SymIdx)
	}

	//TODO: InlTree, Autom
}

func (a *OFuncInfo) Read(b []byte) {
	a.NoSplit = b[0]
	a.Flags = b[1]
	b = b[2:]

	readUint32 := func() uint32 {
		x := binary.LittleEndian.Uint32(b)
		b = b[4:]
		return x
	}

	a.Args = readUint32()
	a.Locals = readUint32()

	a.Pcsp = readUint32()
	a.Pcfile = readUint32()
	a.Pcline = readUint32()
	a.Pcinline = readUint32()
	pcdatalen := readUint32()
	a.Pcdata = make([]uint32, pcdatalen)
	for i := range a.Pcdata {
		a.Pcdata[i] = readUint32()
	}
	a.PcdataEnd = readUint32()
	funcdataofflen := readUint32()
	a.Funcdataoff = make([]uint32, funcdataofflen)
	for i := range a.Funcdataoff {
		a.Funcdataoff[i] = readUint32()
	}
	filelen := readUint32()
	a.File = make([]OSymRef, filelen)
	for i := range a.File {
		a.File[i] = OSymRef{readUint32(), readUint32()}
	}

	//TODO: InlTree, Autom
}

// generate symbols for FuncInfo.
func genFuncInfoSyms(ctxt *Link) {
	infosyms := make([]*LSym, 0, len(ctxt.Text))
	var pcdataoff uint32
	var b bytes.Buffer
	symidx := int32(len(ctxt.defs))
	for _, s := range ctxt.Text {
		if s.Func == nil {
			continue
		}
		nosplit := uint8(0)
		if s.NoSplit() {
			nosplit = 1
		}
		flags := uint8(0)
		if s.Leaf() {
			flags |= FuncFlagLeaf
		}
		if s.CFunc() {
			flags |= FuncFlagCFunc
		}
		if s.ReflectMethod() {
			flags |= FuncFlagReflectMethod
		}
		if ctxt.Flag_shared { // This is really silly
			flags |= FuncFlagShared
		}
		if s.TopFrame() {
			flags |= FuncFlagTopFrame
		}
		o := OFuncInfo{
			NoSplit: nosplit,
			Flags:   flags,
			Args:    uint32(s.Func.Args),
			Locals:  uint32(s.Func.Locals),
		}
		pc := &s.Func.Pcln
		o.Pcsp = pcdataoff
		pcdataoff += uint32(len(pc.Pcsp.P))
		o.Pcfile = pcdataoff
		pcdataoff += uint32(len(pc.Pcfile.P))
		o.Pcline = pcdataoff
		pcdataoff += uint32(len(pc.Pcline.P))
		o.Pcinline = pcdataoff
		pcdataoff += uint32(len(pc.Pcinline.P))
		o.Pcdata = make([]uint32, len(pc.Pcdata))
		for i, pcd := range pc.Pcdata {
			o.Pcdata[i] = pcdataoff
			pcdataoff += uint32(len(pcd.P))
		}
		o.PcdataEnd = pcdataoff
		o.Funcdataoff = make([]uint32, len(pc.Funcdataoff))
		for i, x := range pc.Funcdataoff {
			o.Funcdataoff[i] = uint32(x)
		}
		o.File = make([]OSymRef, len(pc.File))
		for i, f := range pc.File {
			fsym := ctxt.Lookup(f)
			o.File[i] = makeSymRef(fsym)
		}

		o.Write(&b)
		isym := &LSym{
			Type:   objabi.SDATA, // for now, I don't think it matters
			PkgIdx: PkgIdxSelf,
			SymIdx: symidx,
			P:      append([]byte(nil), b.Bytes()...),
		}
		symidx++
		infosyms = append(infosyms, isym)
		s.Func.FuncInfoSym = isym
		b.Reset()
	}
	ctxt.defs = append(ctxt.defs, infosyms...)
}
