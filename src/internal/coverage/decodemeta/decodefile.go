// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package decodemeta

// This package contains APIs and helpers for reading and decoding
// meta-data output files emitted by the runtime when a coverage-instrumented
// binary executes.

import (
	"bufio"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"internal/coverage"
	"internal/coverage/slicereader"
	"internal/coverage/stringtab"
	"io"
	"os"
	"unsafe"
)

type CoverageMetaFileReader struct {
	f          *os.File
	totlen     uint64
	npkgs      uint64
	strtaboff  uint32
	strtablen  uint32
	strtabents uint32
	modname    uint32
	cmode      coverage.CounterMode
	fileHash   [16]byte
	tmp        []byte
	pkgOffsets []uint64
	pkgLengths []uint64
	strtab     *stringtab.StringTableReader
	fileRdr    *bufio.Reader
	fileView   []byte
	debug      bool
}

// NewCoverageMetaFileReader returns a new helper object for reading
// the coverage meta-data output file 'f'. The param 'fileView' is a
// read-only slice containing the contents of 'f' obtained by mmap'ing
// the file read-only; 'fileView' may be nil, in which case the helper
// will read the contents of the file using regular file Read
// operations.
func NewCoverageMetaFileReader(f *os.File, fileView []byte) (*CoverageMetaFileReader, error) {
	r := &CoverageMetaFileReader{
		f:        f,
		fileView: fileView,
		tmp:      make([]byte, 256),
	}

	if err := r.readFileHeader(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *CoverageMetaFileReader) readFileHeader() error {
	var err error

	r.fileRdr = bufio.NewReader(r.f)

	// Read and verify magic string (plus 4 bytes of padding).
	ms := make([]byte, 8)
	if _, err = r.fileRdr.Read(ms); err != nil {
		return err
	}
	if ms[0] != coverage.CovMetaMagic[0] || ms[1] != coverage.CovMetaMagic[1] ||
		ms[2] != coverage.CovMetaMagic[2] || ms[3] != coverage.CovMetaMagic[3] {
		return fmt.Errorf("invalid magic string")
	}

	// Read length, entries and hash
	if r.totlen, err = r.rdUint64(); err != nil {
		return err
	}
	if r.npkgs, err = r.rdUint64(); err != nil {
		return err
	}
	var nr int
	nr, err = r.fileRdr.Read(r.fileHash[:])
	if err != nil || nr != int(unsafe.Sizeof(r.fileHash)) {
		return err
	}
	if r.strtaboff, err = r.rdUint32(); err != nil {
		return err
	}
	if r.strtablen, err = r.rdUint32(); err != nil {
		return err
	}
	if r.strtabents, err = r.rdUint32(); err != nil {
		return err
	}
	if r.modname, err = r.rdUint32(); err != nil {
		return err
	}
	var t uint32
	if t, err = r.rdUint32(); err != nil {
		return err
	}
	r.cmode = coverage.CounterMode(t)
	if _, err = r.rdUint32(); err != nil {
		return err
	}

	// Read package offsets for good measure
	r.pkgOffsets = make([]uint64, r.npkgs)
	for i := uint64(0); i < r.npkgs; i++ {
		if r.pkgOffsets[i], err = r.rdUint64(); err != nil {
			return err
		}
		if r.pkgOffsets[i] > r.totlen {
			return fmt.Errorf("insane pkg offset %d: %d > totlen %d",
				i, r.pkgOffsets[i], r.totlen)
		}
	}
	r.pkgLengths = make([]uint64, r.npkgs)
	for i := uint64(0); i < r.npkgs; i++ {
		if r.pkgLengths[i], err = r.rdUint64(); err != nil {
			return err
		}
		if r.pkgLengths[i] > r.totlen {
			return fmt.Errorf("insane pkg length %d: %d > totlen %d",
				i, r.pkgLengths[i], r.totlen)
		}
	}

	// Read string table.
	b := make([]byte, r.strtablen)
	nr, err = r.fileRdr.Read(b)
	if err != nil {
		return err
	}
	if nr != int(r.strtablen) {
		return fmt.Errorf("error: short read on string table")
	}
	slr := slicereader.NewReader(b, false /* not readonly */)
	r.strtab = stringtab.NewReader(slr)
	r.strtab.Read(int(r.strtabents))

	if r.debug {
		fmt.Fprintf(os.Stderr, "=-= read-in header is: %+v\n", *r)
	}

	return nil
}

func (r *CoverageMetaFileReader) rdUint64() (uint64, error) {
	r.tmp = r.tmp[:0]
	r.tmp = append(r.tmp, make([]byte, 8)...)
	n, err := r.fileRdr.Read(r.tmp)
	if err != nil {
		return 0, err
	}
	if n != 8 {
		return 0, fmt.Errorf("premature end of file on read")
	}
	v := binary.LittleEndian.Uint64(r.tmp)
	return v, nil
}

func (r *CoverageMetaFileReader) rdUint32() (uint32, error) {
	r.tmp = r.tmp[:0]
	r.tmp = append(r.tmp, make([]byte, 4)...)
	n, err := r.fileRdr.Read(r.tmp)
	if err != nil {
		return 0, err
	}
	if n != 4 {
		return 0, fmt.Errorf("premature end of file on read")
	}
	v := binary.LittleEndian.Uint32(r.tmp)
	return v, nil
}

func (r *CoverageMetaFileReader) NumPackages() uint64 {
	return r.npkgs
}

func (r *CoverageMetaFileReader) ModuleName() string {
	return r.strtab.Get(r.modname)
}

func (r *CoverageMetaFileReader) CounterMode() coverage.CounterMode {
	return r.cmode
}

func (r *CoverageMetaFileReader) FileHash() [16]byte {
	return r.fileHash
}

func (r *CoverageMetaFileReader) GetPackageDecoder(pkIdx uint32, payloadbuf []byte) (*CoverageMetaDataDecoder, []byte, error) {
	pp, err := r.GetPackagePayload(pkIdx, payloadbuf)
	if r.debug {
		fmt.Fprintf(os.Stderr, "=-= pkidx=%d payload length is %d hash=%s\n",
			pkIdx, len(pp), fmt.Sprintf("%x", md5.Sum(pp)))
	}
	if err != nil {
		return nil, nil, err
	}
	return NewCoverageMetaDataDecoder(pp, r.fileView != nil), pp, nil
}

func (r *CoverageMetaFileReader) GetPackagePayload(pkIdx uint32, payloadbuf []byte) ([]byte, error) {

	// Determine correct offset/length.
	if uint64(pkIdx) >= r.npkgs {
		return nil, fmt.Errorf("GetPackagePayload: illegal pkg index %d", pkIdx)
	}
	off := r.pkgOffsets[pkIdx]
	len := r.pkgLengths[pkIdx]

	if r.debug {
		fmt.Fprintf(os.Stderr, "=-= for pk %d, off=%d len=%d\n", pkIdx, off, len)
	}

	if r.fileView != nil {
		return r.fileView[off : off+len], nil
	}

	payload := payloadbuf[:0]
	if cap(payload) < int(len) {
		payload = make([]byte, 0, len)
	}
	payload = append(payload, make([]byte, len)...)
	if _, err := r.f.Seek(int64(off), os.SEEK_SET); err != nil {
		return nil, err
	}
	if _, err := io.ReadFull(r.f, payload); err != nil {
		return nil, err
	}
	return payload, nil
}
