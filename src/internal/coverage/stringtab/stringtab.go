// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package stringtab

import (
	"fmt"
	"io"
)

// This package implements a string table writer utility, for
// use in emitting coverage meta-data and counter-data files.

type StringTableWriter struct {
	stab  map[string]uint32
	strs  []string
	tmp   []byte
	fname string
}

func AppendUleb128(b []byte, v uint) []byte {
	for {
		c := uint8(v & 0x7f)
		v >>= 7
		if v != 0 {
			c |= 0x80
		}
		b = append(b, c)
		if c&0x80 == 0 {
			break
		}
	}
	return b
}

func (stw *StringTableWriter) InitStringTableWriter(fname string) {
	stw.stab = make(map[string]uint32)
	stw.tmp = make([]byte, 64)
	stw.fname = fname
}

func (stw *StringTableWriter) StringTableNentries() uint32 {
	return uint32(len(stw.strs))
}

func (stw *StringTableWriter) LookupString(s string) uint32 {
	if idx, ok := stw.stab[s]; ok {
		return idx
	}
	idx := uint32(len(stw.strs))
	stw.stab[s] = idx
	stw.strs = append(stw.strs, s)
	return idx
}

func (stw *StringTableWriter) StringTableSize() uint32 {
	rval := uint32(0)
	for _, s := range stw.strs {
		stw.tmp = stw.tmp[:0]
		slen := uint(len(s))
		stw.tmp = AppendUleb128(stw.tmp, slen)
		rval += uint32(len(stw.tmp)) + uint32(slen)
	}
	return rval
}

func (stw *StringTableWriter) WriteStringTable(w io.Writer) error {
	var err error
	var nw int
	for _, s := range stw.strs {
		stw.tmp = stw.tmp[:0]
		stw.tmp = AppendUleb128(stw.tmp, uint(len(s)))
		if nw, err = w.Write(stw.tmp); err != nil {
			return fmt.Errorf("error writing %s: %v\n", stw.fname, err)
		}
		if nw == 0 {
			return fmt.Errorf("error writing %s: short write uleb len\n", stw.fname)
		}
		if nw, err = w.Write([]byte(s)); err != nil {
			return fmt.Errorf("error writing %s: %v\n", stw.fname, err)
		}
		if nw != len(s) {
			return fmt.Errorf("error writing %s: short write %d string\n", stw.fname, nw)
		}
	}
	return nil
}
