// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package locklog provides a reader for the runtime lock operation log.
package locklog

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
)

type LogOp byte

// Log operation codes. These must be kept in sync with src/runtime/locklog.go.
const (
	LogOpAcquire LogOp = iota
	LogOpRelease
	lockLogOpReleaseLast
	LogOpMayAcquire

	// These are hidden from the caller.
	lockLogOpNewClass
	lockLogOpLabel

	lockLogOpBits = 3 // Number of bits to store lockLogOp*
)

type LogReader struct {
	r       *bufio.Reader
	block   []byte
	pos     int
	perM    map[int]*logPerM
	mid     int
	curM    *logPerM
	classes []string
}

type Record struct {
	Op        LogOp
	M         int
	LockAddr  uint64
	Stack     []uint64
	LockClass string
	LockRank  uint64
}

type logPerM struct {
	id        int
	last      uint64
	lockClass string
	lockRank  uint64
}

// NewLogReader prepares to read from the runtime lock log on r.
//
// It first performs a handshake on r, and calls load with information
// from the handshake. When load returns, it completes the handshake,
// which lets load read information from the binary even if it's a
// temporary file.
func NewLogReader(r io.ReadWriter, load func(exePath string, runtimeMain uint64) error) (*LogReader, error) {
	// Handshake.
	hdr := make([]byte, 8*2+4)
	if _, err := io.ReadFull(r, hdr); err != nil {
		return nil, fmt.Errorf("log handshake failed: %w", err)
	}
	magic := string(hdr[:8])
	if magic != "locklog\x00" {
		return nil, fmt.Errorf("log handshake failed: bad magic number %q", magic)
	}
	runtimeMain := binary.LittleEndian.Uint64(hdr[8:])
	exeLen := binary.LittleEndian.Uint32(hdr[8*2:])
	exePath := make([]byte, exeLen)
	if _, err := io.ReadFull(r, exePath); err != nil {
		return nil, fmt.Errorf("log handshake failed: %w", err)
	}

	// Process the handshake.
	if err := load(string(exePath), runtimeMain); err != nil {
		return nil, err
	}

	// Acknowledge.
	if _, err := r.Write([]byte{0}); err != nil {
		return nil, fmt.Errorf("log handshake failed: %w", err)
	}

	// Buffer the input.
	br := bufio.NewReader(r)
	return &LogReader{r: br, perM: make(map[int]*logPerM), classes: []string{""}}, nil
}

// Next populates rec with the next record in the log. It return
// io.EOF if there are no more records, or another error if I/O fails
// or a record is truncated or corrupted.
func (r *LogReader) Next(rec *Record) error {
	const debugLog = false

nextRecord:
	if r.pos == len(r.block) {
		// Read the next block.
		hdr := make([]byte, 8)
		if _, err := io.ReadFull(r.r, hdr); err != nil {
			return err
		}

		bytes := int(binary.LittleEndian.Uint32(hdr[0:]) - 8)
		mid := binary.LittleEndian.Uint32(hdr[4:])

		r.mid = int(mid)
		if cap(r.block) < bytes {
			r.block = make([]byte, bytes)
		} else {
			r.block = r.block[:bytes]
		}
		r.pos = 0

		if r.perM[r.mid] == nil {
			r.perM[r.mid] = new(logPerM)
		}
		r.curM = r.perM[r.mid]

		if _, err := io.ReadFull(r.r, r.block); err != nil {
			if err == io.EOF {
				err = io.ErrUnexpectedEOF
			}
			return err
		}
	}

	// Read the op
	op := r.block[r.pos]
	rec.Op = LogOp(op & ((1 << lockLogOpBits) - 1))
	r.pos++
	arg := int(op >> lockLogOpBits)
	rec.M = r.mid

	switch rec.Op {
	default:
		return fmt.Errorf("bad log op %d", rec.Op)

	case LogOpAcquire:
		rec.LockAddr = r.readUint64()
		r.curM.last = rec.LockAddr
		rec.Stack = rec.Stack[:0]
		for i := 0; i < arg; i++ {
			rec.Stack = append(rec.Stack, r.readUint64())
		}
		rec.LockClass = r.curM.lockClass
		rec.LockRank = r.curM.lockRank
		r.curM.lockClass, r.curM.lockRank = "", 0

		if debugLog {
			log.Printf("[%d] acquire %#x %s %d", r.mid, rec.LockAddr, rec.LockClass, rec.LockRank)
		}

	case LogOpRelease:
		rec.LockAddr = r.readUint64()

		if debugLog {
			log.Printf("[%d] release %#x", r.mid, rec.LockAddr)
		}

	case LogOpMayAcquire:
		rec.LockAddr = r.readUint64()
		rec.Stack = rec.Stack[:0]
		for i := 0; i < arg; i++ {
			rec.Stack = append(rec.Stack, r.readUint64())
		}

		if debugLog {
			log.Printf("[%d] mayAcquire %#x", r.mid, rec.LockAddr)
		}

	case lockLogOpReleaseLast:
		rec.Op = LogOpRelease
		rec.LockAddr = r.curM.last

		if debugLog {
			log.Printf("[%d] releaseLast %#x", r.mid, rec.LockAddr)
		}

	case lockLogOpNewClass:
		nameLen := r.readUint32()
		label := string(r.block[r.pos : r.pos+int(nameLen)])
		r.pos += int(nameLen)
		r.classes = append(r.classes, label)
		if debugLog {
			log.Printf("newClass %d %s", len(r.classes)-1, label)
		}
		// These records are invisible to the reader.
		goto nextRecord

	case lockLogOpLabel:
		r.curM.lockClass = r.classes[r.readUint32()]
		r.curM.lockRank = r.readUint64()
		if debugLog {
			log.Printf("[%d] label %s %d", r.mid, r.curM.lockClass, r.curM.lockRank)
		}
		goto nextRecord
	}

	return nil
}

func (r *LogReader) readUint32() uint32 {
	v := binary.LittleEndian.Uint32(r.block[r.pos:])
	r.pos += 4
	return v
}

func (r *LogReader) readUint64() uint64 {
	v := binary.LittleEndian.Uint64(r.block[r.pos:])
	r.pos += 8
	return v
}
