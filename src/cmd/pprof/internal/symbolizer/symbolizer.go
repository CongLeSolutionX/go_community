// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package symbolizer provides a routine to populate a profile with
// symbol, file and line number information. It relies on the
// addr2liner and demangler packages to do the actual work.
package symbolizer

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"cmd/pprof/internal/plugin"
	"cmd/pprof/internal/profile"
)

// Symbolize adds symbol and line number information to all locations
// in a profile. mode enables some options to control
// symbolization. Currently only recognizes "force", which causes it
// to overwrite any existing data.
func Symbolize(mode string, prof *profile.Profile, obj plugin.ObjTool, ui plugin.UI) error {
	force := false
	// Disable some mechanisms based on mode string.
	for _, o := range strings.Split(strings.ToLower(mode), ":") {
		switch o {
		case "force":
			force = true
		default:
		}
	}

	if len(prof.Mapping) == 0 {
		return fmt.Errorf("no known mappings")
	}

	mt, err := newMapping(prof, obj, ui, force)
	if err != nil {
		return err
	}
	defer mt.close()

	functions := make(map[profile.Function]*profile.Function)
	for _, l := range mt.prof.Location {
		m := l.Mapping
		segment := mt.segments[m]
		if segment == nil {
			// Nothing to do
			continue
		}

		stack, err := segment.SourceLine(l.Address)
		if err != nil || len(stack) == 0 {
			// No answers from addr2line
			continue
		}

		l.Line = make([]profile.Line, len(stack))
		for i, frame := range stack {
			if frame.Func != "" {
				m.HasFunctions = true
			}
			if frame.File != "" {
				m.HasFilenames = true
			}
			if frame.Line != 0 {
				m.HasLineNumbers = true
			}
			f := &profile.Function{
				Name:       frame.Func,
				SystemName: frame.Func,
				Filename:   frame.File,
			}
			if fp := functions[*f]; fp != nil {
				f = fp
			} else {
				functions[*f] = f
				f.ID = uint64(len(mt.prof.Function)) + 1
				mt.prof.Function = append(mt.prof.Function, f)
			}
			l.Line[i] = profile.Line{
				Function: f,
				Line:     int64(frame.Line),
			}
		}

		if len(stack) > 0 {
			m.HasInlineFrames = true
		}
	}
	return nil
}

// newMapping creates a mappingTable for a profile.
func newMapping(prof *profile.Profile, obj plugin.ObjTool, ui plugin.UI, force bool) (*mappingTable, error) {
	mt := &mappingTable{
		prof:     prof,
		segments: make(map[*profile.Mapping]plugin.ObjFile),
	}

	// Identify used mappings
	mappings := make(map[*profile.Mapping]bool)
	for _, l := range prof.Location {
		mappings[l.Mapping] = true
	}

	for _, m := range prof.Mapping {
		if !mappings[m] {
			continue
		}
		// Do not attempt to re-symbolize a mapping that has already been symbolized.
		if !force && (m.HasFunctions || m.HasFilenames || m.HasLineNumbers) {
			continue
		}

		f, err := locateFile(obj, m.File, m.BuildID, m.Start)
		if err != nil {
			ui.PrintErr("Local symbolization failed for ", filepath.Base(m.File), ": ", err)
			// Move on to other mappings
			continue
		}

		if fid := f.BuildID(); m.BuildID != "" && fid != "" && fid != m.BuildID {
			// Build ID mismatch - ignore.
			f.Close()
			continue
		}

		mt.segments[m] = f
	}

	return mt, nil
}

// locateFile opens a local file for symbolization on the search path
// at $PPROF_BINARY_PATH. Looks inside these directories for files
// named $BUILDID/$BASENAME and $BASENAME (if build id is available).
func locateFile(obj plugin.ObjTool, file, buildID string, start uint64) (plugin.ObjFile, error) {
	// Construct search path to examine
	searchPath := os.Getenv("PPROF_BINARY_PATH")
	if searchPath == "" {
		// Use $HOME/pprof/binaries as default directory for local symbolization binaries
		searchPath = filepath.Join(os.Getenv("HOME"), "pprof", "binaries")
	}

	// Collect names to search: {buildid/basename, basename}
	var fileNames []string
	if baseName := filepath.Base(file); buildID != "" {
		fileNames = []string{filepath.Join(buildID, baseName), baseName}
	} else {
		fileNames = []string{baseName}
	}
	for _, path := range filepath.SplitList(searchPath) {
		for nameIndex, name := range fileNames {
			file := filepath.Join(path, name)
			if f, err := obj.Open(file, start); err == nil {
				fileBuildID := f.BuildID()
				if buildID == "" || buildID == fileBuildID {
					return f, nil
				}
				f.Close()
				if nameIndex == 0 {
					// If this is the first name, the path includes the build id. Report inconsistency.
					return nil, fmt.Errorf("found file %s with inconsistent build id %s", file, fileBuildID)
				}
			}
		}
	}
	// Try original file name
	f, err := obj.Open(file, start)
	if err == nil && buildID != "" {
		if fileBuildID := f.BuildID(); fileBuildID != "" && fileBuildID != buildID {
			// Mismatched build IDs, ignore
			f.Close()
			return nil, fmt.Errorf("mismatched build ids %s != %s", fileBuildID, buildID)
		}
	}
	return f, err
}

// mappingTable contains the mechanisms for symbolization of a
// profile.
type mappingTable struct {
	prof     *profile.Profile
	segments map[*profile.Mapping]plugin.ObjFile
}

// Close releases any external processes being used for the mapping.
func (mt *mappingTable) close() {
	for _, segment := range mt.segments {
		segment.Close()
	}
}

// SymbolizeFromProfile extracts symbol info from a legacy profile (if present).
// The symbol info has the following form:
// # 0x4d8348 os.Open+0x48 /usr/local/go/src/os/file.go:246
func SymbolizeFromProfile(prof *profile.Profile, source string) error {
	f, err := os.Open(source)
	if err != nil {
		return err
	}
	defer f.Close()
	locRE := regexp.MustCompile(`#\s+0x([[:xdigit:]]+)\s+(([[:alnum:]]|\pN|[./()*])+)\+0x[[:xdigit:]]+\s+([^:]+):([0-9]+)`)
	lines := make(map[uint64]profile.Line)
	funcs := make(map[string]*profile.Function)
	s := bufio.NewScanner(f)
	for s.Scan() {
		match := locRE.FindStringSubmatch(s.Text())
		if match == nil {
			continue
		}
		pc, err := strconv.ParseUint(match[1], 16, 64)
		if err != nil {
			continue
		}
		if _, ok := lines[pc]; ok {
			continue
		}
		line, err := strconv.ParseUint(match[5], 10, 64)
		if err != nil {
			continue
		}
		fn := match[2]
		f := funcs[fn]
		if f == nil {
			f = &profile.Function{ID: uint64(len(prof.Function) + 1), Name: fn, SystemName: fn, Filename: match[4]}
			prof.Function = append(prof.Function, f)
			funcs[fn] = f
		}
		lines[pc] = profile.Line{Function: f, Line: int64(line)}
	}
	if err := s.Err(); err != nil {
		return err
	}
	for _, l := range prof.Location {
		line, ok := lines[l.Address]
		if ok {
			l.Line = []profile.Line{line}
			continue
		}
		line, ok = lines[l.Address+1]
		if ok {
			l.Line = []profile.Line{line}
			continue
		}
	}
	return nil
}
