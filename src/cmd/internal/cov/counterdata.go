// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cov

import (
	"encoding/binary"
	"internal/coverage"
	"io"
	"os"
)

// This file contains helpers for reading counter data files emitted
// during the executions of a coverage-instrumented binary.

type CounterDataReader struct {
	f    *os.File
	mr   *mReader
	hdr  coverage.CounterFileHeader
	u32b []byte
}

func NewCounterDataReader(fn string) (*CounterDataReader, error) {
	f, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	mr, err2 := NewMreader(f)
	if err2 != nil {
		return nil, err2
	}
	r := &CounterDataReader{
		f:    f,
		mr:   mr,
		u32b: make([]byte, 4),
	}
	// Read header
	if err = binary.Read(r.mr, binary.LittleEndian, &r.hdr); err != nil {
		return nil, err
	}
	return r, nil
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
	u32b := make([]byte, 4)
	if cdr.hdr.BigEndian {
		rdu32 = func() (uint32, error) {
			n, err := cdr.mr.Read(u32b)
			if err != nil {
				return 0, err
			}
			if n != 4 {
				return 0, io.EOF
			}
			return binary.BigEndian.Uint32(u32b), nil
		}
	} else {
		rdu32 = func() (uint32, error) {
			n, err := cdr.mr.Read(u32b)
			if err != nil {
				return 0, err
			}
			if n != 4 {
				return 0, io.EOF
			}
			return binary.LittleEndian.Uint32(u32b), nil
		}
	}

	// Read number of counters. If zero, assume that this is a
	// dead counter region, and move ahead until we hit something
	// non-zero (or EOF).
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
	p.Counters = p.Counters[:0]
	if uint32(cap(p.Counters)) < nc {
		p.Counters = make([]uint32, 0, nc)
	}
	for i := uint32(0); i < nc; i++ {
		v, err := rdu32()
		if err != nil {
			return err
		}
		p.Counters = append(p.Counters, v)
	}
	return nil
}
