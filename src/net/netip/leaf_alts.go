// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Stuff that exists in std, but we can't use due to being a dependency
// of net, for go/build deps_test policy reasons.

package netip

import (
	"errors"
	"strconv"
)

func errorf(format string, arg ...interface{}) error {
	return errors.New(sprintf(format, arg...))
}

func sprintf(format string, args ...interface{}) string {
	var out []byte
	inPercent := false
	for i := 0; i < len(format); i++ {
		c := format[i]
		if !inPercent {
			if c == '%' {
				inPercent = true
			} else {
				out = append(out, c)
			}
			continue
		}
		if len(args) == 0 {
			panic("missing argument")
		}
		arg := args[0]
		inPercent = false
		if c == 'v' {
			switch arg.(type) {
			case int, uint8, uint16:
				c = 'd'
			case string:
				c = 's'
			case error:
				arg = arg.(error).Error()
				c = 's'
			}
		}
		switch c {
		case '%':
			out = append(out, '%')
			continue
		case 'd':
			switch arg := arg.(type) {
			case int:
				out = strconv.AppendInt(out, int64(arg), 10)
			case uint8:
				out = strconv.AppendUint(out, uint64(arg), 10)
			case uint16:
				out = strconv.AppendUint(out, uint64(arg), 10)
			default:
				panic("unhandled %d type")
			}
		case 'q':
			switch arg := arg.(type) {
			case string:
				out = strconv.AppendQuote(out, arg)
			default:
				panic("unhandled %q type")
			}
		case 's':
			switch arg := arg.(type) {
			case string:
				out = append(out, arg...)
			default:
				panic("unhandled %s type")
			}
		default:
			panic("unsupported fmt-ish pattern")
		}
	}
	if inPercent {
		panic("trailing %")
	}
	if len(args) != 0 {
		panic("extra arguments")
	}
	return string(out)
}

func stringsIndexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}

func stringsLastIndexByte(s string, b byte) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == b {
			return i
		}
	}
	return -1
}

func beUint64(b []byte) uint64 {
	_ = b[7] // bounds check hint to compiler; see golang.org/issue/14808
	return uint64(b[7]) | uint64(b[6])<<8 | uint64(b[5])<<16 | uint64(b[4])<<24 |
		uint64(b[3])<<32 | uint64(b[2])<<40 | uint64(b[1])<<48 | uint64(b[0])<<56
}

func bePutUint64(b []byte, v uint64) {
	_ = b[7] // early bounds check to guarantee safety of writes below
	b[0] = byte(v >> 56)
	b[1] = byte(v >> 48)
	b[2] = byte(v >> 40)
	b[3] = byte(v >> 32)
	b[4] = byte(v >> 24)
	b[5] = byte(v >> 16)
	b[6] = byte(v >> 8)
	b[7] = byte(v)
}

func bePutUint32(b []byte, v uint32) {
	_ = b[3] // early bounds check to guarantee safety of writes below
	b[0] = byte(v >> 24)
	b[1] = byte(v >> 16)
	b[2] = byte(v >> 8)
	b[3] = byte(v)
}
