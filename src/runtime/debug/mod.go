// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package debug

import (
	"fmt"
	"strings"
)

// exported from runtime
func modinfo() string

// ReadBuildInfo returns the build information embedded
// in the running binary. The information is available only
// in binaries built with module support.
func ReadBuildInfo() (info *BuildInfo, ok bool) {
	data := modinfo()
	if len(data) < 32 {
		return nil, false
	}
	data = data[16 : len(data)-16]
	info, err := ParseBuildInfo(data)
	return info, err == nil
}

// BuildInfo represents the build information read from a Go binary.
type BuildInfo struct {
	Path string    // The main package path
	Main Module    // The module containing the main package
	Deps []*Module // Module dependencies
}

func (bi *BuildInfo) String() string {
	sb := &strings.Builder{}
	if bi.Path != "" {
		fmt.Fprintf(sb, "path\t%s\n", bi.Path)
	}
	var formatMod func(string, Module)
	formatMod = func(word string, m Module) {
		sb.WriteString(word)
		sb.WriteByte('\t')
		sb.WriteString(m.Path)
		mv := m.Version
		if mv == "" {
			mv = "(devel)"
		}
		sb.WriteByte('\t')
		sb.WriteString(mv)
		if m.Replace == nil {
			sb.WriteByte('\t')
			sb.WriteString(m.Sum)
		} else {
			sb.WriteByte('\n')
			formatMod("=>", *m.Replace)
		}
		sb.WriteByte('\n')
	}
	if bi.Main.Path != "" {
		formatMod("mod", bi.Main)
	}
	for _, dep := range bi.Deps {
		formatMod("dep", *dep)
	}

	return sb.String()
}

// Module represents a module.
type Module struct {
	Path    string  // module path
	Version string  // module version
	Sum     string  // checksum
	Replace *Module // replaced by this module
}

// ParseBuildInfo parses a build info string obtained from a Go binary.
func ParseBuildInfo(data string) (info *BuildInfo, err error) {
	lineNum := 1
	defer func() {
		if err != nil {
			err = fmt.Errorf("could not parse Go build info: line %d: %w", lineNum, err)
		}
	}()

	const (
		pathLine = "path\t"
		modLine  = "mod\t"
		depLine  = "dep\t"
		repLine  = "=>\t"
	)

	readModuleLine := func(elem []string) (Module, error) {
		if len(elem) != 2 && len(elem) != 3 {
			return Module{}, fmt.Errorf("expected 2 or 3 columns; got %d", len(elem))
		}
		sum := ""
		if len(elem) == 3 {
			sum = elem[2]
		}
		return Module{
			Path:    elem[0],
			Version: elem[1],
			Sum:     sum,
		}, nil
	}

	info = &BuildInfo{}
	var (
		last *Module
		line string
		ok   bool
	)
	// Reverse of BuildInfo.String()
	for len(data) > 0 {
		line, data, ok = strings.Cut(data, "\n")
		if !ok {
			break
		}
		switch {
		case strings.HasPrefix(line, pathLine):
			elem := line[len(pathLine):]
			info.Path = elem
		case strings.HasPrefix(line, modLine):
			elem := strings.Split(line[len(modLine):], "\t")
			last = &info.Main
			*last, err = readModuleLine(elem)
			if err != nil {
				return nil, err
			}
		case strings.HasPrefix(line, depLine):
			elem := strings.Split(line[len(depLine):], "\t")
			last = new(Module)
			info.Deps = append(info.Deps, last)
			*last, err = readModuleLine(elem)
			if err != nil {
				return nil, err
			}
		case strings.HasPrefix(line, repLine):
			elem := strings.Split(line[len(repLine):], "\t")
			if len(elem) != 3 {
				return nil, fmt.Errorf("expected 3 columns for replacement; got %d", len(elem))
			}
			if last == nil {
				return nil, fmt.Errorf("replacement with no module on previous line")
			}
			last.Replace = &Module{
				Path:    elem[0],
				Version: elem[1],
				Sum:     elem[2],
			}
			last = nil
		}
		lineNum++
	}
	return info, nil
}
