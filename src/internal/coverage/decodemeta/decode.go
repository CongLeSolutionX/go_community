// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package decodemeta

// This package contains APIs and helpers for decoding a single package's
// meta data "blob" emitted by the compiler when covereage instrumentation
// is turned on.

import (
	"fmt"
	"internal/coverage"
	"internal/coverage/slicereader"
	"internal/coverage/stringtab"
	"os"
)

// See comments in the encodecovmeta package for details on the format.

type CoverageMetaDataDecoder struct {
	r       *slicereader.Reader
	size    uint32
	pkgpath uint32
	nfiles  uint32
	nfuncs  uint32
	strtab  *stringtab.StringTableReader
	tmp     []byte
	debug   bool
}

func NewCoverageMetaDataDecoder(b []byte, readonly bool) *CoverageMetaDataDecoder {
	slr := slicereader.NewReader(b, readonly)
	x := &CoverageMetaDataDecoder{
		r:    slr,
		size: uint32(len(b)),
		tmp:  make([]byte, 0, 256),
	}
	x.readHeader()
	x.readStringTable()
	return x
}

func (d *CoverageMetaDataDecoder) readHeader() {
	d.size = d.r.ReadUint32()
	d.pkgpath = d.r.ReadUint32()
	d.nfiles = d.r.ReadUint32()
	d.nfuncs = d.r.ReadUint32()
	if d.debug {
		fmt.Fprintf(os.Stderr, "=-= after readHeader: %+v\n", d)
	}
}

func (d *CoverageMetaDataDecoder) readStringTable() {
	// Seek to the correct location to read the string table.
	stringTableLocation := int64(coverage.CovMetaHeaderSize + 4*d.nfuncs)
	d.r.SeekTo(stringTableLocation)

	// Read the table itself.
	d.strtab = stringtab.NewReader(d.r)
	d.strtab.Read(int(d.nfiles))
}

func (d *CoverageMetaDataDecoder) PackagePath() string {
	return d.strtab.Get(d.pkgpath)
}

func (d *CoverageMetaDataDecoder) NumFuncs() uint32 {
	return d.nfuncs
}

// ReadFunc reads the coverage meta-data for the function with index
// 'findex', filling it into the FuncDesc pointed to by 'f'.
func (d *CoverageMetaDataDecoder) ReadFunc(fidx uint32, f *coverage.FuncDesc) error {
	if fidx >= d.nfuncs {
		return fmt.Errorf("illegal function index")
	}

	// Seek to the correct location to read the function offset and read it.
	funcOffsetLocation := int64(coverage.CovMetaHeaderSize + 4*fidx)
	d.r.SeekTo(funcOffsetLocation)
	foff := d.r.ReadUint32()

	// Sanity check
	if foff < uint32(funcOffsetLocation) || foff > d.size {
		return fmt.Errorf("malformed func offset %d", foff)
	}

	// Seek to the correct location to read the function.
	d.r.SeekTo(int64(foff))

	// Preamble containing number of units, file, and function.
	numUnits := uint32(d.r.ReadULEB128())
	fnameidx := uint32(d.r.ReadULEB128())
	fileidx := uint32(d.r.ReadULEB128())

	f.Srcfile = d.strtab.Get(fileidx)
	f.Funcname = d.strtab.Get(fnameidx)

	// Now the units
	f.Units = f.Units[:0]
	if cap(f.Units) < int(numUnits) {
		f.Units = make([]coverage.CoverableUnit, 0, numUnits)
	}
	for k := uint32(0); k < numUnits; k++ {
		f.Units = append(f.Units,
			coverage.CoverableUnit{
				StLine:  uint32(d.r.ReadULEB128()),
				StCol:   uint32(d.r.ReadULEB128()),
				EnLine:  uint32(d.r.ReadULEB128()),
				EnCol:   uint32(d.r.ReadULEB128()),
				NxStmts: uint32(d.r.ReadULEB128()),
			})
	}
	return nil
}
