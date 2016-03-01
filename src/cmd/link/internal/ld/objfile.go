// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ld

// Writing and reading of Go object files.
//
// Originally, Go object files were Plan 9 object files, but no longer.
// Now they are more like standard object files, in that each symbol is defined
// by an associated memory image (bytes) and a list of relocations to apply
// during linking. We do not (yet?) use a standard file format, however.
// For now, the format is chosen to be as simple as possible to read and write.
// It may change for reasons of efficiency, or we may even switch to a
// standard file format if there are compelling benefits to doing so.
// See golang.org/s/go13linker for more background.
//
// The file format is:
//
//	- magic header: "\x00\x00go13ld"
//	- byte 1 - version number
//	- sequence of strings giving dependencies (imported packages)
//	- empty string (marks end of sequence)
//	- sequence of defined symbols
//	- byte 0xff (marks end of sequence)
//	- magic footer: "\xff\xffgo13ld"
//
// All integers are stored in a zigzag varint format.
// See golang.org/s/go12symtab for a definition.
//
// Data blocks and strings are both stored as an integer
// followed by that many bytes.
//
// A symbol reference is a string name followed by a version.
// An empty name corresponds to a nil LSym* pointer.
//
// Each symbol is laid out as the following fields (taken from LSym*):
//
//	- byte 0xfe (sanity check for synchronization)
//	- type [int]
//	- name [string]
//	- version [int]
//	- flags [int]
//		1 dupok
//	- size [int]
//	- gotype [symbol reference]
//	- p [data block]
//	- nr [int]
//	- r [nr relocations, sorted by off]
//
// If type == STEXT, there are a few more fields:
//
//	- args [int]
//	- locals [int]
//	- nosplit [int]
//	- flags [int]
//		1 leaf
//		2 C function
//	- nlocal [int]
//	- local [nlocal automatics]
//	- pcln [pcln table]
//
// Each relocation has the encoding:
//
//	- off [int]
//	- siz [int]
//	- type [int]
//	- add [int]
//	- xadd [int]
//	- sym [symbol reference]
//	- xsym [symbol reference]
//
// Each local has the encoding:
//
//	- asym [symbol reference]
//	- offset [int]
//	- type [int]
//	- gotype [symbol reference]
//
// The pcln table has the encoding:
//
//	- pcsp [data block]
//	- pcfile [data block]
//	- pcline [data block]
//	- npcdata [int]
//	- pcdata [npcdata data blocks]
//	- nfuncdata [int]
//	- funcdata [nfuncdata symbol references]
//	- funcdatasym [nfuncdata ints]
//	- nfile [int]
//	- file [nfile symbol references]
//
// The file layout and meaning of type integers are architecture-independent.
//
// TODO(rsc): The file format is good for a first pass but needs work.
//	- There are SymID in the object file that should really just be strings.
//	- The actual symbol memory images are interlaced with the symbol
//	  metadata. They should be separated, to reduce the I/O required to
//	  load just the metadata.
//	- The symbol references should be shortened, either with a symbol
//	  table or by using a simple backward index to an earlier mentioned symbol.

import (
	"bytes"
	"cmd/internal/obj"
	"fmt"
	"log"
	"strconv"
	"strings"
)

const (
	startmagic = "\x00\x00go13ld"
	endmagic   = "\xff\xffgo13ld"
)

func ldobjfile(ctxt *Link, f *Input, pkg string, pn string) {
	ctxt.Version++
	if magic, err := f.ReadBytes(8); err != nil || string(magic) != startmagic {
		log.Fatalf("%s: invalid object file start magic %q", pn, string(magic))
	}
	if version := f.ReadByte(); version != 1 {
		log.Fatalf("%s: invalid file version number %d", pn, version)
	}

	for {
		lib := f.ReadString()
		if lib == "" {
			break
		}
		addlib(ctxt, pkg, pn, lib)
	}

	for {
		if f.data[f.off] == 0xff {
			break
		}
		readsym(ctxt, f, pkg, pn)
	}

	if magic, err := f.ReadBytes(8); err != nil || string(magic) != endmagic {
		log.Fatalf("%s: invalid object file end magic %q", pn, string(magic))
	}
}

var readsym_ndup int

func readsym(ctxt *Link, f *Input, pkg string, pn string) {
	if f.ReadByte() != 0xfe {
		log.Fatalf("readsym out of sync")
	}
	t := f.ReadInt()
	name := rdsymName(f, pkg)
	v := f.ReadInt()
	if v != 0 && v != 1 {
		log.Fatalf("invalid symbol version %d", v)
	}
	flags := f.ReadInt()
	dupok := flags & 1
	local := false
	if flags&2 != 0 {
		local = true
	}
	size := f.ReadInt()
	typ := rdsym(ctxt, f, pkg)
	data := f.ReadData()
	nreloc := f.ReadInt()

	if v != 0 {
		v = ctxt.Version
	}
	s := Linklookup(ctxt, name, v)
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
		if s.Type != obj.SBSS && s.Type != obj.SNOPTRBSS && dupok == 0 && s.Dupok == 0 {
			log.Fatalf("duplicate symbol %s (types %d and %d) in %s and %s", s.Name, s.Type, t, s.File, pn)
		}
		if len(s.P) > 0 {
			dup = s
			s = linknewsym(ctxt, ".dup", readsym_ndup)
			readsym_ndup++ // scratch
		}
	}

overwrite:
	s.File = pkg
	s.Dupok = uint8(dupok)
	if t == obj.SXREF {
		log.Fatalf("bad sxref")
	}
	if t == 0 {
		log.Fatalf("missing type for %s in %s", name, pn)
	}
	if t == obj.SBSS && (s.Type == obj.SRODATA || s.Type == obj.SNOPTRBSS) {
		t = int(s.Type)
	}
	s.Type = int16(t)
	if s.Size < int64(size) {
		s.Size = int64(size)
	}
	s.Local = local
	if typ != nil { // if bss sym defined multiple times, take type from any one def
		s.Gotype = typ
	}
	if dup != nil && typ != nil {
		dup.Gotype = typ
	}
	s.P = data
	s.P = s.P[:len(data)]
	if nreloc > 0 {
		s.R = make([]Reloc, nreloc)
		s.R = s.R[:nreloc]
		var r *Reloc
		for i := 0; i < nreloc; i++ {
			r = &s.R[i]
			r.Off = f.ReadInt32()
			r.Siz = f.ReadUint8()
			r.Type = f.ReadInt32()
			r.Add = f.ReadInt64()
			f.ReadInt64() // Xadd, ignored
			r.Sym = rdsym(ctxt, f, pkg)
			rdsym(ctxt, f, pkg) // Xsym, ignored
		}
	}

	if len(s.P) > 0 && dup != nil && len(dup.P) > 0 && strings.HasPrefix(s.Name, "gclocals·") {
		// content-addressed garbage collection liveness bitmap symbol.
		// double check for hash collisions.
		if !bytes.Equal(s.P, dup.P) {
			log.Fatalf("dupok hash collision for %s in %s and %s", s.Name, s.File, pn)
		}
	}

	if s.Type == obj.STEXT {
		s.Args = f.ReadInt32()
		s.Locals = f.ReadInt32()
		s.Nosplit = f.ReadUint8()
		v := f.ReadInt()
		s.Leaf = uint8(v & 1)
		s.Cfunc = uint8(v & 2)
		n := f.ReadInt()
		var a *Auto
		for i := 0; i < n; i++ {
			a = new(Auto)
			a.Asym = rdsym(ctxt, f, pkg)
			a.Aoffset = f.ReadInt32()
			a.Name = rdint16(f)
			a.Gotype = rdsym(ctxt, f, pkg)
			a.Link = s.Autom
			s.Autom = a
		}

		s.Pcln = new(Pcln)
		pc := s.Pcln
		pc.Pcsp.P = f.ReadData()
		pc.Pcfile.P = f.ReadData()
		pc.Pcline.P = f.ReadData()
		n = f.ReadInt()
		pc.Pcdata = make([]Pcdata, n)
		pc.Npcdata = n
		for i := 0; i < n; i++ {
			pc.Pcdata[i].P = f.ReadData()
		}
		n = f.ReadInt()
		pc.Funcdata = make([]*LSym, n)
		pc.Funcdataoff = make([]int64, n)
		pc.Nfuncdata = n
		for i := 0; i < n; i++ {
			pc.Funcdata[i] = rdsym(ctxt, f, pkg)
		}
		for i := 0; i < n; i++ {
			pc.Funcdataoff[i] = f.ReadInt64()
		}
		n = f.ReadInt()
		pc.File = make([]*LSym, n)
		pc.Nfile = n
		for i := 0; i < n; i++ {
			pc.File[i] = rdsym(ctxt, f, pkg)
		}

		if dup == nil {
			if s.Onlist != 0 {
				log.Fatalf("symbol %s listed multiple times", s.Name)
			}
			s.Onlist = 1
			if ctxt.Etextp != nil {
				ctxt.Etextp.Next = s
			} else {
				ctxt.Textp = s
			}
			ctxt.Etextp = s
		}
	}

	if ctxt.Debugasm != 0 {
		fmt.Fprintf(ctxt.Bso, "%s ", s.Name)
		if s.Version != 0 {
			fmt.Fprintf(ctxt.Bso, "v=%d ", s.Version)
		}
		if s.Type != 0 {
			fmt.Fprintf(ctxt.Bso, "t=%d ", s.Type)
		}
		if s.Dupok != 0 {
			fmt.Fprintf(ctxt.Bso, "dupok ")
		}
		if s.Cfunc != 0 {
			fmt.Fprintf(ctxt.Bso, "cfunc ")
		}
		if s.Nosplit != 0 {
			fmt.Fprintf(ctxt.Bso, "nosplit ")
		}
		fmt.Fprintf(ctxt.Bso, "size=%d value=%d", int64(s.Size), int64(s.Value))
		if s.Type == obj.STEXT {
			fmt.Fprintf(ctxt.Bso, " args=%#x locals=%#x", uint64(s.Args), uint64(s.Locals))
		}
		fmt.Fprintf(ctxt.Bso, "\n")
		var c int
		var j int
		for i := 0; i < len(s.P); {
			fmt.Fprintf(ctxt.Bso, "\t%#04x", uint(i))
			for j = i; j < i+16 && j < len(s.P); j++ {
				fmt.Fprintf(ctxt.Bso, " %02x", s.P[j])
			}
			for ; j < i+16; j++ {
				fmt.Fprintf(ctxt.Bso, "   ")
			}
			fmt.Fprintf(ctxt.Bso, "  ")
			for j = i; j < i+16 && j < len(s.P); j++ {
				c = int(s.P[j])
				if ' ' <= c && c <= 0x7e {
					fmt.Fprintf(ctxt.Bso, "%c", c)
				} else {
					fmt.Fprintf(ctxt.Bso, ".")
				}
			}

			fmt.Fprintf(ctxt.Bso, "\n")
			i += 16
		}

		var r *Reloc
		for i := 0; i < len(s.R); i++ {
			r = &s.R[i]
			fmt.Fprintf(ctxt.Bso, "\trel %d+%d t=%d %s+%d\n", int(r.Off), r.Siz, r.Type, r.Sym.Name, int64(r.Add))
		}
	}
}

func rdint16(f *Input) int16 {
	n := f.ReadInt64()
	if int64(int16(n)) != n {
		log.Panicf("%v out of range for int16", n)
	}
	return int16(n)
}

// rdBuf is used by rdsymName as scratch for building strings.
var rdBuf []byte
var emptyPkg = []byte(`"".`)

// rdsymName reads a symbol name, replacing all "". with pkg.
func rdsymName(f *Input, pkg string) string {
	n := f.ReadInt()
	if n == 0 {
		f.ReadInt64()
		return ""
	}

	origName := f.data[f.off : f.off+n]
	f.off += n

	if len(rdBuf) < n {
		rdBuf = make([]byte, n, 2*n)
	}
	rdBuf = rdBuf[:0]
	for {
		i := bytes.Index(origName, emptyPkg)
		if i == -1 {
			rdBuf = append(rdBuf, origName...)
			break
		}
		rdBuf = append(rdBuf, origName[:i]...)
		rdBuf = append(rdBuf, pkg...)
		rdBuf = append(rdBuf, '.')
		origName = origName[i+len(emptyPkg):]
	}
	name := string(rdBuf)
	rdBuf = rdBuf[:0]
	return name
}

func rdsym(ctxt *Link, f *Input, pkg string) *LSym {
	name := rdsymName(f, pkg)
	if name == "" {
		return nil
	}
	v := f.ReadInt()
	if v != 0 {
		v = ctxt.Version
	}
	s := Linklookup(ctxt, name, v)

	if v == 0 && s.Name[0] == '$' && s.Type == 0 {
		if strings.HasPrefix(s.Name, "$f32.") {
			x, _ := strconv.ParseUint(s.Name[5:], 16, 32)
			i32 := int32(x)
			s.Type = obj.SRODATA
			s.Local = true
			Adduint32(ctxt, s, uint32(i32))
			s.Reachable = false
		} else if strings.HasPrefix(s.Name, "$f64.") || strings.HasPrefix(s.Name, "$i64.") {
			x, _ := strconv.ParseUint(s.Name[5:], 16, 64)
			i64 := int64(x)
			s.Type = obj.SRODATA
			s.Local = true
			Adduint64(ctxt, s, uint64(i64))
			s.Reachable = false
		}
	}
	if v == 0 && strings.HasPrefix(s.Name, "runtime.gcbits.") {
		s.Local = true
	}
	return s
}
