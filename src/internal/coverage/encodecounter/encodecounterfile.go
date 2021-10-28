// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package encodecounter

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"internal/coverage"
	"internal/coverage/stringtab"
	"os"
)

// This package contains APIs and helpers for encoding initial portions
// of the counter data files emitted at runtime when coverage instrumention
// is enabled.

type CoverageDataFileWriter struct {
	stringtab.StringTableWriter
	cfname  string
	f       *os.File
	w       *bufio.Writer
	ch      coverage.CounterFileHeader
	args    []string
	tmp     []byte
	cflavor coverage.CounterFlavor
	debug   bool
}

func NewCoverageDataFileWriter(cfname string, f *os.File, args []string, flav coverage.CounterFlavor) *CoverageDataFileWriter {
	r := &CoverageDataFileWriter{
		f:       f,
		cfname:  cfname,
		w:       bufio.NewWriter(f),
		args:    args,
		tmp:     make([]byte, 64),
		cflavor: flav,
	}
	r.InitStringTableWriter(cfname)
	r.LookupString("")
	if r.debug {
		fmt.Fprintf(os.Stderr, "=-= new covcounters writer, args: %+v\n", args)
	}
	for _, a := range args {
		r.LookupString(a)
	}
	return r
}

type CounterVisitorFcn func(pkid uint32, funcid uint32, counters []uint32) bool
type CounterVisitor func(f CounterVisitorFcn) bool

// Write writes the contents of the count-data file to the file
// previously supplied to NewCoverageDataFileWriter. The file is not
// closed following the write. If an error takes place, it will be
// logged to standard error, and the return value will be FALSE. If
// the write is successful, TRUE will be returned.
func (cfw *CoverageDataFileWriter) Write(metaFileHash [16]byte, visitor CounterVisitor) bool {
	return cfw.writeHeader(metaFileHash) &&
		cfw.writeStringTable() &&
		cfw.writeArgs() &&
		cfw.writeCounters(visitor) &&
		cfw.backPatchHeader()
}

func (cfw *CoverageDataFileWriter) writeHeader(metaFileHash [16]byte) bool {
	// Emit proto-header, mainly for magic and final hash. For most
	// of the header fields we'll write the real values later on.
	cfw.ch = coverage.CounterFileHeader{
		Magic:     coverage.CovCounterMagic,
		MetaHash:  metaFileHash,
		BigEndian: false,
		CFlavor:   cfw.cflavor,
	}
	if err := binary.Write(cfw.w, binary.LittleEndian, cfw.ch); err != nil {
		fmt.Fprintf(os.Stderr, "error writing %s: %v\n", cfw.cfname, err)
		return false
	}
	if cfw.debug {
		fmt.Fprintf(os.Stderr, "=-= wrote header\n")
	}
	return true
}

func (cfw *CoverageDataFileWriter) writeBytes(b []byte) bool {
	if len(b) == 0 {
		return true
	}
	nw, err := cfw.w.Write(b)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error writing %s: %v\n", cfw.cfname, err)
		return false
	}
	if len(b) != nw {
		fmt.Fprintf(os.Stderr, "error writing %s: short write\n", cfw.cfname)
		return false
	}
	return true
}

// curOffset returns the current offset in the file being written; it
// flushes the bufio.Writer we're working with and then gets the offset
// from the underlying file.
func (cfw *CoverageDataFileWriter) curOffset() (int64, bool) {
	if err := cfw.w.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "error writing %s: %v\n", cfw.cfname, err)
		return 0, false
	}
	off, err := cfw.f.Seek(0, os.SEEK_CUR)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error seeking %s: %v\n", cfw.cfname, err)
		return 0, false
	}
	return off, true
}

func (cfw *CoverageDataFileWriter) padToFourByteBoundary() bool {
	off, ok := cfw.curOffset()
	if !ok {
		return false
	}
	zeros := []byte{0, 0, 0, 0}
	rem := uint32(off) % 4
	if rem != 0 {
		pad := zeros[:(4 - rem)]
		if cfw.debug {
			fmt.Fprintf(os.Stderr, "=-= writing pad %+v\n", pad)
		}
		if !cfw.writeBytes(pad) {
			return false
		}
	}
	if cfw.debug {
		fmt.Fprintf(os.Stderr, "=-= padToFourByteBoundary done\n")
	}
	return true
}

func (cfw *CoverageDataFileWriter) writeStringTable() bool {
	off, ok := cfw.curOffset()
	if !ok {
		return false
	}
	cfw.ch.StrTabOff = uint32(off)
	cfw.ch.StrTabNentries = cfw.StringTableNentries()
	if err := cfw.WriteStringTable(cfw.w); err != nil {
		fmt.Fprintf(os.Stderr, "error writing %s: %v\n", cfw.cfname, err)
		return false
	}
	if !cfw.padToFourByteBoundary() {
		return false
	}
	off, ok = cfw.curOffset()
	if !ok {
		return false
	}
	cfw.ch.StrTabLen = uint32(off) - cfw.ch.StrTabOff
	if cfw.debug {
		fmt.Fprintf(os.Stderr, "=-= wrote string table\n")
	}
	return true
}

func (cfw *CoverageDataFileWriter) writeArgs() bool {
	// Args section looks like:
	// - series of string table refs
	// - optional padding
	off, ok := cfw.curOffset()
	if !ok {
		return false
	}
	cfw.ch.ArgsOff = uint32(off)
	cfw.ch.ArgsNentries = uint32(len(cfw.args))
	for _, arg := range cfw.args {
		cfw.tmp = cfw.tmp[:0]
		cfw.tmp = stringtab.AppendUleb128(cfw.tmp, uint(cfw.LookupString(arg)))
		if !cfw.writeBytes(cfw.tmp) {
			return false
		}
	}
	if cfw.cflavor == coverage.CtrRaw {
		if !cfw.padToFourByteBoundary() {
			return false
		}
	}
	off, ok = cfw.curOffset()
	if !ok {
		return false
	}
	cfw.ch.ArgsLen = uint32(off) - cfw.ch.ArgsOff
	return true
}

func (cfw *CoverageDataFileWriter) writeCounters(visitor CounterVisitor) bool {
	// Notes:
	// - this version writes everything little-endian, which means
	//   a call is needed encode every value (expensive)
	// - we may want to move to a model in which we just blast out
	//   all counters, or possibly mmap the file and do the write
	//   implicitly.
	ctrb := make([]byte, 4)
	wrval := func(val uint32) bool {
		var buf []byte
		var towr int
		if cfw.cflavor == coverage.CtrRaw {
			binary.LittleEndian.PutUint32(ctrb, val)
			buf = ctrb
			towr = 4
		} else if cfw.cflavor == coverage.CtrULeb128 {
			cfw.tmp = cfw.tmp[:0]
			cfw.tmp = stringtab.AppendUleb128(cfw.tmp, uint(val))
			buf = cfw.tmp
			towr = len(buf)
		} else {
			panic("internal error: bad counter flavor")
		}
		sz, err := cfw.w.Write(buf)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error writing %s: %v\n", cfw.cfname, err)
			return false
		}
		if sz != towr {
			fmt.Fprintf(os.Stderr, "error writing %s: short write\n", cfw.cfname)
			return false
		}
		return true
	}

	// Write out entries for each live function.
	nentries := uint64(0)
	emitter := func(pkid uint32, funcid uint32, counters []uint32) bool {
		if !wrval(uint32(len(counters))) || !wrval(pkid) || !wrval(funcid) {
			return false
		}
		for _, val := range counters {
			if !wrval(val) {
				return false
			}
		}
		nentries++
		return true
	}
	if !visitor(emitter) {
		return false
	}
	cfw.ch.FcnEntries = nentries
	return true
}

func (cfw *CoverageDataFileWriter) backPatchHeader() bool {
	// Collect final offset (length of file), update header.
	off, ok := cfw.curOffset()
	if !ok {
		return false
	}
	cfw.ch.TotalLength = uint64(off)

	if cfw.debug {
		fmt.Fprintf(os.Stderr, "=-= back-patched header for %s: %+v\n",
			cfw.cfname, cfw.ch)
	}

	// Seek to the start of the file.
	off, err := cfw.f.Seek(0, os.SEEK_SET)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error seeking %s: %v\n", cfw.cfname, err)
		return false
	}

	cfw.w = bufio.NewWriter(cfw.f)
	if err := binary.Write(cfw.w, binary.LittleEndian, cfw.ch); err != nil {
		fmt.Fprintf(os.Stderr, "error writing %s: %v\n", cfw.cfname, err)
		return false
	}
	if err := cfw.w.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "error writing %s: %v\n", cfw.cfname, err)
		return false
	}
	return true
}
