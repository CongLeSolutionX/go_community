// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package symbolizer provides a routine to populate a profile with
// symbol, file and line number information. It relies on the
// addr2liner and demangle packages to do the actual work.
package symbolizer

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"cmd/pprof/internal/plugin"
	"cmd/pprof/internal/symbolz"
	"internal/pprof/profile"
)

// Symbolizer implements the plugin.Symbolize interface.
type Symbolizer struct {
	Obj plugin.ObjTool
	UI  plugin.UI
}

// test taps for dependency injection
var symbolzSymbolize = symbolz.Symbolize
var localSymbolize = doLocalSymbolize

// Symbolize attempts to symbolize profile p. First uses binutils on
// local binaries; if the source is a URL it attempts to get any
// missed entries using symbolz.
func (s *Symbolizer) Symbolize(mode string, sources plugin.MappingSources, p *profile.Profile) error {
	remote, local, force, demanglerMode := true, true, false, ""
	for _, o := range strings.Split(strings.ToLower(mode), ":") {
		switch o {
		case "none", "no":
			return nil
		case "local", "fastlocal":
			remote, local = false, true
		case "remote":
			remote, local = true, false
		case "", "force":
			force = true
		default:
			switch d := strings.TrimPrefix(o, "demangle="); d {
			case "full", "none", "templates":
				demanglerMode = d
				force = true
				continue
			case "default":
				continue
			}
			s.UI.PrintErr("ignoring unrecognized symbolization option: " + mode)
			s.UI.PrintErr("expecting -symbolize=[local|fastlocal|remote|none][:force][:demangle=[none|full|templates|default]")
		}
	}

	var err error
	if local {
		// Symbolize locally.
		if err = localSymbolize(mode, p, s.Obj, s.UI); err != nil {
			s.UI.PrintErr("local symbolization: " + err.Error())
		}
	}
	if remote {
		if err = symbolzSymbolize(sources, postURL, p, s.UI); err != nil {
			return err // Ran out of options.
		}
	}

	Demangle(p, force, demanglerMode)
	return nil
}

// postURL issues a POST to a URL over HTTP.
func postURL(source, post string) ([]byte, error) {
	resp, err := http.Post(source, "application/octet-stream", strings.NewReader(post))
	if err != nil {
		return nil, fmt.Errorf("http post %s: %v", source, err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server response: %s", resp.Status)
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

// doLocalSymbolize adds symbol and line number information to all locations
// in a profile. mode enables some options to control
// symbolization.
func doLocalSymbolize(mode string, prof *profile.Profile, obj plugin.ObjTool, ui plugin.UI) error {
	force := false
	// Disable some mechanisms based on mode string.
	for _, o := range strings.Split(strings.ToLower(mode), ":") {
		switch {
		case o == "force":
			force = true
		default:
		}
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
			// Nothing to do.
			continue
		}

		stack, err := segment.SourceLine(l.Address)
		if err != nil || len(stack) == 0 {
			// No answers from addr2line.
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

// Demangle updates the function names in a profile with demangled C++
// names, simplified according to demanglerMode. If force is set,
// overwrite any names that appear already demangled.
func Demangle(prof *profile.Profile, force bool, demanglerMode string) {
	// No demangle support.
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

	missingBinaries := false
	for midx, m := range prof.Mapping {
		if !mappings[m] {
			continue
		}

		// Do not attempt to re-symbolize a mapping that has already been symbolized.
		if !force && (m.HasFunctions || m.HasFilenames || m.HasLineNumbers) {
			continue
		}

		if m.File == "" {
			if midx == 0 {
				ui.PrintErr("Main binary filename not available.\n" +
					"Try passing the path to the main binary before the profile.")
				continue
			}
			missingBinaries = true
			continue
		}

		// Skip well-known system mappings
		name := filepath.Base(m.File)
		if name == "[vdso]" || strings.HasPrefix(name, "linux-vdso") {
			continue
		}

		// Skip mappings pointing to a source URL
		if m.BuildID == "" {
			if u, err := url.Parse(m.File); err == nil && u.IsAbs() {
				continue
			}
		}

		f, err := obj.Open(m.File, m.Start, m.Limit, m.Offset)
		if err != nil {
			ui.PrintErr("Local symbolization failed for ", name, ": ", err)
			continue
		}
		if fid := f.BuildID(); m.BuildID != "" && fid != "" && fid != m.BuildID {
			ui.PrintErr("Local symbolization failed for ", name, ": build ID mismatch")
			f.Close()
			continue
		}

		mt.segments[m] = f
	}
	if missingBinaries {
		ui.PrintErr("Some binary filenames not available. Symbolization may be incomplete.")
	}
	return mt, nil
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
