// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package macho provides functionalities to handle Mach-O
// beyond the debug/macho package, for the toolchain.
package macho

import (
	"debug/macho"
	"encoding/binary"
	"io"
	"unsafe"
)

// LoadCmd is macho.LoadCmd with its length, which is also
// the load command header in the Mach-O file.
type LoadCmd struct {
	Cmd macho.LoadCmd
	Len uint32
}

type LoadCmdReader struct {
	offset, next int64
	f            io.ReadSeeker
	order        binary.ByteOrder
}

func NewLoadCmdReader(f io.ReadSeeker, order binary.ByteOrder, nextOffset int64) LoadCmdReader {
	return LoadCmdReader{next: nextOffset, f: f, order: order}
}

func (r *LoadCmdReader) Next() (LoadCmd, error) {
	var cmd LoadCmd

	r.offset = r.next
	if _, err := r.f.Seek(r.offset, 0); err != nil {
		return cmd, err
	}
	if err := binary.Read(r.f, r.order, &cmd); err != nil {
		return cmd, err
	}
	r.next = r.offset + int64(cmd.Len)
	return cmd, nil
}

func (r LoadCmdReader) ReadAt(offset int64, data interface{}) error {
	if _, err := r.f.Seek(r.offset+offset, 0); err != nil {
		return err
	}
	return binary.Read(r.f, r.order, data)
}

func (r LoadCmdReader) WriteAt(offset int64, data interface{}) error {
	if _, err := r.f.Seek(r.offset+offset, 0); err != nil {
		return err
	}
	return binary.Write(r.f.(io.Writer), r.order, data)
}

func (r LoadCmdReader) Offset() int64 { return r.offset }

func FileHeaderSize(f *macho.File) int64 {
	offset := int64(unsafe.Sizeof(f.FileHeader))
	if is64bit := f.Magic == macho.Magic64; is64bit {
		// mach_header_64 has one extra uint32.
		offset += 4
	}
	return offset
}
