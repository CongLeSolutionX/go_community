// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package filebuf implements io.SeekReader for os files.
// This is useful only for very large files with lots of
// seeking. (otherwise use ioutil.ReadFile or bufio)
package filebuf

// PJW: this code is obscure! fix it

import (
	"fmt"
	"io"
	"log"
	"os"
)

// Buf is the implemented interface
type Buf interface {
	io.ReadCloser
	io.Seeker
	Size() int64
	Stats() Stat
}

// Buflen is the size of the buffer.
// The code is designed to never need to reread unnecessarily,
// he thinks. [check this PJW] (FromReader is not good.)
const Buflen = 1 << 20

// fbuf is a buffered file with seeking.
type fbuf struct {
	Name     string
	fd       *os.File
	size     int64
	bufloc   int64        // file loc of beginning of fixed
	bufpos   int32        // seekptr is at bufloc+bufpos
	fixed    [Buflen]byte // backing store for buf
	fixedlen int          // how much of fixed is valid file contents
	buf      []byte       // buf is fixed[0:fixedlen]
	// statistics
	seeks int   // number of calls to fd.Seek
	reads int   // number of calls to fd.Read
	bytes int64 // number of bytes read by fd.Read
}

// Stat returns the number of underlying seeks and reads, and bytes read
type Stat struct {
	Seeks int
	Reads int
	Bytes int64
}

// Stats returns the stats so far
func (fb *fbuf) Stats() Stat {
	return Stat{fb.seeks, fb.reads, fb.bytes}
}

// Size returns the file size
func (fb *fbuf) Size() int64 {
	return fb.size
}

// New returns an initialized *fbuf or an error
func New(fname string) (Buf, error) {
	fd, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	fi, err := fd.Stat()
	if err != nil || fi.Mode().IsDir() {
		return nil, fmt.Errorf("not readable: %s", fname)
	}
	return &fbuf{Name: fname, fd: fd, size: fi.Size()}, nil
}

// Read implements io.Reader. It may return a positive
// number of bytes read with io.EOF
func (fb *fbuf) Read(p []byte) (int, error) {
	if len(fb.buf[fb.bufpos:]) >= len(p) {
		copy(p, fb.buf[fb.bufpos:])
		fb.bufpos += int32(len(p))
		return len(p), nil
	}
	done := 0
	if len(fb.buf[fb.bufpos:]) > 0 {
		m := copy(p, fb.buf[fb.bufpos:])
		done = m
		fb.bufpos += int32(done)
	}
	// used up buffered data. logical seek pointer is at bufloc+bufpos.
	// if this is >= f.Size, bail out. PJW: really do this?
	for done < len(p) { // PJW: check err?
		loc, err := fb.fd.Seek(0, io.SeekCurrent)
		if loc != fb.bufloc+int64(fb.bufpos) {
			panic(fmt.Sprintf("%v loc=%d bufloc=%d bufpos=%d", err, loc, fb.bufloc,
				fb.bufpos))
		}
		fb.seeks++
		if loc >= fb.size {
			// PJW: this doesn't smell right
			fb.bufpos = int32(len(fb.buf))
			fb.bufloc = loc - int64(fb.fixedlen)
			return done, io.EOF
		}
		n, err := fb.fd.Read(fb.fixed[:])
		if n != 0 {
			fb.fixedlen = n
		}
		fb.reads++
		m := copy(p[done:], fb.fixed[:n])
		done += m
		if err != nil {
			if err == io.EOF { //PJW: check n==0? Is bufpos right?
				fb.bufpos = int32(len(fb.buf))
				fb.bufloc = loc - int64(fb.fixedlen)
				return done, io.EOF
			}
			return 0, err
		}
		fb.bytes += int64(n)
		fb.bufpos = int32(m)
		fb.bufloc = loc
		fb.buf = fb.fixed[:n]
	}
	return len(p), nil
}

// Seek implements io.Seeker. (<unchanged>, io.EOF) is returned for seeks off the end.
func (fb *fbuf) Seek(offset int64, whence int) (int64, error) {
	seekpos := offset
	switch whence {
	case io.SeekCurrent:
		seekpos += fb.bufloc + int64(fb.bufpos)
	case io.SeekEnd:
		seekpos += fb.size
	}
	if seekpos < 0 || seekpos > fb.size { // PJW: seekpos >= f.Size? No
		return fb.bufloc + int64(fb.bufpos), io.EOF
	}
	// if seekpos is inside fixed, just adjust buf and bufpos
	if seekpos >= fb.bufloc && seekpos <= int64(fb.fixedlen)+fb.bufloc {
		fb.bufpos = int32(seekpos - fb.bufloc)
		return seekpos, nil
	}
	// PJW: are you sure this is right? it's right, but do we do too much reading?
	// PJW: should we forbid seeking past the end?
	fb.buf, fb.bufpos, fb.bufloc = nil, 0, seekpos
	n, err := fb.fd.Seek(seekpos, io.SeekStart)
	fb.seeks++
	if n != seekpos || err != nil {
		log.Fatalf("seek failed %v %d != %d", err, n, seekpos) // PJW!
		return -1, fmt.Errorf("seek failed (%d!= %d) %v", n, seekpos,
			err)
	}
	return seekpos, nil
}

// Close closes the underlying file
func (fb *fbuf) Close() error {
	if fb.fd != nil {
		return fb.fd.Close()
	}
	return nil
}
