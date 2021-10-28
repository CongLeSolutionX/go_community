// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package encodemeta

// This package contains APIs and helpers for encoding the meta-data "blob" for
// a single Go package, creat by the compiler when coverage instrumentation
// is turned on.

import (
	"encoding/binary"
	"fmt"
	"internal/coverage"
	"internal/coverage/stringtab"
	"io"
	"os"
)

type CoverageMetaDataBuilder struct {
	stringtab.StringTableWriter
	pkgpath uint32
	funcs   []funcDesc
	tmp     []byte // temp work slice
	debug   bool
}

func NewCoverageMetaDataBuilder(pkgpath string) *CoverageMetaDataBuilder {
	x := &CoverageMetaDataBuilder{
		tmp: make([]byte, 0, 256),
	}
	x.InitStringTableWriter("<byte buffer>")
	x.LookupString("")
	x.pkgpath = x.LookupString(pkgpath)
	return x
}

type funcDesc struct {
	encoded []byte
}

// AddFunc registers a new function with the meta data builder.
func (b *CoverageMetaDataBuilder) AddFunc(f coverage.FuncDesc) uint {
	fd := funcDesc{}
	b.tmp = b.tmp[:0]
	b.tmp = stringtab.AppendUleb128(b.tmp, uint(len(f.Units)))
	b.tmp = stringtab.AppendUleb128(b.tmp, uint(b.LookupString(f.Funcname)))
	b.tmp = stringtab.AppendUleb128(b.tmp, uint(b.LookupString(f.Srcfile)))
	for _, u := range f.Units {
		b.tmp = stringtab.AppendUleb128(b.tmp, uint(u.StLine))
		b.tmp = stringtab.AppendUleb128(b.tmp, uint(u.StCol))
		b.tmp = stringtab.AppendUleb128(b.tmp, uint(u.EnLine))
		b.tmp = stringtab.AppendUleb128(b.tmp, uint(u.EnCol))
		b.tmp = stringtab.AppendUleb128(b.tmp, uint(u.NxStmts))
	}
	fd.encoded = make([]byte, len(b.tmp))
	copy(fd.encoded, b.tmp)
	rv := uint(len(b.funcs))
	b.funcs = append(b.funcs, fd)
	return rv
}

func (b *CoverageMetaDataBuilder) emitFuncOffsets(w io.WriteSeeker, off int64) int64 {
	nFuncs := len(b.funcs)
	var foff int64 = coverage.CovMetaHeaderSize + int64(b.StringTableSize()) + int64(nFuncs)*4
	for idx := 0; idx < nFuncs; idx++ {
		b.wrUint32(w, uint32(foff))
		foff += int64(len(b.funcs[idx].encoded))
	}
	return off + (int64(len(b.funcs)) * 4)
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
	// Write packagepath (a string table index).
	b.wrUint32(w, uint32(b.pkgpath))
	// Write number of files and functions.
	b.wrUint32(w, uint32(b.StringTableNentries()))
	b.wrUint32(w, uint32(len(b.funcs)))

	off := int64(coverage.CovMetaHeaderSize)

	// Write function offsets section
	off = b.emitFuncOffsets(w, off)

	if err := b.WriteStringTable(w); err != nil {
		panic(fmt.Sprintf("internal error writing string table: %v", err))
	}
	off += int64(b.StringTableSize())

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
