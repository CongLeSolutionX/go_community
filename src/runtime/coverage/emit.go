// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package coverage

import (
	"bufio"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"internal/coverage"
	"internal/coverage/encodemeta"
	"os"
	"path/filepath"
	"reflect"
	"time"
	"unsafe"
)

// This file contains functions that support the writing of data files
// emitted at the end of code coverage testing runs, from instrumented
// executables.

// covmetablob is a container for holding the meta-data symbol (an
// RODATA variable) for an instrumented Go package. Here "P" points to
// the symbol itself, "Len" is the length of the sym in bytes, and
// "Hash" is an md5sum for the sym computed by the compiler.
type covmetablob struct {
	// Important: any changes to this struct should also be made
	// in the file runtime/covermeta.go.
	P    *byte
	Len  uint32
	Hash [16]byte
}

//go:linkname runtime_getcovmetalist runtime.getcovmetalist
func runtime_getcovmetalist() []covmetablob

// covcounterblob is a container for encapsulating a counter section
// (BSS variable) for an instrumented Go module. Here "Counters" points to
// the counter payload and Len is the number of uint32 entries in the
// section.
type covcounterblob struct {
	// Important: any changes to this struct should also be made in
	// the runtime/covermeta.go, since that code expects a
	// specific format here.
	Counters *uint32
	Len      uint64
}

//go:linkname runtime_getcovcounterlist runtime.getcovcounterlist
func runtime_getcovcounterlist() []covcounterblob

// emitState holds useful state information during the emit process.
type emitState struct {
	mfname string   // path of final meta-data output file
	mftmp  string   // path to meta-data temp file (if needed)
	mf     *os.File // open os.File for meta-data temp file
	cfname string   // path of final counter data file
	cf     *os.File // open os.File for counter data file

	// output directory
	outdir string

	// List of meta-data symbols obtained from the runtime
	metalist []covmetablob

	// List of counter-data symbols obtained from the runtime
	counterlist []covcounterblob

	// emit debug trace output
	debug bool
}

// FIXME: thread through build ID here as well, not sure how.
func emitData() {

	// Ask the runtime for the list of coverage meta-data symbols.
	ml := runtime_getcovmetalist()
	cl := runtime_getcovcounterlist()

	if len(ml) == 0 || len(cl) == 0 {
		return
	}

	s := &emitState{
		metalist:    ml,
		counterlist: cl,
		debug:       os.Getenv("GOCOVERDEBUG") != "",
	}

	h := md5.New()
	tlen := uint64(unsafe.Sizeof(coverage.MetaFileHeader{}))
	for _, entry := range ml {
		if _, err := h.Write(entry.Hash[:]); err != nil {
			// Is this the right away report this error?
			panic(fmt.Sprintf("internal error: md5 sum failed: %v", err))
		}
		tlen += uint64(entry.Len)
	}
	finalHash := md5.Sum(nil)

	// Open output files.
	s.openOutputFiles(finalHash, tlen)
	if s.cf == nil {
		// something went wrong, bail here.
		return
	}

	// Emit meta-data file if needed.
	if s.mf != nil {
		if !s.emitMetaDataFile(tlen, finalHash) {
			return
		}
	}

	// Emit counter data file.
	s.emitCounterDataFile(finalHash)
}

// openMeta determines whether we need to emit a meta-data output
// file, or whether we can reuse the existing file in the coverage out
// dir. It updates mfname/mftmp/mf fields in 'od', returning false on
// error and true for success.
func (s *emitState) openMeta(metaHash [16]byte, metaLen uint64) bool {

	// FIXME: fill in real value here.
	buildid := 0

	// Open meta-outfile for reading to see if it exists.
	fn := fmt.Sprintf("%s.%x.%x", coverage.MetaFilePref, metaHash, buildid)
	s.mfname = filepath.Join(s.outdir, fn)
	fi, err := os.Stat(s.mfname)
	if err != nil || fi.Size() != int64(metaLen) {
		// We need a new meta-file.
		s.mftmp = s.mfname + fmt.Sprintf("%d", time.Now().UnixNano())
		s.mf, err = os.OpenFile(s.mftmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error opening %s: %v\n", s.mftmp, err)
			return false
		}
	}
	return true
}

// openCounter opens an output file for the counter data portion of a
// test coverage run. If updates the 'cfname' and 'cf' fields in the
// passed in state struct, returning false on error (if there was an
// error opening the file) or true for success.
func (s *emitState) openCounter(metaHash [16]byte) bool {
	// FIXME: fill in real value here.
	buildid := 0

	fn := fmt.Sprintf("%s.%x.%x.%d", coverage.CounterFilePref, metaHash, buildid, time.Now().UnixNano())
	s.cfname = filepath.Join(s.outdir, fn)
	var err error
	s.cf, err = os.OpenFile(s.cfname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening %s: %v\n", s.cfname, err)
		return false
	}
	return true
}

// openOutputFiles opens output files in preparation for emitting
// coverage data. In the case of the meta-data file, openOutputFiles
// may determine that we can reuse an existing meta-data file in the
// outdir, in which case it will leave the 'mf' field in the state
// struct as nil. If a new meta-file is needed, the field 'mfname'
// will be the final desired path of the meta file, 'mftmp' will be a
// temporary file, and 'mf' will be an open os.File pointer for
// 'mftmp'. The idea is that the client/caller will write content into
// 'mf', close it, and then rename 'mftmp' to 'mfname'. This function
// also opens the counter data output file, setting 'cf' and 'cfname'
// in the state struct.
func (s *emitState) openOutputFiles(metaHash [16]byte, metaLen uint64) bool {
	s.outdir = os.Getenv("GOCOVERDIR")
	if s.outdir == "" {
		fmt.Fprintf(os.Stderr, "warning: GOCOVERDIR not set, no coverage data emtted\n")
		return false
	}
	fi, err := os.Stat(s.outdir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: GOCOVERDIR setting %q inaccessible (err: %v); no coverage data emtted\n", s.outdir, err)
		return false
	}
	if !fi.IsDir() {
		fmt.Fprintf(os.Stderr, "warning: GOCOVERDIR setting %q not a directory; no coverage data emtted\n", s.outdir)
		return false
	}

	if !s.openMeta(metaHash, metaLen) {
		return false
	}
	if !s.openCounter(metaHash) {
		return false
	}
	return true
}

func (s *emitState) emitMetaDataFile(tlen uint64, finalHash [16]byte) bool {
	mfw := encodemeta.NewCoverageMetaFileWriter(s.mftmp, s.mf)

	blobs := [][]byte{}
	var sd []byte
	bufHdr := (*reflect.SliceHeader)(unsafe.Pointer(&sd))
	for _, e := range s.metalist {
		bufHdr.Data = uintptr(unsafe.Pointer(e.P))
		bufHdr.Len = int(e.Len)
		bufHdr.Cap = int(e.Len)
		blobs = append(blobs, sd)
	}
	err := mfw.Write(finalHash, blobs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error writing %s: %v\n", s.mftmp, err)
		return false
	}

	// Temp file has now been flushed and closed. Rename the temp to the
	// final desired path.
	if err = os.Rename(s.mftmp, s.mfname); err != nil {
		fmt.Fprintf(os.Stderr, "error writing %s: rename from %s failed: %v\n", s.mfname, s.mftmp, err)
		return false
	}

	// We're done.
	return true
}

type counterVisFcn func(pkid uint32, funcid uint32, counters []uint32) bool

func (s *emitState) counterVisitor(f counterVisFcn) bool {
	var sd []uint32
	bufHdr := (*reflect.SliceHeader)(unsafe.Pointer(&sd))

	for _, c := range s.counterlist {
		bufHdr.Data = uintptr(unsafe.Pointer(c.Counters))
		bufHdr.Len = int(c.Len)
		bufHdr.Cap = int(c.Len)
		i := 0
		for i < len(sd) {
			// Skip until we find a non-zero value.
			for i < len(sd) && sd[i] == 0 {
				i++
			}
			if i >= len(sd) {
				break
			}
			if s.debug {
				fmt.Fprintf(os.Stderr, "=+= %d: visit live fcn nc=%d\n",
					i, sd[i])
			}

			// We found a function that was executed. Visit it.
			nCtrs := sd[i]
			pkgId := sd[i+1]
			funcId := sd[i+2]
			counters := sd[i+3 : i+3+int(nCtrs)]
			if !f(pkgId, funcId, counters) {
				return false
			}
			i += 3 + int(nCtrs)
		}
	}
	return true
}

// emitCounterDataFile emits the counter data portion of a
// coverage output file (to the os.File 's.cf').
func (s *emitState) emitCounterDataFile(finalHash [16]byte) bool {

	// Notes:
	// - this version writes everything little-endian, which means
	//   a call is needed encode every value (expensive)
	// - only 'live' functions are written, which requires extra
	//   work here but keeps the generated file smaller.

	u32sz := uint64(unsafe.Sizeof(uint32(1)))
	etot := uint64(0)
	tot := uint64(unsafe.Sizeof(coverage.CounterFileHeader{}))
	sizer := func(pkid uint32, funcid uint32, counters []uint32) bool {
		etot++
		tot += uint64((3 + len(counters)) * int(u32sz))
		return true
	}

	// Compute output file size info.
	s.counterVisitor(sizer)

	w := bufio.NewWriter(s.cf)

	// Emit header. At the moment we're emitting everything
	// little-endian, but we want to leave the door open for
	// the possibility of emitting using native endianity and
	// then having the reader adjust accordingly.
	ch := coverage.CounterFileHeader{
		Magic:       coverage.CovCounterMagic,
		TotalLength: tot,
		Entries:     etot,
		MetaHash:    finalHash,
		BigEndian:   false,
	}
	if s.debug {
		fmt.Fprintf(os.Stderr, "=+= ch: %+v\n", ch)
	}
	var err error
	if err = binary.Write(w, binary.LittleEndian, ch); err != nil {
		fmt.Fprintf(os.Stderr, "error writing %s: %v\n", s.cfname, err)
		return false
	}

	ctrb := make([]byte, u32sz)
	wru32 := func(val uint32) bool {
		binary.LittleEndian.PutUint32(ctrb, val)
		var sz int
		if sz, err = w.Write(ctrb); err != nil || sz != int(u32sz) {
			fmt.Fprintf(os.Stderr, "error writing %s: %v\n", s.cfname, err)
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
	if !s.counterVisitor(emitter) {
		return false
	}

	// Flush and close.
	if err = w.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "error writing %s: %v\n", s.cfname, err)
		return false
	}
	if err = s.cf.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "error closing %s: %v\n", s.cfname, err)
		return false
	}
	return true
}
