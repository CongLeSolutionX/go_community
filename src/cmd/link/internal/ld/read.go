// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ld

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
)

// objReader reads an object file.
type objReader struct {
	objDataReader

	file    string
	scratch []byte
}

// newObjReader loads an *objReader from a source file.
//
// If the file fits in a []byte and the OS supports it, it is mmaped
// entirely into memory. Otherwise, it is read as a stream.
//
// Note that if files are mmaped, they are never unmmaped. Slices onto
// files are stored in *LSym P fields, which live for the life of the
// linker process and so there is no earlier point to unmap.
func newObjReader(file string) (*objReader, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	size := fi.Size()
	if size == 0 {
		f.Close()
		return &objReader{}, nil
	}

	var data []byte
	if int64(int(size)) == size && Debug['m'] == 0 {
		data = mmap(f, size)
	}
	if data == nil {
		// mmap not supported (or disabled)
		return &objReader{
			file: file,
			objDataReader: &objFile{
				file: file,
				f:    f,
				r:    bufio.NewReaderSize(f, 1<<18),
			},
		}, nil
	}

	f.Close()
	return &objReader{
		file: file,
		objDataReader: &objMmap{
			file: file,
			data: data,
		},
	}, nil
}

// readInt64 reads a varint.
func (f *objReader) readInt64() int64 {
	n := uint64(0)
	for shift := 0; ; shift += 7 {
		if shift >= 64 {
			log.Fatalf("%s: varint at %d is corrupt", f.file, f.offset())
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

// readData reads a varint N and returns the N following bytes.
// The slice returned may be a view onto the original data. Take care.
func (f *objReader) readData() []byte {
	n := f.readInt64()
	b, err := f.readBytes(n)
	if err != nil {
		panic(fmt.Sprintf("ld.objReader: cannot read %d bytes: %v", n, err))
	}
	return b
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

// objDataReader is the backing store reader for an object file.
type objDataReader interface {
	io.Reader

	close()

	// readLine reads up until '\n', returning any read bytes including '\n'.
	// If no '\n' is encountered, an error is returned.
	readLine() (string, error)

	// readByte reads a byte.
	readByte() byte

	// seek sets the input offset to off if it is within range, or returns false.
	seek(off int64) bool

	// readBytes reads the next n bytes from an Input.
	// The return slice may be a view onto the original data. Take care.
	//
	// If the read is short, io.EOF is returned.
	readBytes(n int64) (b []byte, err error)

	offset() int64

	size() int64

	discard(n int) error

	// peek returns the next n bytes without advancing the reader.
	// The slice returned may be a view onto the original data. Take care.
	peek(n int) []byte
}

// objMmap is an objDataReader that loads a file with mmap.
type objMmap struct {
	file string
	data []byte
	off  int
}

func (f *objMmap) close() {
	f.data = nil
	f.off = 0
}

func (f *objMmap) readLine() (string, error) {
	i := bytes.IndexByte(f.data[f.off:], '\n')
	if i == -1 {
		return "", io.EOF
	}
	i++
	s := string(f.data[f.off : f.off+i])
	f.off += i
	return s, nil
}

func (f *objMmap) readByte() byte {
	b := f.data[f.off]
	f.off++
	return b
}

func (f *objMmap) seek(off int64) bool {
	if int64(int(off)) != off {
		panic(fmt.Sprintf("seeking too far: %d", off))
	}
	if int(off) > len(f.data) {
		return false
	}
	f.off = int(off)
	return true
}

func (f *objMmap) readBytes(n int64) (b []byte, err error) {
	if int64(int(n)) != n {
		return nil, fmt.Errorf("readBytes: n for %s too big: %d", f.file, n)
	}
	if n < 0 {
		return nil, fmt.Errorf("readBytes: invalid n: %d (f.off=%d, len f.data=%d)", n, f.off, len(f.data))
	}
	if f.off >= len(f.data) {
		return nil, io.EOF
	}
	end := f.off + int(n)
	if end > len(f.data) {
		end = len(f.data)
		err = io.EOF
	}
	b = f.data[f.off:end:end]
	f.off = end
	return b, nil
}

func (f *objMmap) offset() int64 { return int64(f.off) }
func (f *objMmap) size() int64   { return int64(len(f.data)) }

func (f *objMmap) discard(n int) error {
	f.off++
	if f.off > len(f.data) {
		return io.ErrUnexpectedEOF
	}
	return nil
}

func (f *objMmap) Read(p []byte) (n int, err error) {
	n = copy(p, f.data[f.off:])
	f.off += n
	if n < len(p) {
		return n, io.EOF
	}
	return n, nil
}

func (f *objMmap) peek(n int) []byte {
	return f.data[f.off : f.off+n]
}
