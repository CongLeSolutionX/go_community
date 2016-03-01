// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ld

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

// objFile is an objReader that steams from an *os.File.
type objFile struct {
	file    string
	scratch []byte
	f       *os.File
	r       *bufio.Reader
}

func (f *objFile) close() {
	if f.f != nil {
		f.f.Close()
	}
	f.f = nil
	f.r = nil
}

func (f *objFile) readLine() (string, error) {
	return f.r.ReadString('\n')
}

func (f *objFile) readByte() byte {
	b, err := f.r.ReadByte()
	if err != nil {
		panic(err)
	}
	return b
}

func (f *objFile) seek(off int64) bool {
	_, err := f.f.Seek(off, 0)
	f.r.Reset(f.f)
	if err != nil {
		return false
	}
	return true
}

func (f *objFile) readBytes(n int64) (b []byte, err error) {
	b = make([]byte, n)
	n2, err := io.ReadFull(f.r, b)
	return b[:n2], err
}

func (f *objFile) offset() int64 {
	off, err := f.f.Seek(0, 1)
	if err != nil {
		panic(fmt.Sprintf("ld.objFile: %v", err))
	}
	off -= int64(f.r.Buffered())
	return off
}

func (f *objFile) size() int64 {
	off := f.offset() + int64(f.r.Buffered())
	size, err := f.f.Seek(0, 2)
	if err != nil {
		panic(fmt.Sprintf("ld.objFile: offset seek 1: %v", err))
	}
	if _, err := f.f.Seek(off, 0); err != nil {
		panic(fmt.Sprintf("ld.objFile: offset seek 2: %v", err))
	}
	return size
}

func (f *objFile) discard(n int) error {
	// Cannot use Discard method, it came after Go 1.4.
	for n > 0 {
		if _, err := f.r.ReadByte(); err != nil {
			return err
		}
		n--
	}
	return nil
}

func (f *objFile) Read(p []byte) (n int, err error) {
	return f.r.Read(p)
}

func (f *objFile) peek(n int) []byte {
	b, err := f.r.Peek(n)
	if err != nil {
		panic(fmt.Sprintf("ld.objFile: %s: %v", f.file, err))
	}
	return b
}
