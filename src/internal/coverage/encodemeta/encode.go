// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package encodemeta

// This package contains APIs and helpers for encoding the meta-data "blob" for
// a single Go package, creat by the compiler when coverage instrumentation
// is turned on.

import (
	"encoding/binary"
	"internal/coverage"
	"io"
	"os"
)

const metaFormat = `
--header----------
  | size: size of this blob in bytes
  | packagepath: <path to p>
  | module: <modulename>
  | classification: ...
  | nfiles: 1
  | nfunctions: 2
  --func offsets table------
  <offset to func 0>
  <offset to func 1>
  --file + function table------
  | <uleb128 len> 4
  | <data> "p.go"
  | <uleb128 len> 5
  | <data> "small"
  | <uleb128 len> 6
  | <data> "Medium"
  --func 1------
  | <uleb128> num units: 3
  | <uleb128> func name: 1 (index into string table)
  | <uleb128> file: 0 (index into string table)
  | <unit 0>:  F0   L6     L8    2
  | <unit 1>:  F0   L9     L9    1
  | <unit 2>:  F0   L11    L11   1
  --func 2------
  | <uleb128> num units: 1
  | <uleb128> func name: 2 (index into string table)
  | <uleb128> file: 0 (index into string table)
  | <unit 0>:  F0   L15    L19   5
  ---end-----------
`

type CoverageMetaDataBuilder struct {
	pkgpath ffTabIdx
	modname ffTabIdx
	fftab   map[string]ffTabIdx // file+function table
	funcs   []funcDesc
	tmp     []byte // temp work slice
}

func NewCoverageMetaDataBuilder(pkgpath string, modname string) *CoverageMetaDataBuilder {
	x := &CoverageMetaDataBuilder{
		fftab: make(map[string]ffTabIdx),
		tmp:   make([]byte, 0, 256),
	}
	x.pkgpath = x.slookup(pkgpath)
	x.modname = x.slookup(modname)
	return x
}

func (b *CoverageMetaDataBuilder) slookup(s string) ffTabIdx {
	if v, ok := b.fftab[s]; ok {
		return v
	}
	v := ffTabIdx(len(b.fftab))
	b.fftab[s] = v
	return v
}

type ffTabIdx uint32

type funcDesc struct {
	encoded []byte
}

func appendUleb128(b []byte, v uint) []byte {
	for {
		c := uint8(v & 0x7f)
		v >>= 7
		if v != 0 {
			c |= 0x80
		}
		b = append(b, c)
		if c&0x80 == 0 {
			break
		}
	}
	return b
}

// AddFunc registers a new function with the meta data builder.
func (b *CoverageMetaDataBuilder) AddFunc(f coverage.FuncDesc) uint {
	fd := funcDesc{}
	b.tmp = b.tmp[:0]
	b.tmp = appendUleb128(b.tmp, uint(len(f.Units)))
	b.tmp = appendUleb128(b.tmp, uint(b.slookup(f.Funcname)))
	b.tmp = appendUleb128(b.tmp, uint(b.slookup(f.Srcfile)))
	for _, u := range f.Units {
		b.tmp = appendUleb128(b.tmp, uint(u.StLine))
		b.tmp = appendUleb128(b.tmp, uint(u.StCol))
		b.tmp = appendUleb128(b.tmp, uint(u.EnLine))
		b.tmp = appendUleb128(b.tmp, uint(u.EnCol))
		b.tmp = appendUleb128(b.tmp, uint(u.NxStmts))
	}
	fd.encoded = make([]byte, len(b.tmp))
	copy(fd.encoded, b.tmp)
	rv := uint(len(b.funcs))
	b.funcs = append(b.funcs, fd)
	return rv
}

func (b *CoverageMetaDataBuilder) emitFuncOffsets(w io.WriteSeeker, off int64) int64 {
	nFuncs := len(b.funcs)
	var foff int64 = coverage.CovMetaHeaderSize + b.stringTableSize() + int64(nFuncs)*4
	for idx := 0; idx < nFuncs; idx++ {
		b.wrUint32(w, uint32(foff))
		foff += int64(len(b.funcs[idx].encoded))
	}
	return off + (int64(len(b.funcs)) * 4)
}

func (b *CoverageMetaDataBuilder) stringTableSize() int64 {
	sz := int64(0)
	for s := range b.fftab {
		b.tmp = b.tmp[:0]
		b.tmp = appendUleb128(b.tmp, uint(len(s)))
		sz += int64(len(b.tmp) + len(s))
	}
	return sz
}

// emitStrTable writes the meta-data string table to the specified
// symbol, returning a new offset along with an error value.
func (b *CoverageMetaDataBuilder) emitStrTable(w io.WriteSeeker, off int64) int64 {
	// Write contents of the string table. This is a series of pairs
	// [L,S] where L is a uleb128 encoded length and S is the bytes
	// for the string.
	strlist := make([]string, len(b.fftab))
	for k, v := range b.fftab {
		strlist[v] = k
	}
	b.tmp = b.tmp[:0]
	for _, s := range strlist {
		b.tmp = appendUleb128(b.tmp, uint(len(s)))
		b.tmp = append(b.tmp, []byte(s)...)
	}

	// Write the blob.
	ew := len(b.tmp)
	if nw, err := w.Write(b.tmp); err != nil || ew != nw {
		panic("unexpected write failure")
	}
	return off + int64(ew)
}

func (b *CoverageMetaDataBuilder) emitFunc(w io.WriteSeeker, off int64, f funcDesc) int64 {
	ew := len(f.encoded)
	if nw, err := w.Write(f.encoded); err != nil || ew != nw {
		panic("unexpected write failure")
	}
	return off + int64(ew)
}

func (b *CoverageMetaDataBuilder) wrUint32(w io.WriteSeeker, v uint32) {
	b.tmp = b.tmp[:0]
	b.tmp = append(b.tmp, []byte{0, 0, 0, 0}...)
	binary.LittleEndian.PutUint32(b.tmp, v)
	if nw, err := w.Write(b.tmp); nw != 4 || err != nil {
		panic("unexpected write failure")
	}
}

// Emit writes the meta-data accumulated so far in this builder to 'w'.
// We assume that 'w' is backed by an underlying byte slice, and that
// writes / seeks to 'w' will never fail (this code will panic if so).
func (b *CoverageMetaDataBuilder) Emit(w io.WriteSeeker) {
	// Placeholder for length, to be updated later.
	b.wrUint32(w, 0)
	// Write packagepath and module name (both string table indices)
	b.wrUint32(w, uint32(b.pkgpath))
	b.wrUint32(w, uint32(b.modname))
	// Write number of files and functions.
	b.wrUint32(w, uint32(len(b.fftab)))
	b.wrUint32(w, uint32(len(b.funcs)))

	off := int64(coverage.CovMetaHeaderSize)

	// Write function offsets section
	off = b.emitFuncOffsets(w, off)

	// Write string table
	off = b.emitStrTable(w, off)

	// Write functions
	for _, f := range b.funcs {
		off = b.emitFunc(w, off, f)
	}

	// Back-patch the length.
	totalLength := uint32(off)
	if _, err := w.Seek(0, os.SEEK_SET); err != nil {
		panic("unexpected seek failure")
	}
	b.wrUint32(w, totalLength)
}
