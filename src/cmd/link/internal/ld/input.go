// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ld

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"syscall"
)

// Input is an input data file.
type Input struct {
	file string
	data []byte
	off  int
}

// LoadInput loads an *Input from a source file.
//
// Note that if files are mmaped, they are never unmaped. Slices onto
// files are stored in *LSym P fields, which live for the life of the
// linker process and so there is no earlier to point to unmap.
func LoadInput(file string) (*Input, error) {
	/* TODO: on OS X, ReadFile is approximately the same speed as tip,
	and mmap is slightly slower.
	TODO: test linux

	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	*/
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	size := fi.Size()
	if size == 0 {
		return &Input{}, nil
	}
	if size < 0 {
		return nil, fmt.Errorf("ld: file %q has negative size", file)
	}
	if size != int64(int(size)) {
		return nil, fmt.Errorf("ld: file %q is too big", file)
	}

	data, err := syscall.Mmap(int(f.Fd()), 0, int(size), syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		return nil, fmt.Errorf("ld: mmap failed: %v", err)
	}

	input := &Input{
		file: file,
		data: data,
	}
	return input, nil
}

// ReadLine reads up until '\n', returning any read bytes including '\n'.
// If no '\n' is encountered, an error is returned.
func (f *Input) ReadLine() (string, error) {
	i := bytes.IndexByte(f.data[f.off:], '\n')
	if i == -1 {
		return "", io.EOF
	}
	i++
	s := string(f.data[f.off : f.off+i])
	f.off += i
	return s, nil
}

// ReadByte reads a byte.
func (f *Input) ReadByte() byte {
	b := f.data[f.off]
	f.off++
	return b
}

// Seek sets the input offset to off if it is within range, or returns false.
func (f *Input) Seek(off int) bool {
	if off > len(f.data) {
		return false
	}
	f.off = off
	return true
}

// ReadBytes reads the next n bytes from an Input.
// The return slice is a view onto the original data. Take care.
//
// If the read is short, io.EOF is returned.
func (f *Input) ReadBytes(n int) (b []byte, err error) {
	if n < 0 {
		return nil, fmt.Errorf("ReadBytes: invalid n: %d (f.off=%d, len f.data=%d)", n, f.off, len(f.data))
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
func (f *Input) Read(p []byte) (n int, err error) {
	n = copy(p, f.data[f.off:])
	if n < len(p) {
		return n, io.EOF
	}
	return n, nil
}

// ReadInt64 reads an varint.
func (f *Input) ReadInt64() int64 {
	n := uint64(0)
	for shift := 0; ; shift += 7 {
		if shift >= 64 {
			log.Fatalf("%s: varint at %d is corrupt", f.file, f.off)
		}
		c := f.ReadByte()
		n |= uint64(c&0x7F) << uint(shift)
		if c&0x80 == 0 {
			break
		}
	}
	return int64(n>>1) ^ (int64(n<<63) >> 63)
}

// ReadInt reads a varint ensures its value fits in an int.
// To keep the linker portable, this means the integer is limited to 4 bytes.
func (f *Input) ReadInt() int {
	n := f.ReadInt64()
	if int64(int(n)) != n {
		log.Panicf("%v out of range for int", n)
	}
	return int(n)
}

// ReadUint8 reads a varint ensures its value fits in a uint8.
func (f *Input) ReadUint8() uint8 {
	n := f.ReadInt64()
	if int64(uint8(n)) != n {
		log.Panicf("%v out of range for uint8", n)
	}
	return uint8(n)
}

// ReadInt8 reads a varint ensures its value fits in a int8.
func (f *Input) ReadInt8() int8 {
	n := f.ReadInt64()
	if int64(int8(n)) != n {
		log.Panicf("%v out of range for int8", n)
	}
	return int8(n)
}

// ReadInt32 reads a varint ensures its value fits in a int32.
func (f *Input) ReadInt32() int32 {
	n := f.ReadInt64()
	if int64(int32(n)) != n {
		log.Panicf("%v out of range for int32", n)
	}
	return int32(n)
}

// ReadString reads a varint N and returns the N following bytes.
func (f *Input) ReadString() string {
	return string(f.ReadData())
}

// ReadData reads a varint N and returns the N following bytes.
// The slice returned is a view onto the original data. Take care.
func (f *Input) ReadData() []byte {
	n := f.ReadInt()
	b := f.data[f.off : f.off+n]
	f.off += n
	return b
}
