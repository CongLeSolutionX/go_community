// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package decodecounter

import (
	"encoding/binary"
	"fmt"
	"internal/coverage"
	"internal/coverage/slicereader"
	"internal/coverage/stringtab"
	"io"
	"os"
)

// This file contains helpers for reading counter data files emitted
// during the executions of a coverage-instrumented binary.

type CounterDataReader struct {
	f     *os.File
	stab  *stringtab.StringTableReader
	args  []string
	mr    io.Reader
	hdr   coverage.CounterFileHeader
	u32b  []byte
	u8b   []byte
	debug bool
}

func NewCounterDataReader(fn string, f *os.File, r io.Reader) (*CounterDataReader, error) {
	cdr := &CounterDataReader{
		f:    f,
		mr:   r,
		u32b: make([]byte, 4),
		u8b:  make([]byte, 1),
	}
	// Read header
	if err := binary.Read(r, binary.LittleEndian, &cdr.hdr); err != nil {
		return nil, err
	}
	if cdr.debug {
		fmt.Fprintf(os.Stderr, "=-= counter file header: %+v\n", cdr.hdr)
	}
	// Read string table
	if err := cdr.readStringTable(); err != nil {
		return nil, err
	}
	if err := cdr.readArgs(); err != nil {
		return nil, err
	}
	return cdr, nil
}

func (cdr *CounterDataReader) readStringTable() error {
	b := make([]byte, cdr.hdr.StrTabLen)
	nr, err := cdr.mr.Read(b)
	if err != nil {
		return err
	}
	if nr != int(cdr.hdr.StrTabLen) {
		return fmt.Errorf("error: short read on string table")
	}
	slr := slicereader.NewReader(b, false /* not readonly */)
	cdr.stab = stringtab.NewReader(slr)
	cdr.stab.Read(int(cdr.hdr.StrTabNentries))
	return nil
}

func (cdr *CounterDataReader) readArgs() error {
	b := make([]byte, cdr.hdr.ArgsLen)
	nr, err := cdr.mr.Read(b)
	if err != nil {
		return err
	}
	if nr != int(cdr.hdr.ArgsLen) {
		return fmt.Errorf("error: short read on args table")
	}
	slr := slicereader.NewReader(b, false /* not readonly */)
	cdr.args = make([]string, 0, cdr.hdr.ArgsNentries)
	for i := uint32(0); i < cdr.hdr.ArgsNentries; i++ {
		idx := slr.ReadULEB128()
		cdr.args = append(cdr.args, cdr.stab.Get(uint32(idx)))
	}
	return nil
}

// Args returns the program arguments (saved from os.Args during
// the run of the instrumented binary) read from the counter
// data file.
func (cdr *CounterDataReader) Args() []string {
	return cdr.args
}

// FuncPayload encapsulates the counter data payload for a single
// function as read from a counter data file.
type FuncPayload struct {
	PkgIdx   uint32
	FuncIdx  uint32
	Counters []uint32
}

func (cdr *CounterDataReader) NextFunc(p *FuncPayload) error {
	var rdu32 func() (uint32, error)
	if cdr.hdr.CFlavor == coverage.CtrULeb128 {
		rdu32 = func() (uint32, error) {
			var shift uint
			var value uint64
			for {
				_, err := cdr.mr.Read(cdr.u8b)
				if err != nil {
					return 0, err
				}
				b := cdr.u8b[0]
				value |= (uint64(b&0x7F) << shift)
				if b&0x80 == 0 {
					break
				}
				shift += 7
			}
			return uint32(value), nil
		}
	} else if cdr.hdr.CFlavor == coverage.CtrRaw {
		if cdr.hdr.BigEndian {
			rdu32 = func() (uint32, error) {
				n, err := cdr.mr.Read(cdr.u32b)
				if err != nil {
					return 0, err
				}
				if n != 4 {
					return 0, io.EOF
				}
				return binary.BigEndian.Uint32(cdr.u32b), nil
			}
		} else {
			rdu32 = func() (uint32, error) {
				n, err := cdr.mr.Read(cdr.u32b)
				if err != nil {
					return 0, err
				}
				if n != 4 {
					return 0, io.EOF
				}
				return binary.LittleEndian.Uint32(cdr.u32b), nil
			}
		}
	} else {
		panic("internal error: unknown counter flavor")
	}

	// Read number of counters. For raw counter files, if zero, assume
	// that this is a dead counter region, and move ahead until we hit
	// something non-zero (or EOF).
	var nc uint32
	var err error
	for {
		nc, err = rdu32()
		if err == io.EOF {
			return io.EOF
		}
		if nc != 0 {
			break
		}
	}

	// Read package and func indices.
	p.PkgIdx, err = rdu32()
	if err != nil {
		return err
	}
	p.FuncIdx, err = rdu32()
	if err != nil {
		return err
	}
	if cap(p.Counters) < 1024 {
		p.Counters = make([]uint32, 0, 1024)
	}
	p.Counters = p.Counters[:0]
	for i := uint32(0); i < nc; i++ {
		v, err := rdu32()
		if err != nil {
			return err
		}
		p.Counters = append(p.Counters, v)
	}
	return nil
}
