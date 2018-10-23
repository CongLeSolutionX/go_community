// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package debug

import (
	"strings"
)

// set using cmd/go/internal/modload.ModInfoProg
var modinfo string

var (
	infoStart = hexDecodeString("3077af0c9274080241e1c107e6d618e6")
	infoEnd   = hexDecodeString("f932433186182072008242104116d8f2")
)

// ReadBuildInfo returns the build information embedded
// in the running binary. The information is available only
// in binaries built with module support.
func ReadBuildInfo() (info *BuildInfo, ok bool) {
	return readBuildInfo(modinfo)
}

// BuildInfo represents the build information read from
// the running binary.
type BuildInfo struct {
	Path string    // The main package path
	Main Module    // The main module information
	Deps []*Module // Module dependencies
}

// Module represents a module.
type Module struct {
	Path    string  // module path
	Version string  // module version
	Sum     string  // checksum
	Replace *Module // replaced by this module
}

func readBuildInfo(data string) (*BuildInfo, bool) {
	i := strings.Index(data, infoStart)
	if i < 0 {
		return nil, false
	}
	j := strings.Index(data[i:], infoEnd)
	if j < 0 {
		return nil, false
	}
	data = data[i+len(infoStart) : i+j]

	const (
		pathLine = "path\t"
		modLine  = "mod\t"
		depLine  = "dep\t"
		repLine  = "=>\t"
	)

	info := &BuildInfo{}

	var line string
	// Reverse of cmd/go/internal/modload.PackageBuildInfo
	for len(data) > 0 {
		i := strings.IndexByte(data, '\n')
		if i < 0 {
			break
		}
		line, data = data[:i], data[i+1:]
		switch {
		case strings.HasPrefix(line, pathLine):
			elem := line[len(pathLine):]
			info.Path = elem
		case strings.HasPrefix(line, modLine):
			elem := strings.Split(line[len(modLine):], "\t")
			if len(elem) != 3 {
				return nil, false
			}
			info.Main = Module{
				Path:    elem[0],
				Version: elem[1],
				Sum:     elem[2],
			}
		case strings.HasPrefix(line, depLine):
			elem := strings.Split(line[len(depLine):], "\t")
			if len(elem) != 2 && len(elem) != 3 {
				return nil, false
			}
			sum := ""
			if len(elem) == 3 {
				sum = elem[2]
			}
			info.Deps = append(info.Deps, &Module{
				Path:    elem[0],
				Version: elem[1],
				Sum:     sum,
			})
		case strings.HasPrefix(line, repLine):
			elem := strings.Split(line[len(repLine):], "\t")
			if len(elem) != 3 {
				return nil, false
			}
			last := len(info.Deps) - 1
			if last < 0 {
				return nil, false
			}
			info.Deps[last].Replace = &Module{
				Path:    elem[0],
				Version: elem[1],
				Sum:     elem[2],
			}
		}
	}
	return info, true
}

// Similar to encoding/hex.DecodeString.
// We don't use encoding/hex package to avoid dependency cycle
//   encoding/hex (test) -> testing -> runtime/debug.
func hexDecodeString(src string) string {
	if len(src)%2 == 1 {
		panic("invalid length")
	}

	dst := make([]byte, len(src)/2)
	for i := 0; i < len(src)/2; i++ {
		a, ok := fromHexChar(src[i*2])
		if !ok {
			panic("invalid byte error")
		}
		b, ok := fromHexChar(src[i*2+1])
		if !ok {
			panic("invalid byte error")
		}
		dst[i] = (a << 4) | b
	}
	return string(dst)
}

func fromHexChar(c byte) (byte, bool) {
	switch {
	case '0' <= c && c <= '9':
		return c - '0', true
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10, true
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10, true
	}

	return 0, false
}
