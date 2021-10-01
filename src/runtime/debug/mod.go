// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package debug

import (
	"fmt"
	"runtime"
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
	info, ok = parseBuildInfo(data)
	if ok {
		info.GoVersion = runtime.Version()
	}
	return info, ok
}

// BuildInfo represents the build information read from a Go binary.
type BuildInfo struct {
	GoVersion string         // Version of Go that produced this binary.
	Path      string         // The main package path
	Main      Module         // The module containing the main package
	Deps      []*Module      // Module dependencies
	Settings  []BuildSetting // Other information about the build.
}

type BuildSetting struct {
	Key, Value string
}

func (bi *BuildInfo) String() string {
	sb := &strings.Builder{}
	if bi.GoVersion != "" {
		fmt.Fprintf(sb, "go\t%s\n", bi.GoVersion)
	}
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
	for _, s := range bi.Settings {
		fmt.Fprintf(sb, "build\t%s\t%s\n", s.Key, s.Value)
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

func parseBuildInfo(data string) (*BuildInfo, bool) {
	readMod := func(elem []string) (Module, bool) {
		if len(elem) != 2 && len(elem) != 3 {
			return Module{}, false
		}
		sum := ""
		if len(elem) == 3 {
			sum = elem[2]
		}
		return Module{
			Path:    elem[0],
			Version: elem[1],
			Sum:     sum,
		}, true
	}

	var (
		info = &BuildInfo{}
		last *Module
		line string
		ok   bool
	)
	// Reverse of cmd/go/internal/modload.PackageBuildInfo
	for len(data) > 0 {
		i := strings.IndexByte(data, '\n')
		if i < 0 {
			break
		}
		line, data = data[:i], data[i+1:]
		i = strings.IndexByte(line, '\t')
		if i <= 0 {
			continue
		}
		word, rest := line[:i], line[i+1:]

		switch word {
		case "path":
			info.Path = rest
		case "mod":
			elem := strings.Split(rest, "\t")
			last = &info.Main
			*last, ok = readMod(elem)
			if !ok {
				return nil, false
			}
		case "dep":
			elem := strings.Split(rest, "\t")
			last = new(Module)
			info.Deps = append(info.Deps, last)
			*last, ok = readMod(elem)
			if !ok {
				return nil, false
			}
		case "=>":
			elem := strings.Split(rest, "\t")
			if len(elem) != 3 {
				return nil, false
			}
			if last == nil {
				return nil, false
			}
			last.Replace = &Module{
				Path:    elem[0],
				Version: elem[1],
				Sum:     elem[2],
			}
			last = nil
		case "build":
			i := strings.IndexByte(rest, '\t')
			if i < 0 {
				return nil, false
			}
			info.Settings = append(info.Settings, BuildSetting{
				Key:   rest[:i],
				Value: rest[i+1:],
			})
		}
	}
	return info, true
}
