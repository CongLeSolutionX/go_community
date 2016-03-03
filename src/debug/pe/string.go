// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pe

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// cstring converts ascii byte sequence b to string. It stops once it finds 0.
func cstring(b []byte) string {
	var i int
	for i = 0; i < len(b) && b[i] != 0; i++ {
	}
	return string(b[0:i])
}

// StringTable provides access to COFF string table.
type StringTable []byte

func newStringTable(fh *FileHeader, r io.ReadSeeker) (StringTable, error) {
	// COFF string table is located right after COFF symbol table.
	offset := fh.PointerToSymbolTable + COFFSymbolSize*fh.NumberOfSymbols
	_, err := r.Seek(int64(offset), os.SEEK_SET)
	if err != nil {
		return nil, fmt.Errorf("fail to seek to string table: %v", err)
	}
	var l uint32
	err = binary.Read(r, binary.LittleEndian, &l)
	if err != nil {
		return nil, fmt.Errorf("fail to read string table length: %v", err)
	}
	// string table length includes itself
	if l <= 4 {
		return nil, nil
	}
	l -= 4
	buf := make([]byte, l)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return nil, fmt.Errorf("fail to read string table: %v", err)
	}
	return StringTable(buf), nil
}

// String extracts string from COFF string table st at offset start.
func (st StringTable) String(start uint32) (string, error) {
	// start includes 4 bytes of string table length
	if start < 4 {
		return "", fmt.Errorf("offset %d is before the start of string table", start)
	}
	start -= 4
	if int(start) > len(st) {
		return "", fmt.Errorf("offset %d is beyond the end of string table", start)
	}
	return cstring(st[start:]), nil
}
