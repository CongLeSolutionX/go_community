// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Writing of Go object files.

// TODO: it may make sense to split this to a separate
// package.

package obj

import (
	"cmd/internal/bio"
	"cmd/internal/objabi"
	"encoding/binary"
	"fmt"
	"io"
)

// New object file format.
//
//    Header struct {
//       Magic   [...]byte   // "\x00go114ld"
//       // TODO: Fingerprint
//       Offsets [...]uint32 // byte offset of each section below
//    }
//
//    Strings [...]struct {
//       Len  uint32 // ???
//       Data [...]byte
//    }
//
//    Autolib [...]stringOff // dependent libraries to load // TODO: add a fingerprint
//
//    PkgIndex [...]stringOff // TODO: combine with Autolib?
//
//    SymbolDefs [...]struct {
//       Name stringOff
//       ABI  uint16
//       Type uint8
//       Flag uint8
//    }
//    NonPkgSyms [...]struct { // defs + refs
//       ... // same as SymbolDefs
//    }
//
//    RelocIndex [...]uint32 // index to Relocs
//    InfoIndex  [...]uint32 // offset to SymInfo
//    DataIndex  [...]uint32 // offset to Data
//
//    Relocs [...]struct {
//       Off  int32
//       Size uint8
//       Type uint8
//       Add  int64
//       Sym  OSymRef
//    }
//
//    SymInfo [...]struct {
//       GoType  symRef
//       NoSplit uint8
//       Flags   uint8 // TODO: compact flags: Nosplit should be a bit, Shared shouldn't be per symbol
//
//       Args   uint32
//       Locals uint32
//
//       Pcsp        uint32 // offset to auxdata // TODO: have a better way to organize them...
//       Pcfile      uint32
//       Pcline      uint32
//       Pcinline    uint32
//       Pcdata      []uint32
//       Funcdata    []symRef
//       Funcdataoff []uint32 // What is this?
//       File        []stringOff
//
//       // TODO:
//       //InlTree
//       //Autom []symRef // annoying
//
//       // TODO: aux symbols
//    }
//
//    Data    [...]byte
//    AuxData [...]byte
//
//    Footer [...]byte // "\xffgo114ld"
//
// stringOff is a uint32 (?) offset that points to the corresponding
// a string, which is a uint32 length followed by that number of bytes.
//
// symRef is struct { PkgIdx, SymIdx uint32 }.
//
// Slice type (e.g. []symRef) is encoded as a length prefix (uint32)
// followed by that number of elements.
//
// The O types below correspond to the encoded data structure in the
// object file. (If we split this to a separate package, the names
// don't need to have the O.)

// Sections
const (
	SectAutolib = iota
	SectPkgIdx
	SectSymdef
	SectNonpkgsym
	SectRelocIdx
	SectInfoIdx
	SectDataIdx
	SectReloc
	SectInfo
	SectData
	SectAuxData
	NSect
)

// File header.
type OHeader struct {
	Magic   string
	Offsets [NSect]uint32
}

func (h *OHeader) Write(w *Writer) {
	w.RawString(h.Magic)
	for _, x := range h.Offsets {
		w.Uint32(x)
	}
}

func (h *OHeader) Read(r *Reader) {
	var b [8]byte // TODO: length is hard-coded
	r.Bytes(b[:])
	h.Magic = string(b[:])
	for i := range h.Offsets {
		h.Offsets[i] = r.Uint32()
	}
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
}

func (s *OSym) Write(w *Writer) {
	w.StringRef(s.Name)
	w.Uint16(s.ABI)
	w.Uint8(s.Type)
	w.Uint8(s.Flag)
}

func (s *OSym) Read(r *Reader, off uint32) {
	s.Name = r.StringAt(r.Uint32At(off))
	s.ABI = r.Uint16At(off + 4)
	s.Type = r.Uint8At(off + 6)
	s.Flag = r.Uint8At(off + 7)
}

func (s *OSym) Size() int {
	return 4 + 2 + 1 + 1
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

// Extra symbol info: funcinfo, etc.
type OSymInfo struct {
	GoType     OSymRef
	*OTextInfo // non-nil only for TEXT symbols
}

// symbol info for TEXT symbols
type OTextInfo struct {
	NoSplit uint8
	Flags   uint8

	Args   uint32
	Locals uint32

	Pcsp        uint32
	Pcfile      uint32
	Pcline      uint32
	Pcinline    uint32
	Pcdata      []uint32
	Funcdata    []OSymRef
	Funcdataoff []uint32
	File        []string

	//TODO: InlTree

	Autom []OSymRef // annoying
}

func (a *OSymInfo) Write(w *Writer) {
	a.GoType.Write(w)
	if a.OTextInfo == nil {
		return
	}
	w.Uint8(a.NoSplit)
	w.Uint8(a.Flags)

	w.Uint32(a.Args)
	w.Uint32(a.Locals)

	w.Uint32(a.Pcsp)
	w.Uint32(a.Pcfile)
	w.Uint32(a.Pcline)
	w.Uint32(a.Pcinline)
	w.Uint32(uint32(len(a.Pcdata)))
	for _, x := range a.Pcdata {
		w.Uint32(x)
	}
	w.Uint32(uint32(len(a.Funcdata)))
	for _, s := range a.Funcdata {
		s.Write(w)
	}
	w.Uint32(uint32(len(a.Funcdataoff)))
	for _, x := range a.Funcdataoff {
		w.Uint32(x)
	}
	w.Uint32(uint32(len(a.File)))
	for _, f := range a.File {
		w.StringRef(f)
	}

	//TODO: InlTree

	// w.Uint32(uint32(len(a.Autom)))
	// for _, s := range a.Autom {
	// 	s.Write(w)
	// }
}

func (a *OSymInfo) Read() {
	panic("not implemented")
}

/*
func (a *OSymInfo) Size() int {
	if a.OTextInfo == nil {
		return a.GoType.Size()
	}
	return a.GoType.Size() + 1 + 1 + 4 + 4 + 4 + 4 + 4 + 4 + (4 + 4*len(a.Pcdata)) + (4 + (&OSymRef{}).Size()*len(a.Funcdata)) + (4 + 4*len(a.Funcdataoff)) + (4 + 4*len(a.File))
	// TODO: Autom: + (4 + (&OSymRef{}).Size()*len(a.Autom))
}
*/

// Entry point of writing new object file.
func WriteObjFile2(ctxt *Link, b *bio.Writer, pkgpath string) {
	w := Writer{
		ctxt:    ctxt,
		wr:      b,
		pkgpath: objabi.PathToPrefix(pkgpath),
	}

	start := w.wr.Offset()
	w.init()

	// Header
	// We just reserve the space. We'll fill in the offsets later.
	h := OHeader{Magic: "\x00go114ld"}
	h.Write(&w)

	// String table
	w.StringTable()

	// Autolib
	h.Offsets[SectAutolib] = w.off
	for _, pkg := range ctxt.Imports {
		w.StringRef(pkg)
	}

	// Package references
	h.Offsets[SectPkgIdx] = w.off
	for _, pkg := range w.pkglist {
		w.StringRef(pkg)
	}

	// Symbol definitions
	h.Offsets[SectSymdef] = w.off
	for _, s := range ctxt.defs[1:] {
		w.Sym(s)
	}

	// Non-pkg symbol defs + refs
	h.Offsets[SectNonpkgsym] = w.off
	for _, s := range ctxt.nonpkgsyms[1:] {
		w.Sym(s)
	}

	// Reloc indexes
	h.Offsets[SectRelocIdx] = w.off
	nreloc := uint32(0)
	lists := [][]*LSym{ctxt.defs[1:], ctxt.nonpkgsyms[1:]}
	for _, list := range lists {
		for _, s := range list {
			w.Uint32(nreloc)
			nreloc += uint32(len(s.R))
		}
	}
	w.Uint32(nreloc)

	// Symbol Info indexes
	h.Offsets[SectInfoIdx] = w.off
	infoOff := uint32(0)
	for _, list := range lists {
		for _, s := range list {
			w.Uint32(infoOff)
			infoOff += uint32(OSymInfoSize(s))
		}
	}
	w.Uint32(infoOff)

	// Data indexes
	h.Offsets[SectDataIdx] = w.off
	dataOff := uint32(0)
	for _, list := range lists {
		for _, s := range list {
			w.Uint32(dataOff)
			dataOff += uint32(len(s.P))
		}
	}
	w.Uint32(dataOff)

	// Relocs
	h.Offsets[SectReloc] = w.off
	for _, list := range lists {
		for _, s := range list {
			for i := range s.R {
				w.Reloc(&s.R[i])
			}
		}
	}

	// Symbol info data
	h.Offsets[SectInfo] = w.off
	for _, list := range lists {
		for _, s := range list {
			w.Info(s)
		}
	}

	// Data
	h.Offsets[SectData] = w.off
	for _, list := range lists {
		for _, s := range list {
			w.Bytes(s.P)
		}
	}

	// AuxData
	h.Offsets[SectAuxData] = w.off
	for _, list := range lists {
		for _, s := range list {
			if s.Type == objabi.STEXT && s.Func != nil {
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
	}

	// Footer
	w.RawString("\xffgo114ld")

	// Fix up section offsets
	w.wr.MustSeek(start, 0)
	w.off = 0
	h.Write(&w)
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

// XXX prepare package index list
func (w *Writer) init() {
	w.pkglist = make([]string, len(w.ctxt.pkgIdx))
	for pkg, i := range w.ctxt.pkgIdx {
		w.pkglist[i-PkgIdxRefBase] = pkg
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
	w.ctxt.TraverseSyms(traverseAll, func(s *LSym) {
		if s == nil {
			return
		}
		w.addString(s.Name)
	})
	w.ctxt.TraverseSyms(traverseDefs, func(s *LSym) {
		if s == nil || s.Type != objabi.STEXT {
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
		abi = ^uint16(0)
	}
	flag := uint8(0)
	if s.DuplicateOK() {
		flag |= 1
	}
	if s.Local() {
		flag |= 1 << 1
	}
	if s.MakeTypelink() {
		flag |= 1 << 2
	}
	o := OSym{
		Name: s.Name,
		ABI:  abi,
		Type: uint8(s.Type),
		Flag: flag,
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

func (w *Writer) Info(s *LSym) {
	o := OSymInfo{GoType: makeSymRef(s.Gotype)}
	if s.Type == objabi.STEXT && s.Func != nil {
		nosplit := uint8(0)
		if s.NoSplit() {
			nosplit = 1
		}
		flags := uint8(0)
		if s.Leaf() {
			flags |= 1
		}
		if s.CFunc() {
			flags |= 1 << 1
		}
		if s.ReflectMethod() {
			flags |= 1 << 2
		}
		if w.ctxt.Flag_shared { // This is really silly
			flags |= 1 << 3
		}
		if s.TopFrame() {
			flags |= 1 << 4
		}
		t := OTextInfo{
			NoSplit: nosplit,
			Flags:   flags,
			Args:    uint32(s.Func.Args),
			Locals:  uint32(s.Func.Locals),
		}
		pc := &s.Func.Pcln
		t.Pcsp = w.auxDataOff
		w.auxDataOff += uint32(len(pc.Pcsp.P))
		t.Pcfile = w.auxDataOff
		w.auxDataOff += uint32(len(pc.Pcfile.P))
		t.Pcline = w.auxDataOff
		w.auxDataOff += uint32(len(pc.Pcline.P))
		t.Pcinline = w.auxDataOff
		w.auxDataOff += uint32(len(pc.Pcinline.P))
		t.Pcdata = make([]uint32, len(pc.Pcdata))
		for i, pcd := range pc.Pcdata {
			t.Pcdata[i] = w.auxDataOff
			w.auxDataOff += uint32(len(pcd.P))
		}
		t.Funcdata = make([]OSymRef, len(pc.Funcdata))
		for i, x := range pc.Funcdata {
			t.Funcdata[i] = makeSymRef(x)
		}
		t.Funcdataoff = make([]uint32, len(pc.Funcdataoff))
		for i, x := range pc.Funcdataoff {
			t.Funcdataoff[i] = uint32(x)
		}
		t.File = make([]string, len(pc.File))
		for i, f := range pc.File {
			t.File[i] = f
		}
		o.OTextInfo = &t
	}
	o.Write(w)
}

func OSymInfoSize(s *LSym) int {
	symrefsiz := (&OSymRef{}).Size()
	if s.Type != objabi.STEXT || s.Func == nil {
		return symrefsiz
	}
	pc := &s.Func.Pcln
	return symrefsiz + 1 + 1 + 4 + 4 + 4 + 4 + 4 + 4 + (4 + 4*len(pc.Pcdata)) + (4 + symrefsiz*len(pc.Funcdata)) + (4 + 4*len(pc.Funcdataoff)) + (4 + 4*len(pc.File))
	// TODO: Autom: + (4 + symrefsiz*len(s.Func.Autom))
}

type Reader struct {
	rd interface {
		io.ReadSeeker
		io.ReaderAt
	}
	start uint32
}

func NewReader(rd interface {
	io.ReadSeeker
	io.ReaderAt
}, off uint32) *Reader {
	return &Reader{rd: rd, start: off}
}

func (r *Reader) Bytes(b []byte) {
	_, err := r.rd.Read(b[:])
	if err != nil {
		panic(err)
	}
}

func (r *Reader) BytesAt(b []byte, off uint32) {
	_, err := r.rd.ReadAt(b[:], int64(r.start+off))
	if err != nil {
		panic("corrupted input")
	}
}

func (r *Reader) Uint32() uint32 {
	var b [4]byte
	n, err := r.rd.Read(b[:])
	if n != 4 || err != nil {
		panic("corrupted input")
	}
	return binary.LittleEndian.Uint32(b[:])
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
	b := make([]byte, l)
	n, err := r.rd.ReadAt(b, int64(r.start+off+4))
	if n != int(l) || err != nil {
		panic("corrupted input")
	}
	return string(b)
}
