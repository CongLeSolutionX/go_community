// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package coverage

import (
	"crypto/md5"
	"fmt"
	"internal/coverage"
	"internal/coverage/encodecounter"
	"internal/coverage/encodemeta"
	"os"
	"path/filepath"
	"reflect"
	"runtime/debug"
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

// finalHash is computed at init time from the list of meta-data
// symbols registered during init. It is used both for writing the
// meta-data file and counter-aata files.
var finalHash [16]byte
var finalHashComputed bool
var finalMetaLen uint64

type fileType int

const (
	noFile = 1 << iota
	metaDataFile
	counterDataFile
)

// emitMetaData emits the meta-data output file for this coverage run.
func emitMetaData() {

	// Ask the runtime for the list of coverage meta-data symbols.
	ml := runtime_getcovmetalist()

	if len(ml) == 0 {
		return
	}

	s := &emitState{
		metalist: ml,
		debug:    os.Getenv("GOCOVERDEBUG") != "",
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
	finalHash = md5.Sum(nil)
	finalHashComputed = true
	finalMetaLen = tlen

	// Open output files.
	s.openOutputFiles(finalHash, tlen, metaDataFile)

	// Emit meta-data file if needed.
	if s.mf != nil {
		if !s.emitMetaDataFile(tlen, finalHash) {
			return
		}
	}
}

// emitMetaData emits the counter-data output file for this coverage run.
func emitCounterData() {

	if !finalHashComputed {
		return
	}

	// Ask the runtime for the list of coverage counter symbols.
	cl := runtime_getcovcounterlist()
	if len(cl) == 0 {
		return
	}

	s := &emitState{
		counterlist: cl,
		debug:       os.Getenv("GOCOVERDEBUG") != "",
	}

	// Open output file.
	s.openOutputFiles(finalHash, finalMetaLen, counterDataFile)
	if s.cf == nil {
		// something went wrong, bail here.
		return
	}

	// Emit counter data file.
	s.emitCounterDataFile(finalHash)
}

// openMetaFile determines whether we need to emit a meta-data output
// file, or whether we can reuse the existing file in the coverage out
// dir. It updates mfname/mftmp/mf fields in 'od', returning false on
// error and true for success.
func (s *emitState) openMetaFile(metaHash [16]byte, metaLen uint64) bool {

	// Open meta-outfile for reading to see if it exists.
	fn := fmt.Sprintf("%s.%x", coverage.MetaFilePref, metaHash)
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

// openCounterFile opens an output file for the counter data portion of a
// test coverage run. If updates the 'cfname' and 'cf' fields in the
// passed in state struct, returning false on error (if there was an
// error opening the file) or true for success.
func (s *emitState) openCounterFile(metaHash [16]byte) bool {

	fn := fmt.Sprintf("%s.%x.%d", coverage.CounterFilePref, metaHash, time.Now().UnixNano())
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
func (s *emitState) openOutputFiles(metaHash [16]byte, metaLen uint64, which fileType) bool {
	s.outdir = os.Getenv("GOCOVERDIR")
	if s.outdir == "" {
		fmt.Fprintf(os.Stderr, "warning: GOCOVERDIR not set, no coverage data emitted\n")
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

	if (which&metaDataFile) != 0 && !s.openMetaFile(metaHash, metaLen) {
		return false
	}
	if (which&counterDataFile) != 0 && !s.openCounterFile(metaHash) {
		return false
	}
	return true
}

// emitMetaDataFile emits coverage meta-data to a previously opened
// temporary file (s.mftmp), then renames the generated file to the
// final path (s.mfname).
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
	moduleName := ""
	bip, ok := debug.ReadBuildInfo()
	if ok {
		moduleName = bip.Main.Path
	}
	err := mfw.Write(finalHash, moduleName, blobs)
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

func (s *emitState) counterVisitor(f encodecounter.CounterVisitorFcn) bool {
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
// coverage output file (to the file 's.cf').
func (s *emitState) emitCounterDataFile(finalHash [16]byte) bool {

	cfw := encodecounter.NewCoverageDataFileWriter(s.cfname, s.cf)

	visitor := func(f encodecounter.CounterVisitorFcn) bool {
		return s.counterVisitor(f)
	}

	defer s.cf.Close()

	return cfw.Write(finalHash, visitor)
}
