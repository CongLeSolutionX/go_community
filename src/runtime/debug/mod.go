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

// ReadModInfo returns the module information embedded
// in the running binary. If the information is unavailable
// because the binary was not built with module support,
// ReadModInfo returns nil.
func ReadModInfo() *ModInfo {
	return readModInfo(modinfo)
}

// ModInfo represents the parsed module information.
type ModInfo struct {
	Path string    // The main package path
	Main Module    // The main module information
	Deps []*Module // Module dependencies
}

type Module struct {
	Path    string
	Version string
	Sum     string
	Replace *Module
}

func readModInfo(data string) *ModInfo {
	s, e := string(infoStart), string(infoEnd)
	i := strings.Index(data, s)
	if i < 0 {
		return nil
	}
	j := strings.Index(data[i:], e)
	if j < 0 {
		return nil
	}
	data = data[i+len(s) : i+j]

	const (
		pathLine = "path\t"
		modLine  = "mod\t"
		depLine  = "dep\t"
		repLine  = "=>\t"
	)

	modInfo := &ModInfo{}

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
			elem := strings.SplitAfterN(line, "\t", 2)
			if len(elem) != 2 {
				return nil
			}
			modInfo.Path = elem[1]
		case strings.HasPrefix(line, modLine):
			elem := strings.SplitAfterN(line, "\t", 4)
			if len(elem) != 4 {
				return nil
			}
			modInfo.Main = Module{
				Path:    elem[1],
				Version: elem[2],
				Sum:     elem[3],
			}
		case strings.HasPrefix(line, depLine):
			elem := strings.SplitAfterN(line, "\t", 4)
			if len(elem) != 3 && len(elem) != 4 {
				return nil
			}
			sum := ""
			if len(elem) == 4 {
				sum = elem[3]
			}
			modInfo.Deps = append(modInfo.Deps, &Module{
				Path:    elem[1],
				Version: elem[2],
				Sum:     sum,
			})
		case strings.HasPrefix(line, repLine):
			elem := strings.SplitAfterN(line, "\t", 4)
			if len(elem) != 4 {
				return nil
			}
			last := len(modInfo.Deps) - 1
			if last < 0 {
				return nil
			}
			modInfo.Deps[last].Replace = &Module{
				Path:    elem[1],
				Version: elem[2],
				Sum:     elem[3],
			}
		}
	}
	return modInfo
}

// Copy of encoding/hex.DecodeString, Decode, fromHexChar
// to avoid dependency cycle.
func hexDecodeString(s string) []byte {
	src := []byte(s)
	n := hexDecode(src, src)
	return src[:n]
}

func hexDecode(src, dst []byte) int {
	var i int
	for i = 0; i < len(src)/2; i++ {
		a, ok := fromHexChar(src[i*2])
		if !ok {
			panic("invalid byte error")
		}
		b, ok := fromHexChar(src[i*2+1])
		if !ok {
			return i
		}
		dst[i] = (a << 4) | b
	}
	if len(src)%2 == 1 {
		panic("invalid length error")
	}
	return i
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
