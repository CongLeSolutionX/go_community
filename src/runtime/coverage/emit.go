// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package coverage

import (
	"crypto/md5"
	"fmt"
	"internal/coverage"
	"internal/coverage/encodecounter"
	"internal/coverage/encodemeta"
	"io"
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
	p       *byte
	len     uint32
	hash    [16]byte
	pkgpath string
	pkid    int
	cmode   coverage.CounterMode
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

//go:linkname runtime_getcovpkgmap runtime.getcovpkgmap
func runtime_getcovpkgmap() map[int]int

//go:linkname runtime_reportInsanityInHardcodedList runtime.reportInsanityInHardcodedList
func runtime_reportInsanityInHardcodedList(slot int32, pkgId int32)

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

	// Table to use for remapping hard-coded pkg ids.
	pkgmap map[int]int

	// emit debug trace output
	debug bool
}

// finalHash is computed at init time from the list of meta-data
// symbols registered during init. It is used both for writing the
// meta-data file and counter-aata files.
var finalHash [16]byte
var metaDataEmitAttempted bool
var finalHashComputed bool
var finalMetaLen uint64
var cmode coverage.CounterMode
var gocoverdir string

type fileType int

const (
	noFile = 1 << iota
	metaDataFile
	counterDataFile
)

// emitMetaData emits the meta-data output file for this coverage run.
// This entry point is intended to be invoked by the compiler from
// an instrumented program's main package init func.
func emitMetaData() {
	gocoverdir = os.Getenv("GOCOVERDIR")
	if gocoverdir == "" {
		fmt.Fprintf(os.Stderr, "warning: GOCOVERDIR not set, no coverage data emitted\n")
		return
	}
	if err := emitMetaDataToDirectory(gocoverdir); err != nil {
		fmt.Fprintf(os.Stderr, "error: coverage meta-data emit failed: %v\n", err)
		if os.Getenv("GOCOVERDEBUG") != "" {
			panic("meta-data write failure")
		}
	}
}

func modeClash(m coverage.CounterMode) bool {
	if m == coverage.CtrModeRegOnly {
		return false
	}
	if cmode == coverage.CtrModeInvalid {
		cmode = m
		return false
	} else {
		return cmode != m
	}
}

// emitMetaData emits the meta-data output file to the specified
// directory, returning an error if something went wrong.
func emitMetaDataToDirectory(outdir string) error {

	metaDataEmitAttempted = true

	// Ask the runtime for the list of coverage meta-data symbols.
	ml := runtime_getcovmetalist()

	if len(ml) == 0 {
		return fmt.Errorf("program not built with -coverage")
	}

	s := &emitState{
		metalist: ml,
		debug:    os.Getenv("GOCOVERDEBUG") != "",
		outdir:   outdir,
	}

	if s.debug {
		fmt.Fprintf(os.Stderr, "=+= contents of covmetalist:\n")
		for k, b := range ml {
			fmt.Fprintf(os.Stderr, "=+= slot: %d path: %s ", k, b.pkgpath)
			if b.pkid != -1 {
				fmt.Fprintf(os.Stderr, " hcid: %d", b.pkid)
			}
			fmt.Fprintf(os.Stderr, "\n")
		}
		pm := runtime_getcovpkgmap()
		fmt.Fprintf(os.Stderr, "=+= remap table:\n")
		for from, to := range pm {
			fmt.Fprintf(os.Stderr, "=+= from %d to %d\n",
				uint32(from), uint32(to))
		}
	}

	h := md5.New()
	tlen := uint64(unsafe.Sizeof(coverage.MetaFileHeader{}))
	for _, entry := range ml {
		if _, err := h.Write(entry.hash[:]); err != nil {
			// Is this the right away report this error?
			panic(fmt.Sprintf("internal error: md5 sum failed: %v", err))
		}
		tlen += uint64(entry.len)
		if modeClash(entry.cmode) {
			return fmt.Errorf("coverage counter mode clash: package %s uses mode=%d, but package %s uses mode=%s\n", ml[0].pkgpath, cmode, entry.pkgpath, entry.cmode)
		}
	}

	fh := h.Sum(nil)
	copy(finalHash[:], fh)
	finalHashComputed = true
	finalMetaLen = tlen

	// Open output files.
	if err := s.openOutputFiles(finalHash, tlen, metaDataFile); err != nil {
		return err
	}

	// Emit meta-data file only if needed (may already be present).
	if s.mf != nil {
		if err := s.emitMetaDataFile(tlen, finalHash); err != nil {
			return err
		}
	}
	return nil
}

// emitCounterData emits the counter data output file for this coverage run.
// This entry point is intended to be invoked by the runtime when an
// instrumented program is terminating or calling os.Exit().
func emitCounterData() {
	if gocoverdir == "" || !finalHashComputed {
		return
	}
	if err := emitCounterDataToDirectory(gocoverdir); err != nil {
		fmt.Fprintf(os.Stderr, "error: coverage counter data emit failed: %v\n", err)
		if os.Getenv("GOCOVERDEBUG") != "" {
			panic("counter-data write failure")
		}
	}
}

// emitMetaData emits the counter-data output file for this coverage run.
func emitCounterDataToDirectory(outdir string) error {

	// Ask the runtime for the list of coverage counter symbols.
	cl := runtime_getcovcounterlist()
	if len(cl) == 0 {
		return fmt.Errorf("program not built with -coverage")
	}

	// Ask the runtime for the list of coverage counter symbols.
	pm := runtime_getcovpkgmap()
	s := &emitState{
		counterlist: cl,
		pkgmap:      pm,
		outdir:      outdir,
		debug:       os.Getenv("GOCOVERDEBUG") != "",
	}

	// Open output file.
	if err := s.openOutputFiles(finalHash, finalMetaLen, counterDataFile); err != nil {
		return err
	}
	if s.cf == nil {
		return fmt.Errorf("counter data output file open failed (no additional info")
	}

	// Emit counter data file.
	if err := s.emitCounterDataFile(finalHash, s.cf); err != nil {
		return err
	}
	if err := s.cf.Close(); err != nil {
		return fmt.Errorf("closing counter data file: %v", err)
	}
	return nil
}

// emitMetaData emits counter data for this coverage run to an io.Writer.
func (s *emitState) emitCounterDataToWriter(w io.Writer) error {
	if err := s.emitCounterDataFile(finalHash, w); err != nil {
		return err
	}
	return nil
}

// openMetaFile determines whether we need to emit a meta-data output
// file, or whether we can reuse the existing file in the coverage out
// dir. It updates mfname/mftmp/mf fields in 'od', returning false on
// error and true for success.
func (s *emitState) openMetaFile(metaHash [16]byte, metaLen uint64) error {

	// Open meta-outfile for reading to see if it exists.
	fn := fmt.Sprintf("%s.%x", coverage.MetaFilePref, metaHash)
	s.mfname = filepath.Join(s.outdir, fn)
	fi, err := os.Stat(s.mfname)
	if err != nil || fi.Size() != int64(metaLen) {
		// We need a new meta-file.
		s.mftmp = s.mfname + fmt.Sprintf("%d", time.Now().UnixNano())
		s.mf, err = os.OpenFile(s.mftmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			return fmt.Errorf("opening meta-data file %s: %v", s.mftmp, err)
		}
	}
	return nil
}

// openCounterFile opens an output file for the counter data portion of a
// test coverage run. If updates the 'cfname' and 'cf' fields in the
// passed in state struct, returning an error if something went wrong.
func (s *emitState) openCounterFile(metaHash [16]byte) error {
	processID := os.Getpid()
	fn := fmt.Sprintf(coverage.CounterFileTempl, coverage.CounterFilePref, metaHash, processID, time.Now().UnixNano())
	s.cfname = filepath.Join(s.outdir, fn)
	var err error
	s.cf, err = os.OpenFile(s.cfname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return fmt.Errorf("opening counter data file %s: %v", s.cfname, err)
	}
	return nil
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
func (s *emitState) openOutputFiles(metaHash [16]byte, metaLen uint64, which fileType) error {
	fi, err := os.Stat(s.outdir)
	if err != nil {
		return fmt.Errorf("output directory %q inaccessible (err: %v); no coverage data emtted", s.outdir, err)
	}
	if !fi.IsDir() {
		return fmt.Errorf("output directory %q not a directory; no coverage data emtted", s.outdir)
	}

	if (which & metaDataFile) != 0 {
		if err := s.openMetaFile(metaHash, metaLen); err != nil {
			return err
		}
	}
	if (which & counterDataFile) != 0 {
		if err := s.openCounterFile(metaHash); err != nil {
			return err
		}
	}
	return nil
}

// emitMetaDataFile emits coverage meta-data to a previously opened
// temporary file (s.mftmp), then renames the generated file to the
// final path (s.mfname).
func (s *emitState) emitMetaDataFile(tlen uint64, finalHash [16]byte) error {
	if err := emitMetaDataToWriter(s.mf, s.metalist, cmode, finalHash); err != nil {
		return fmt.Errorf("writing %s: %v\n", s.mftmp, err)
	}
	if err := s.mf.Close(); err != nil {
		return fmt.Errorf("closing meta data temp file: %v", err)
	}

	// Temp file has now been flushed and closed. Rename the temp to the
	// final desired path.
	if err := os.Rename(s.mftmp, s.mfname); err != nil {
		return fmt.Errorf("writing %s: rename from %s failed: %v\n", s.mfname, s.mftmp, err)
	}

	// We're done.
	return nil
}

func emitMetaDataToWriter(w io.Writer, metalist []covmetablob, cmode coverage.CounterMode, finalHash [16]byte) error {
	mfw := encodemeta.NewCoverageMetaFileWriter("<io.Writer>", w)

	blobs := [][]byte{}
	var sd []byte
	bufHdr := (*reflect.SliceHeader)(unsafe.Pointer(&sd))
	for _, e := range metalist {
		bufHdr.Data = uintptr(unsafe.Pointer(e.p))
		bufHdr.Len = int(e.len)
		bufHdr.Cap = int(e.len)
		blobs = append(blobs, sd)
	}
	moduleName := ""
	bip, ok := debug.ReadBuildInfo()
	if ok {
		moduleName = bip.Main.Path
	}
	return mfw.Write(finalHash, moduleName, blobs, cmode)
}

func (s *emitState) NumFuncs() (int, error) {
	var sd []uint32
	bufHdr := (*reflect.SliceHeader)(unsafe.Pointer(&sd))

	totalFuncs := 0
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

			// We found a function that was executed.
			nCtrs := sd[i]
			totalFuncs++

			// Skip over this function.
			i += 3 + int(nCtrs)
		}
	}
	return totalFuncs, nil
}

func (s *emitState) VisitFuncs(f encodecounter.CounterVisitorFcn) error {
	var sd []uint32
	bufHdr := (*reflect.SliceHeader)(unsafe.Pointer(&sd))

	dpkg := uint32(0)
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

			// We found a function that was executed.
			nCtrs := sd[i]
			pkgId := sd[i+1]
			funcId := sd[i+2]
			counters := sd[i+3 : i+3+int(nCtrs)]

			if s.debug {
				if pkgId != dpkg {
					dpkg = pkgId
					fmt.Fprintf(os.Stderr, "\n=+= %d: pk=%d visit live fcn",
						i, pkgId)
				}
				fmt.Fprintf(os.Stderr, " {i=%d F%d NC%d}", i, funcId, nCtrs)
			}

			// Vet and/or fix up package ID. A package ID of zero
			// indicates that there is some new package X that is a
			// runtime dependency, and this package has code that
			// executes before its corresponding init package runs.
			// This is a fatal error that we should only see during
			// Go development (e.g. tip).
			ipk := int32(pkgId)
			if ipk == 0 {
				fmt.Fprintf(os.Stderr, "\n")
				runtime_reportInsanityInHardcodedList(int32(i), ipk)
			} else if ipk < 0 {
				if newId, ok := s.pkgmap[int(ipk)]; ok {
					pkgId = uint32(newId)
				} else {
					fmt.Fprintf(os.Stderr, "\n")
					runtime_reportInsanityInHardcodedList(int32(i), ipk)
				}
			} else {
				pkgId--
			}

			if err := f(pkgId, funcId, counters); err != nil {
				return err
			}
			i += 3 + int(nCtrs)
		}
		if s.debug {
			fmt.Fprintf(os.Stderr, "\n")
		}
	}
	return nil
}

// captureOsArgs captures os.Args() into the format we use to store
// it in the counter data file.
func captureOsArgs() map[string]string {
	m := make(map[string]string)
	m["argc"] = fmt.Sprintf("%d", len(os.Args))
	for k, a := range os.Args {
		m[fmt.Sprintf("argv%d", k)] = a
	}
	return m
}

// emitCounterDataFile emits the counter data portion of a
// coverage output file (to the file 's.cf').
func (s *emitState) emitCounterDataFile(finalHash [16]byte, w io.Writer) error {

	// FIXME: do we want to copy os.Args early (during init) so as to
	// avoid any user program modifications? Or maybe we want to see
	// modifications.

	cfw := encodecounter.NewCoverageDataFileWriter(w, coverage.CtrULeb128)
	if err := cfw.Write(finalHash, captureOsArgs(), s); err != nil {
		return err
	}
	return nil
}
