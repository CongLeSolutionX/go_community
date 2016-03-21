// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ld

// Reading of Go object files.
//
// Originally, Go object files were Plan 9 object files, but no longer.
// Now they are more like standard object files, in that each symbol
// is defined by an associated memory image (bytes) and a list of
// relocations to apply during linking.
// See golang.org/s/go13linker for more background.
//
// An object file is layed out as a zip archive containing the
// following files:
//
//	- imports: sequence of strings giving dependencies
//	- symrefs: sequence of sybols referenced by the defined symbols
//	- data:    the content of the defined symbols
//	- symbols: sequence of defined symbols
//
// All integers are stored in a zigzag varint format.
// See golang.org/s/go12symtab for a definition.
//
// Strings are stored as an integer followed by that many bytes.
//
// A symbol reference is a name (string) followed by a version (int).
//
// Symbols point at each other using an index into the symbol
// reference sequence. In the symbol layout described below
// "symref index" stands for this index. Index 0 corresponds
// to a nil LSym pointer.
//
// Symbol defenitions consume data from the data file by specifing
// length consumed. Data is consumed by symbols in order.
//
// Each symbol is laid out as a sequence of ints
// holding the following fields (taken from LSym):
//
//	- type
//	- name & version [symref index]
//	- flags
//		1 dupok
//	- size
//	- gotype [symref index]
//	- p [data length]
//	- nr
//	- r [nr relocations, sorted by off]
//
// If type == STEXT, there are a few more fields:
//
//	- args
//	- locals
//	- nosplit
//	- flags
//		1<<0 leaf
//		1<<1 C function
//		1<<2 function may call reflect.Type.Method
//	- nlocal
//	- local [nlocal automatics]
//	- pcln [pcln table]
//
// Each relocation has the encoding:
//
//	- off
//	- siz
//	- type
//	- add
//	- sym [symref index]
//
// Each local has the encoding:
//
//	- asym [symref index]
//	- offset
//	- type
//	- gotype [symref index]
//
// The pcln table has the encoding:
//
//	- pcsp [data length]
//	- pcfile [data length]
//	- pcline [data length]
//	- npcdata
//	- pcdata [npcdata data blocks]
//	- nfuncdata
//	- funcdata [nfuncdata symref indexes]
//	- funcdatasym [nfuncdata ints]
//	- nfile
//	- file [nfile symref index]
//
// The file layout and meaning of type integers are architecture-independent.
//
// TODO(rsc): The file format is good for a first pass but needs work.
//	- There are SymID in the object file that should really just be strings.

import (
	"archive/zip"
	"bufio"
	"bytes"
	"cmd/internal/obj"
	"io"
	"log"
	"strconv"
	"strings"
)

var sections = []string{
	"imports",
	"symrefs",
	"data",
	"symbols",
}

func nextFile(reader *bufio.Reader, files []*zip.File, pn string) []*zip.File {
	file := files[0]
	zreader, err := file.Open()
	if err != nil {
		log.Fatalf("%s: error reading object file contents %s", pn, err)
	}
	reader.Reset(zreader)
	return files[1:]
}

func ldobjfile(ctxt *Link, f *io.SectionReader, pkg string, length int64, pn string) {
	zfile, err := zip.NewReader(f, length)
	if err != nil {
		log.Fatalf("%s: error reading object file %s", pn, err)
	}
	if len(zfile.File) != len(sections) {
		log.Fatalf("%s: unrecognized object file layout, number of sections is %d instead of %d", pn, len(zfile.File), len(sections))
	}
	for i := range sections {
		if sections[i] != zfile.File[i].Name {
			log.Fatalf("%s: unrecognized object file layout, expected %s found %s", pn, sections[i], zfile.File[i].Name)
		}
	}

	files := zfile.File
	reader := bufio.NewReader(nil)

	files = nextFile(reader, files, pn)
	var lib string
	for {
		_, err := reader.Peek(1)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("%s: peeking: %v", pn, err)
		}

		lib = rdstring(reader)
		addlib(ctxt, pkg, pn, lib)
	}

	files = nextFile(reader, files, pn)
	ctxt.CurRefs = []*LSym{nil} // zeroth ref is nil
	for {
		_, err := reader.Peek(1)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("%s: peeking: %v", pn, err)
		}
		readref(ctxt, reader, pkg, pn)
	}

	dataLength := files[0].UncompressedSize64
	data := make([]byte, dataLength)

	files = nextFile(reader, files, pn)
	io.ReadFull(reader, data)

	files = nextFile(reader, files, pn)
	for {
		_, err := reader.Peek(1)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("%s: peeking: %v", pn, err)
		}
		readsym(ctxt, reader, &data, pkg, pn)
	}
}

var dupSym = &LSym{Name: ".dup"}

func readsym(ctxt *Link, f *bufio.Reader, buf *[]byte, pkg string, pn string) {
	t := rdint(f)
	s := rdsym(ctxt, f, pkg)
	flags := rdint(f)
	dupok := flags&1 != 0
	local := flags&2 != 0
	size := rdint(f)
	typ := rdsym(ctxt, f, pkg)
	data := rddata(f, buf)
	nreloc := rdint(f)

	var dup *LSym
	if s.Type != 0 && s.Type != obj.SXREF {
		if (t == obj.SDATA || t == obj.SBSS || t == obj.SNOPTRBSS) && len(data) == 0 && nreloc == 0 {
			if s.Size < int64(size) {
				s.Size = int64(size)
			}
			if typ != nil && s.Gotype == nil {
				s.Gotype = typ
			}
			return
		}

		if (s.Type == obj.SDATA || s.Type == obj.SBSS || s.Type == obj.SNOPTRBSS) && len(s.P) == 0 && len(s.R) == 0 {
			goto overwrite
		}
		if s.Type != obj.SBSS && s.Type != obj.SNOPTRBSS && !dupok && !s.Attr.DuplicateOK() {
			log.Fatalf("duplicate symbol %s (types %d and %d) in %s and %s", s.Name, s.Type, t, s.File, pn)
		}
		if len(s.P) > 0 {
			dup = s
			s = dupSym
		}
	}

overwrite:
	s.File = pkg
	if dupok {
		s.Attr |= AttrDuplicateOK
	}
	if t == obj.SXREF {
		log.Fatalf("bad sxref")
	}
	if t == 0 {
		log.Fatalf("missing type for %s in %s", s.Name, pn)
	}
	if t == obj.SBSS && (s.Type == obj.SRODATA || s.Type == obj.SNOPTRBSS) {
		t = int(s.Type)
	}
	s.Type = int16(t)
	if s.Size < int64(size) {
		s.Size = int64(size)
	}
	s.Attr.Set(AttrLocal, local)
	if typ != nil {
		s.Gotype = typ
	}
	if dup != nil && typ != nil { // if bss sym defined multiple times, take type from any one def
		dup.Gotype = typ
	}
	s.P = data
	if nreloc > 0 {
		s.R = make([]Reloc, nreloc)
		var r *Reloc
		for i := 0; i < nreloc; i++ {
			r = &s.R[i]
			r.Off = rdint32(f)
			r.Siz = rduint8(f)
			r.Type = rdint32(f)
			r.Add = rdint64(f)
			r.Sym = rdsym(ctxt, f, pkg)
		}
	}

	if s.Type == obj.STEXT {
		s.Args = rdint32(f)
		s.Locals = rdint32(f)
		if rduint8(f) != 0 {
			s.Attr |= AttrNoSplit
		}
		flags := rdint(f)
		if flags&(1<<2) != 0 {
			s.Attr |= AttrReflectMethod
		}
		n := rdint(f)
		s.Autom = make([]Auto, n)
		for i := 0; i < n; i++ {
			s.Autom[i] = Auto{
				Asym:    rdsym(ctxt, f, pkg),
				Aoffset: rdint32(f),
				Name:    rdint16(f),
				Gotype:  rdsym(ctxt, f, pkg),
			}
		}

		s.Pcln = new(Pcln)
		pc := s.Pcln
		pc.Pcsp.P = rddata(f, buf)
		pc.Pcfile.P = rddata(f, buf)
		pc.Pcline.P = rddata(f, buf)
		n = rdint(f)
		pc.Pcdata = make([]Pcdata, n)
		pc.Npcdata = n
		for i := 0; i < n; i++ {
			pc.Pcdata[i].P = rddata(f, buf)
		}
		n = rdint(f)
		pc.Funcdata = make([]*LSym, n)
		pc.Funcdataoff = make([]int64, n)
		pc.Nfuncdata = n
		for i := 0; i < n; i++ {
			pc.Funcdata[i] = rdsym(ctxt, f, pkg)
		}
		for i := 0; i < n; i++ {
			pc.Funcdataoff[i] = rdint64(f)
		}
		n = rdint(f)
		pc.File = make([]*LSym, n)
		pc.Nfile = n
		for i := 0; i < n; i++ {
			pc.File[i] = rdsym(ctxt, f, pkg)
		}

		if dup == nil {
			if s.Attr.OnList() {
				log.Fatalf("symbol %s listed multiple times", s.Name)
			}
			s.Attr |= AttrOnList
			if ctxt.Etextp != nil {
				ctxt.Etextp.Next = s
			} else {
				ctxt.Textp = s
			}
			ctxt.Etextp = s
		}
	}
}

func readref(ctxt *Link, f *bufio.Reader, pkg string, pn string) {
	name := rdsymName(f, pkg)
	v := rdint(f)
	if v != 0 && v != 1 {
		log.Fatalf("invalid symbol version %d", v)
	}
	if v == 1 {
		v = ctxt.Version
	}
	s := Linklookup(ctxt, name, v)
	ctxt.CurRefs = append(ctxt.CurRefs, s)

	if s == nil || v != 0 {
		return
	}
	if s.Name[0] == '$' && len(s.Name) > 5 && s.Type == 0 && len(s.P) == 0 {
		x, err := strconv.ParseUint(s.Name[5:], 16, 64)
		if err != nil {
			log.Panicf("failed to parse $-symbol %s: %v", s.Name, err)
		}
		s.Type = obj.SRODATA
		s.Attr |= AttrLocal
		switch s.Name[:5] {
		case "$f32.":
			if uint64(uint32(x)) != x {
				log.Panicf("$-symbol %s too large: %d", s.Name, x)
			}
			Adduint32(ctxt, s, uint32(x))
		case "$f64.", "$i64.":
			Adduint64(ctxt, s, x)
		default:
			log.Panicf("unrecognized $-symbol: %s", s.Name)
		}
		s.Attr.Set(AttrReachable, false)
	}
	if strings.HasPrefix(s.Name, "runtime.gcbits.") {
		s.Attr |= AttrLocal
	}
}

func rdint64(f *bufio.Reader) int64 {
	uv := uint64(0)
	for shift := uint(0); ; shift += 7 {
		if shift >= 64 {
			log.Fatalf("corrupt input")
		}
		c, err := f.ReadByte()
		if err != nil {
			log.Fatalln("error reading input: ", err)
		}
		uv |= uint64(c&0x7F) << shift
		if c&0x80 == 0 {
			break
		}
	}

	return int64(uv>>1) ^ (int64(uint64(uv)<<63) >> 63)
}

func rdint(f *bufio.Reader) int {
	n := rdint64(f)
	if int64(int(n)) != n {
		log.Panicf("%v out of range for int", n)
	}
	return int(n)
}

func rdint32(f *bufio.Reader) int32 {
	n := rdint64(f)
	if int64(int32(n)) != n {
		log.Panicf("%v out of range for int32", n)
	}
	return int32(n)
}

func rdint16(f *bufio.Reader) int16 {
	n := rdint64(f)
	if int64(int16(n)) != n {
		log.Panicf("%v out of range for int16", n)
	}
	return int16(n)
}

func rduint8(f *bufio.Reader) uint8 {
	n := rdint64(f)
	if int64(uint8(n)) != n {
		log.Panicf("%v out of range for uint8", n)
	}
	return uint8(n)
}

// rdBuf is used by rdstring and rdsymName as scratch for reading strings.
var rdBuf []byte
var emptyPkg = []byte(`"".`)

func rdstring(f *bufio.Reader) string {
	n := rdint(f)
	if len(rdBuf) < n {
		rdBuf = make([]byte, n)
	}
	io.ReadFull(f, rdBuf[:n])
	return string(rdBuf[:n])
}

func rddata(f *bufio.Reader, buf *[]byte) []byte {
	n := rdint(f)
	p := (*buf)[:n:n]
	*buf = (*buf)[n:]
	return p
}

// rdsymName reads a symbol name, replacing all "". with pkg.
func rdsymName(f *bufio.Reader, pkg string) string {
	n := rdint(f)
	if n == 0 {
		rdint64(f)
		return ""
	}

	if len(rdBuf) < n {
		rdBuf = make([]byte, n, 2*n)
	}
	origName := rdBuf[:n]
	io.ReadFull(f, origName)
	adjName := rdBuf[n:n]
	for {
		i := bytes.Index(origName, emptyPkg)
		if i == -1 {
			adjName = append(adjName, origName...)
			break
		}
		adjName = append(adjName, origName[:i]...)
		adjName = append(adjName, pkg...)
		adjName = append(adjName, '.')
		origName = origName[i+len(emptyPkg):]
	}
	name := string(adjName)
	if len(adjName) > len(rdBuf) {
		rdBuf = adjName // save the larger buffer for reuse
	}
	return name
}

func rdsym(ctxt *Link, f *bufio.Reader, pkg string) *LSym {
	i := rdint(f)
	return ctxt.CurRefs[i]
}
