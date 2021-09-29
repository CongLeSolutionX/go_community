// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package encodemeta

import (
	"bufio"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"internal/coverage"
	"internal/coverage/stringtab"
	"os"
	"unsafe"
)

// This package contains APIs and helpers for writing out a meta-data
// file (composed of a file header, offsets/lengths, and then a series of
// meta-data blobs emitted by the compiler, one per Go package).

type CoverageMetaFileWriter struct {
	stringtab.StringTableWriter
	f      *os.File
	mfname string
	w      *bufio.Writer
	tmp    []byte
	debug  bool
}

func NewCoverageMetaFileWriter(mfname string, f *os.File) *CoverageMetaFileWriter {
	r := &CoverageMetaFileWriter{
		f:      f,
		mfname: mfname,
		w:      bufio.NewWriter(f),
		tmp:    make([]byte, 64),
	}
	r.InitStringTableWriter(mfname)
	r.LookupString("")
	return r
}

func (m *CoverageMetaFileWriter) Write(finalHash [16]byte, modname string, blobs [][]byte, mode coverage.CounterMode) error {
	mhsz := uint64(unsafe.Sizeof(coverage.MetaFileHeader{}))
	modnameStrIdx := m.LookupString(modname)
	stSize := m.StringTableSize()
	stOffset := mhsz + uint64(16*len(blobs))
	preambleLength := stOffset + uint64(stSize)

	if m.debug {
		fmt.Fprintf(os.Stderr, "=+= sizeof(MetaFileHeader)=%d\n", mhsz)
		fmt.Fprintf(os.Stderr, "=+= preambleLength=%d\n", preambleLength)
	}

	// Compute total size
	tlen := preambleLength
	for i := 0; i < len(blobs); i++ {
		tlen += uint64(len(blobs[i]))
	}

	// Emit header
	mh := coverage.MetaFileHeader{
		Magic:        coverage.CovMetaMagic,
		TotalLength:  tlen,
		Entries:      uint64(len(blobs)),
		MetaHash:     finalHash,
		StrTabOffset: uint32(stOffset),
		StrTabLength: stSize,
		StrTabEnts:   m.StringTableNentries(),
		ModuleName:   modnameStrIdx,
		CMode:        mode,
	}
	var err error
	if err = binary.Write(m.w, binary.LittleEndian, mh); err != nil {
		return fmt.Errorf("error writing %s: %v\n", m.mfname, err)
	}

	if m.debug {
		fmt.Fprintf(os.Stderr, "=+= len(blobs) is %d\n", mh.Entries)
	}

	// Emit package offsets section followed by package lengths section.
	off := preambleLength
	off2 := mhsz
	buf := make([]byte, 8)
	for _, blob := range blobs {
		binary.LittleEndian.PutUint64(buf, off)
		if _, err = m.w.Write(buf); err != nil {
			return fmt.Errorf("error writing %s: %v\n", m.mfname, err)
		}
		if m.debug {
			fmt.Fprintf(os.Stderr, "=+= pkg offset %d 0x%x\n", off, off)
		}
		off += uint64(len(blob))
		off2 += 8
	}
	for _, blob := range blobs {
		bl := uint64(len(blob))
		binary.LittleEndian.PutUint64(buf, bl)
		if _, err = m.w.Write(buf); err != nil {
			return fmt.Errorf("error writing %s: %v\n", m.mfname, err)
		}
		if m.debug {
			fmt.Fprintf(os.Stderr, "=+= pkg len %d 0x%x\n", bl, bl)
		}
		off2 += 8
	}

	// Emit string table
	if err = m.WriteStringTable(m.w); err != nil {
		return err
	}

	// Now emit blobs themselves.
	for k, blob := range blobs {
		if m.debug {
			fmt.Fprintf(os.Stderr, "=-= writing blob %d len %d at off=%d hash %s\n", k, len(blob), off2, fmt.Sprintf("%x", md5.Sum(blob)))
		}
		if _, err = m.w.Write(blob); err != nil {
			return fmt.Errorf("error writing %s: %v\n", m.mfname, err)
		}
		if m.debug {
			fmt.Fprintf(os.Stderr, "=+= wrote package payload of %d bytes\n",
				len(blob))
		}
		off2 += uint64(len(blob))
	}

	// Flush and close the file, we're done.
	if err = m.w.Flush(); err != nil {
		return fmt.Errorf("error writing %s: %v\n", m.mfname, err)
	}
	if err = m.f.Close(); err != nil {
		return fmt.Errorf("error closing %s: %v\n", m.mfname, err)
	}
	return nil
}
