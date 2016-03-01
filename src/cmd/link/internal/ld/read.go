// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ld

import (
	"bytes"
	"fmt"
	"io"
	"log"
)

// objReader reads an object file into memory.
type objReader struct {
	file    string
	data    []byte
	off     int
	scratch []byte
}

// readLine reads up until '\n', returning any read bytes including '\n'.
// If no '\n' is encountered, an error is returned.
func (f *objReader) readLine() (string, error) {
	i := bytes.IndexByte(f.data[f.off:], '\n')
	if i == -1 {
		return "", io.EOF
	}
	i++
	s := string(f.data[f.off : f.off+i])
	f.off += i
	return s, nil
}

// readByte reads a byte.
func (f *objReader) readByte() byte {
	b := f.data[f.off]
	f.off++
	return b
}

// seek sets the input offset to off if it is within range, or returns false.
func (f *objReader) seek(off int) bool {
	if off > len(f.data) {
		return false
	}
	f.off = off
	return true
}

// readBytes reads the next n bytes from an Input.
// The return slice is a view onto the original data. Take care.
//
// If the read is short, io.EOF is returned.
func (f *objReader) readBytes(n int) (b []byte, err error) {
	if n < 0 {
		return nil, fmt.Errorf("readBytes: invalid n: %d (f.off=%d, len f.data=%d)", n, f.off, len(f.data))
	}
	if f.off >= len(f.data) {
		return nil, io.EOF
	}
	end := f.off + n
	if end > len(f.data) {
		end = len(f.data)
		err = io.EOF
	}
	b = f.data[f.off:end]
	//b = append(b, f.data[f.off:end]...)
	f.off = end
	return b, nil
}

// Read implements io.Reader.
func (f *objReader) Read(p []byte) (n int, err error) {
	n = copy(p, f.data[f.off:])
	f.off += n
	if n < len(p) {
		return n, io.EOF
	}
	return n, nil
}

// readInt64 reads a varint.
func (f *objReader) readInt64() int64 {
	n := uint64(0)
	for shift := 0; ; shift += 7 {
		if shift >= 64 {
			log.Fatalf("%s: varint at %d is corrupt", f.file, f.off)
		}
		c := f.readByte()
		n |= uint64(c&0x7F) << uint(shift)
		if c&0x80 == 0 {
			break
		}
	}
	return int64(n>>1) ^ (int64(n<<63) >> 63)
}

// readInt reads a varint ensures its value fits in an int.
// To keep the linker portable, this means the integer is limited to 4 bytes.
func (f *objReader) readInt() int {
	n := f.readInt64()
	if int64(int(n)) != n {
		log.Panicf("%v out of range for int", n)
	}
	return int(n)
}

// readUint8 reads a varint ensures its value fits in a uint8.
func (f *objReader) readUint8() uint8 {
	n := f.readInt64()
	if int64(uint8(n)) != n {
		log.Panicf("%v out of range for uint8", n)
	}
	return uint8(n)
}

// readInt8 reads a varint ensures its value fits in a int8.
func (f *objReader) readInt8() int8 {
	n := f.readInt64()
	if int64(int8(n)) != n {
		log.Panicf("%v out of range for int8", n)
	}
	return int8(n)
}

// readInt16 reads a varint ensures its value fits in a int16.
func (f *objReader) readInt16() int16 {
	n := f.readInt64()
	if int64(int16(n)) != n {
		log.Panicf("%v out of range for int16", n)
	}
	return int16(n)
}

// readInt32 reads a varint ensures its value fits in a int32.
func (f *objReader) readInt32() int32 {
	n := f.readInt64()
	if int64(int32(n)) != n {
		log.Panicf("%v out of range for int32", n)
	}
	return int32(n)
}

// readString reads a varint N and returns the N following bytes.
func (f *objReader) readString() string {
	return string(f.readData())
}

var emptyPkg = []byte(`"".`)

// readSymName reads a string, replacing any "". with pkg.
func (f *objReader) readSymName(pkg string) string {
	origName := f.readData()
	if len(origName) == 0 {
		f.readInt64()
		return ""
	}

	if f.scratch == nil {
		f.scratch = make([]byte, 0, 32)
	}
	buf := f.scratch[:0]
	for {
		i := bytes.Index(origName, emptyPkg)
		if i == -1 {
			buf = append(buf, origName...)
			break
		}
		buf = append(buf, origName[:i]...)
		buf = append(buf, pkg...)
		buf = append(buf, '.')
		origName = origName[i+len(emptyPkg):]
	}
	name := string(buf)
	f.scratch = buf[:0]
	return name
}

// readData reads a varint N and returns the N following bytes.
// The slice returned is a view onto the original data. Take care.
func (f *objReader) readData() []byte {
	n := f.readInt()
	b := f.data[f.off : f.off+n]
	f.off += n
	return b
}
