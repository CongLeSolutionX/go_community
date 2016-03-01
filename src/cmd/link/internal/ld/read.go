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
//
// Either the entire file is loaded into memory as data (including
// possibly being mmaped), or the file is read as a stream via f/r.
type objReader struct {
	file    string
	scratch []byte

	data []byte // complete file, if nil objReader uses f/r
	off  int    // offset if data is non-nil

	f *os.File
	r *bufio.Reader // reader for f if data is nil
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

	input := &objReader{
		file: file,
	}

	if int64(int(size)) == size && Debug['m'] == 0 {
		// File fits in a slice.
		input.data = mmap(f, size)
	}
	if input.data == nil {
		// mmap not supported (or disabled)
		input.f = f
		input.r = bufio.NewReaderSize(f, 1<<18)
		input.off = -1
		return input, nil
	}

	f.Close()
	return input, nil
}

func (f *objReader) close() {
	if f.f != nil {
		f.f.Close()
	}
	f.data = nil
	f.off = 0
	f.f = nil
	f.r = nil
}

// readLine reads up until '\n', returning any read bytes including '\n'.
// If no '\n' is encountered, an error is returned.
func (f *objReader) readLine() (string, error) {
	if f.r != nil {
		return f.r.ReadString('\n')
	}

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
	if f.r != nil {
		b, err := f.r.ReadByte()
		if err != nil {
			panic(err)
		}
		return b
	}
	b := f.data[f.off]
	f.off++
	return b
}

// seek sets the input offset to off if it is within range, or returns false.
func (f *objReader) seek(off int64) bool {
	if f.r != nil {
		_, err := f.f.Seek(off, 0)
		f.r.Reset(f.f)
		if err != nil {
			return false
		}
		return true
	}
	if int64(int(off)) != off {
		panic(fmt.Sprintf("seeking too far: %d", off))
	}
	if int(off) > len(f.data) {
		return false
	}
	f.off = int(off)
	return true
}

// readBytes reads the next n bytes from an Input.
// The return slice may be a view onto the original data. Take care.
//
// If the read is short, io.EOF is returned.
func (f *objReader) readBytes(n int) (b []byte, err error) {
	if f.r != nil {
		b = make([]byte, n)
		_, err = io.ReadFull(f.r, b)
		return b, err
	}
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
	b = f.data[f.off:end:end]
	f.off = end
	return b, nil
}

func (f *objReader) offset() int64 {
	if f.r != nil {
		off, err := f.f.Seek(0, 1)
		if err != nil {
			panic(fmt.Sprintf("ld.objReader: %v", err))
		}
		off -= int64(f.r.Buffered())
		return off
	}
	return int64(f.off)
}

func (f *objReader) size() int64 {
	if f.r != nil {
		off := f.offset()
		size, err := f.f.Seek(0, 2)
		if err != nil {
			panic(fmt.Sprintf("ld.objReader: %v", err))
		}
		f.seek(off)
		return size
	}
	return int64(len(f.data))
}

func (f *objReader) discard(n int) error {
	if f.r != nil {
		// Cannot use Discard method, it came after Go 1.4.
		for n > 0 {
			if _, err := f.r.ReadByte(); err != nil {
				return err
			}
			n--
		}
		return nil
	}
	f.off++
	if f.off > len(f.data) {
		return io.ErrUnexpectedEOF
	}
	return nil
}

// Read implements io.Reader.
func (f *objReader) Read(p []byte) (n int, err error) {
	if f.r != nil {
		return f.r.Read(p)
	}
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
// The slice returned may be a view onto the original data. Take care.
func (f *objReader) readData() []byte {
	n := f.readInt()
	if f.r != nil {
		b := make([]byte, n)
		if _, err := io.ReadFull(f.r, b); err != nil {
			panic(fmt.Sprintf("ld.objReader: %s: %v", f.file, err))
		}
		return b
	}
	b := f.data[f.off : f.off+n : f.off+n]
	f.off += n
	return b
}

// peek returns the next n bytes without advancing the reader.
// The slice returned may be a view onto the original data. Take care.
func (f *objReader) peek(n int) []byte {
	if f.r != nil {
		b, err := f.r.Peek(n)
		if err != nil {
			panic(fmt.Sprintf("ld.objReader: %s: %v", f.file, err))
		}
		return b
	}
	return f.data[f.off : f.off+n]
}
