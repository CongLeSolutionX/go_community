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
	"internal/coverage/rtcov"
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

// getCovMetaList returns a list of meta-data blobs registered
// for the currently executing instrumented program. It is defined in the
// runtime.
func getCovMetaList() []rtcov.CovMetaBlob

// getCovCounterList returns a list of counter-data blobs registered
// for the currently executing instrumented program. It is defined in the
// runtime.
func getCovCounterList() []rtcov.CovCounterBlob

// getCovPkgMap returns a map storing the remapped package IDs for
// hard-coded runtime packages (see internal/coverage/pkgid.go for
// more on why hard-coded package IDs are needed). This function
// is defined in the runtime.
func getCovPkgMap() map[int]int

// emitState holds useful state information during the emit process.
//
// When an instrumented program finishes execution and starts the
// process of writing out coverage data, it's possible that an
// existing meta-data file already exists in the output directory. In
// this case openOutputFiles() below will leave the 'mf' field below
// as nil. If a new meta-data file is needed, field 'mfname' will be
// the final desired path of the meta file, 'mftmp' will be a
// temporary file, and 'mf' will be an open os.File pointer for
// 'mftmp'. The meta-data file payload will be written to 'mf', the
// temp file will be then closed and renamed (from 'mftmp' to
// 'mfname'), so as to insure that the meta-data file is created
// atomically; we want this so that things work smoothly in cases
// where there are several instances of a given instrumented program
// all terminating at the same time and trying to create meta-data
// files simultaneously.
//
// For counter data files there is less chance of a collision, hence
// the openOutputFiles() stores the counter data file in 'cfname' and
// then places the *io.File into 'cf'.
type emitState struct {
	mfname string   // path of final meta-data output file
	mftmp  string   // path to meta-data temp file (if needed)
	mf     *os.File // open os.File for meta-data temp file
	cfname string   // path of final counter data file
	cf     *os.File // open os.File for counter data file
	outdir string   // output directory

	// List of meta-data symbols obtained from the runtime
	metalist []rtcov.CovMetaBlob

	// List of counter-data symbols obtained from the runtime
	counterlist []rtcov.CovCounterBlob

	// Table to use for remapping hard-coded pkg ids.
	pkgmap map[int]int

	// emit debug trace output
	debug bool
}

var (
	// finalHash is computed at init time from the list of meta-data
	// symbols registered during init. It is used both for writing the
	// meta-data file and counter-data files.
	finalHash [16]byte
	// Set to true when we've computed finalHash + finalMetaLen.
	finalHashComputed bool
	// Total meta-data length.
	finalMetaLen uint64
	// Records whether we've already attempted to write meta-data.
	metaDataEmitAttempted bool
	// Counter mode for this instrumented program run.
	cmode coverage.CounterMode
	// Cached value of GOCOVERDIR environment variable.
	goCoverDir string
	// Copy of os.Args made at init time, converted into map format.
	capturedOsArgs map[string]string
)

// fileType is used to select between counter-data files and
// meta-data files.
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
	goCoverDir = os.Getenv("GOCOVERDIR")
	if goCoverDir == "" {
		fmt.Fprintf(os.Stderr, "warning: GOCOVERDIR not set, no coverage data emitted\n")
		return
	}
	if err := emitMetaDataToDirectory(goCoverDir); err != nil {
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
	}
	return cmode != m
}

// emitMetaData emits the meta-data output file to the specified
// directory, returning an error if something went wrong.
func emitMetaDataToDirectory(outdir string) error {

	metaDataEmitAttempted = true

	// Ask the runtime for the list of coverage meta-data symbols.
	ml := getCovMetaList()

	// In the normal case (go build -o prog.exe ... ; ./prog.exe)
	// len(ml) will always be non-zero, but we check here since at
	// some point this function will be reachable via user-callable
	// APIs (for example, to write out coverage data from a server
	// program that doesn't ever call os.Exit).
	if len(ml) == 0 {
		return fmt.Errorf("program not built with -coverage")
	}

	s := &emitState{
		metalist: ml,
		debug:    os.Getenv("GOCOVERDEBUG") != "",
		outdir:   outdir,
	}

	// Capture os.Args() now so as to avoid issues if args
	// are rewritten during program execution.
	capturedOsArgs = captureOsArgs()

	if s.debug {
		fmt.Fprintf(os.Stderr, "=+= contents of covmetalist:\n")
		for k, b := range ml {
			fmt.Fprintf(os.Stderr, "=+= slot: %d path: %s ", k, b.PkgPath)
			if b.PkgID != -1 {
				fmt.Fprintf(os.Stderr, " hcid: %d", b.PkgID)
			}
			fmt.Fprintf(os.Stderr, "\n")
		}
		pm := getCovPkgMap()
		fmt.Fprintf(os.Stderr, "=+= remap table:\n")
		for from, to := range pm {
			fmt.Fprintf(os.Stderr, "=+= from %d to %d\n",
				uint32(from), uint32(to))
		}
	}

	h := md5.New()
	tlen := uint64(unsafe.Sizeof(coverage.MetaFileHeader{}))
	for _, entry := range ml {
		if _, err := h.Write(entry.Hash[:]); err != nil {
			return err
		}
		tlen += uint64(entry.Len)
		ecm := coverage.CounterMode(entry.CounterMode)
		if modeClash(ecm) {
			return fmt.Errorf("coverage counter mode clash: package %s uses mode=%d, but package %s uses mode=%s\n", ml[0].PkgPath, cmode, entry.PkgPath, ecm)
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
	if s.needMetaDataFile() {
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
	if goCoverDir == "" || !finalHashComputed {
		return
	}
	if err := emitCounterDataToDirectory(goCoverDir); err != nil {
		fmt.Fprintf(os.Stderr, "error: coverage counter data emit failed: %v\n", err)
		if os.Getenv("GOCOVERDEBUG") != "" {
			panic("counter-data write failure")
		}
	}
}

// emitMetaData emits the counter-data output file for this coverage run.
func emitCounterDataToDirectory(outdir string) error {

	// Ask the runtime for the list of coverage counter symbols.
	cl := getCovCounterList()
	if len(cl) == 0 {
		return fmt.Errorf("program not built with -coverage")
	}

	// Ask the runtime for the list of coverage counter symbols.
	pm := getCovPkgMap()
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

// openMetaFile determines whether we need to emit a meta-data output
// file, or whether we can reuse the existing file in the coverage out
// dir. It updates mfname/mftmp/mf fields in 's', returning an error
// if something went wrong. See the comment on the emitState type
// definition above for more on how file opening is managed.
func (s *emitState) openMetaFile(metaHash [16]byte, metaLen uint64) error {

	// Open meta-outfile for reading to see if it exists.
	fn := fmt.Sprintf("%s.%x", coverage.MetaFilePref, metaHash)
	s.mfname = filepath.Join(s.outdir, fn)
	fi, err := os.Stat(s.mfname)
	if err != nil || fi.Size() != int64(metaLen) {
		// We need a new meta-file.
		s.mftmp = s.mfname + fmt.Sprintf("%d", time.Now().UnixNano())
		s.mf, err = os.Create(s.mftmp)
		if err != nil {
			return fmt.Errorf("creating meta-data file %s: %v", s.mftmp, err)
		}
	}
	return nil
}

// openCounterFile opens an output file for the counter data portion
// of a test coverage run. If updates the 'cfname' and 'cf' fields in
// 's', returning an error if something went wrong.
func (s *emitState) openCounterFile(metaHash [16]byte) error {
	processID := os.Getpid()
	fn := fmt.Sprintf(coverage.CounterFileTempl, coverage.CounterFilePref, metaHash, processID, time.Now().UnixNano())
	s.cfname = filepath.Join(s.outdir, fn)
	var err error
	s.cf, err = os.Create(s.cfname)
	if err != nil {
		return fmt.Errorf("creating counter data file %s: %v", s.cfname, err)
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
	if err := writeMetaData(s.mf, s.metalist, cmode, finalHash); err != nil {
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

	return nil
}

// needMetaDataFile returns TRUE if we need to emit a meta-data file
// for this program run. It should be used only after
// openOutputFiles() has been invoked.
func (s *emitState) needMetaDataFile() bool {
	return s.mf != nil
}

func writeMetaData(w io.Writer, metalist []rtcov.CovMetaBlob, cmode coverage.CounterMode, finalHash [16]byte) error {
	mfw := encodemeta.NewCoverageMetaFileWriter("<io.Writer>", w)

	// Note: "sd" is re-initialized on each iteration of the loop
	// below, and would normally be declared inside the loop, but
	// placed here escape analysis since we capture it in bufHdr.
	var sd []byte
	bufHdr := (*reflect.SliceHeader)(unsafe.Pointer(&sd))

	var blobs [][]byte
	for _, e := range metalist {
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
		for i := 0; i < len(sd); i++ {
			// Skip ahead until the next non-zero value.
			if sd[i] == 0 {
				continue
			}

			// We found a function that was executed.
			nCtrs := sd[i]
			totalFuncs++

			// Skip over this function.
			i += coverage.FirstCtrOffset + int(nCtrs) - 1
		}
	}
	return totalFuncs, nil
}

func (s *emitState) VisitFuncs(f encodecounter.CounterVisitorFn) error {
	var sd []uint32
	bufHdr := (*reflect.SliceHeader)(unsafe.Pointer(&sd))

	dpkg := uint32(0)
	for _, c := range s.counterlist {
		bufHdr.Data = uintptr(unsafe.Pointer(c.Counters))
		bufHdr.Len = int(c.Len)
		bufHdr.Cap = int(c.Len)
		for i := 0; i < len(sd); i++ {
			// Skip ahead until the next non-zero value.
			if sd[i] == 0 {
				continue
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
				reportErrorInHardcodedList(int32(i), ipk)
			} else if ipk < 0 {
				if newId, ok := s.pkgmap[int(ipk)]; ok {
					pkgId = uint32(newId)
				} else {
					fmt.Fprintf(os.Stderr, "\n")
					reportErrorInHardcodedList(int32(i), ipk)
				}
			} else {
				// The package ID value stored in the counter array
				// has 1 added to it (so as to preclude the
				// possibility of a zero value ; see
				// runtime.addCovMeta), so subtract off 1 here to form
				// the real package ID.
				pkgId--
			}

			if err := f(pkgId, funcId, counters); err != nil {
				return err
			}

			// Skip over this function.
			i += coverage.FirstCtrOffset + int(nCtrs) - 1
		}
		if s.debug {
			fmt.Fprintf(os.Stderr, "\n")
		}
	}
	return nil
}

// captureOsArgs converts os.Args() into the format we use to store
// this info in the counter data file (counter data file "args"
// section is a generic key-value collection). See the 'args' section
// in internal/coverage/defs.go for more info.
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
	cfw := encodecounter.NewCoverageDataFileWriter(w, coverage.CtrULeb128)
	if err := cfw.Write(finalHash, capturedOsArgs, s); err != nil {
		return err
	}
	return nil
}

func reportErrorInHardcodedList(slot int32, pkgId int32) {
	metaList := getCovMetaList()
	pkgMap := getCovPkgMap()

	println("internal error in coverage meta-data tracking:")
	println("encountered bad pkg ID ", pkgId, " at slot ", slot)
	println("list of hard-coded runtime package IDs needs revising.")
	println("[see the comment on the 'rtPkgs' var in ")
	println(" <goroot>/src/internal/coverage/pkid.go]")
	println("registered list:")
	for k, b := range metaList {
		print("slot: ", k, " path='", b.PkgPath, "' ")
		if b.PkgID != -1 {
			print(" hard-coded id: ", b.PkgID)
		}
		println("")
	}
	println("remap table:")
	for from, to := range pkgMap {
		println("from ", from, " to ", to)
	}
}
