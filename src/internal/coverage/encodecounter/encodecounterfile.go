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
	"unsafe"
)

// This package contains APIs and helpers for encoding initial portions
// of the counter data files emitted at runtime when coverage instrumention
// is enabled.

type CoverageDataFileWriter struct {
	stringtab.StringTableWriter
	f      *os.File
	cfname string
	w      *bufio.Writer
	tmp    []byte
	debug  bool
}

func NewCoverageDataFileWriter(cfname string, f *os.File) *CoverageDataFileWriter {
	r := &CoverageDataFileWriter{
		f:      f,
		cfname: cfname,
		w:      bufio.NewWriter(f),
		tmp:    make([]byte, 64),
	}
	r.InitStringTableWriter(cfname)
	r.LookupString("")
	return r
}

type CounterVisitorFcn func(pkid uint32, funcid uint32, counters []uint32) bool
type CounterVisitor func(f CounterVisitorFcn) bool

func (cfw *CoverageDataFileWriter) Write(finalHash [16]byte, visitor CounterVisitor) bool {

	// Notes:
	// - this version writes everything little-endian, which means
	//   a call is needed encode every value (expensive)
	// - we may want to move to a model in which we just blast out
	//   all counters, or possibly mmap the file and do the write
	//   implicitly.
	u32sz := uint64(unsafe.Sizeof(uint32(1)))
	etot := uint64(0)
	tot := uint64(unsafe.Sizeof(coverage.CounterFileHeader{}))
	sizer := func(pkid uint32, funcid uint32, counters []uint32) bool {
		etot++
		tot += uint64((3 + len(counters)) * int(u32sz))
		return true
	}

	// Compute output file size info.
	visitor(sizer)

	w := bufio.NewWriter(cfw.f)

	// Emit header. At the moment we're emitting everything
	// little-endian, but we want to leave the door open for the
	// possibility of emitting using native endianity and then having
	// the reader adjust accordingly.
	ch := coverage.CounterFileHeader{
		Magic:       coverage.CovCounterMagic,
		TotalLength: tot,
		Entries:     etot,
		MetaHash:    finalHash,
		BigEndian:   false,
	}
	if cfw.debug {
		fmt.Fprintf(os.Stderr, "=+= ch: %+v\n", ch)
	}
	var err error
	if err = binary.Write(w, binary.LittleEndian, ch); err != nil {
		fmt.Fprintf(os.Stderr, "error writing %s: %v\n", cfw.cfname, err)
		return false
	}

	ctrb := make([]byte, u32sz)
	wru32 := func(val uint32) bool {
		binary.LittleEndian.PutUint32(ctrb, val)
		var sz int
		if sz, err = w.Write(ctrb); err != nil || sz != int(u32sz) {
			fmt.Fprintf(os.Stderr, "error writing %s: %v\n", cfw.cfname, err)
			return false
		}
		return true
	}

	// Write out entries for each live function.
	emitter := func(pkid uint32, funcid uint32, counters []uint32) bool {
		ok := wru32(uint32(len(counters))) && wru32(pkid) && wru32(funcid)
		if !ok {
			return false
		}
		for _, val := range counters {
			if !wru32(val) {
				return false
			}
		}
		return true
	}
	if !visitor(emitter) {
		return false
	}

	// Flush.
	if err = w.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "error writing %s: %v\n", cfw.cfname, err)
		return false
	}
	return true

}
