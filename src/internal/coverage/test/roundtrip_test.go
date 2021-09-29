// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package test

import (
	"fmt"
	"internal/coverage"
	"internal/coverage/decodemeta"
	"internal/coverage/encodemeta"
	"os"
	"path/filepath"
	"testing"
)

type dummyReadWriteSeeker struct {
	payload []byte
	off     int64
}

func (d *dummyReadWriteSeeker) Write(p []byte) (n int, err error) {
	amt := len(p)
	towrite := d.payload[d.off:]
	if len(towrite) < amt {
		d.payload = append(d.payload, make([]byte, amt-len(towrite))...)
		towrite = d.payload[d.off:]
	}
	copy(towrite, p)
	d.off += int64(amt)
	return amt, nil
}

func (d *dummyReadWriteSeeker) Seek(offset int64, whence int) (int64, error) {
	if whence == os.SEEK_SET {
		d.off = offset
		return offset, nil
	} else if whence == os.SEEK_CUR {
		d.off += offset
		return d.off, nil
	}
	// other modes not supported
	panic("bad")
}

func (d *dummyReadWriteSeeker) Read(p []byte) (n int, err error) {
	amt := len(p)
	toread := d.payload[d.off:]
	if len(toread) < amt {
		amt = len(toread)
	}
	copy(p, toread)
	d.off += int64(amt)
	return amt, nil
}

func cmpFuncDesc(want, got coverage.FuncDesc) string {
	swant := fmt.Sprintf("%+v", want)
	sgot := fmt.Sprintf("%+v", got)
	if swant == sgot {
		return ""
	}
	return fmt.Sprintf("wanted %q got %q", swant, sgot)
}

func TestMetaDataEmptyPackage(t *testing.T) {
	// Make sure that encoding/decoding works properly with packages
	// that don't actually have any functions.
	p := "empty/package"
	b := encodemeta.NewCoverageMetaDataBuilder(p)
	drws := &dummyReadWriteSeeker{}
	b.Emit(drws)
	drws.off = 0
	dec := decodemeta.NewCoverageMetaDataDecoder(drws.payload, false)
	nf := dec.NumFuncs()
	if nf != 0 {
		t.Errorf("dec.NumFuncs(): got %d want %d", nf, 0)
	}
	pp := dec.PackagePath()
	if pp != p {
		t.Errorf("dec.PackagePath(): got %s want %s", pp, p)
	}
}

func TestMetaDataEncoderDecoder(t *testing.T) {
	// Test encode path.
	b := encodemeta.NewCoverageMetaDataBuilder("foopkg")
	f1 := coverage.FuncDesc{
		Funcname: "func",
		Srcfile:  "foo.go",
		Units: []coverage.CoverableUnit{
			coverage.CoverableUnit{StLine: 1, StCol: 2, EnLine: 3, EnCol: 4, NxStmts: 5},
			coverage.CoverableUnit{StLine: 6, StCol: 7, EnLine: 8, EnCol: 9, NxStmts: 10},
		},
	}
	idx := b.AddFunc(f1)
	if idx != 0 {
		t.Errorf("b.AddFunc(f1) got %d want %d", idx, 0)
	}

	f2 := coverage.FuncDesc{
		Funcname: "xfunc",
		Srcfile:  "bar.go",
		Units: []coverage.CoverableUnit{
			coverage.CoverableUnit{StLine: 1, StCol: 2, EnLine: 3, EnCol: 4, NxStmts: 5},
			coverage.CoverableUnit{StLine: 6, StCol: 7, EnLine: 8, EnCol: 9, NxStmts: 10},
			coverage.CoverableUnit{StLine: 11, StCol: 12, EnLine: 13, EnCol: 14, NxStmts: 15},
		},
	}
	idx = b.AddFunc(f2)
	if idx != 1 {
		t.Errorf("b.AddFunc(f2) got %d want %d", idx, 0)
	}

	// Emit into a dummer writer.
	drws := &dummyReadWriteSeeker{}
	b.Emit(drws)

	if false {
		fmt.Fprintf(os.Stderr, "=-= payload len is %d", len(drws.payload))
		for i := 0; i < len(drws.payload); i++ {
			if i%8 == 0 {
				fmt.Fprintf(os.Stderr, "\n%d: ", i)
			}
			fmt.Fprintf(os.Stderr, " %x:%c", drws.payload[i], drws.payload[i])
		}
		fmt.Fprintf(os.Stderr, "\n")
	}

	// Test decode path.
	drws.off = 0
	dec := decodemeta.NewCoverageMetaDataDecoder(drws.payload, false)
	nf := dec.NumFuncs()
	if nf != 2 {
		t.Errorf("dec.NumFuncs(): got %d want %d", nf, 2)
	}

	cases := []coverage.FuncDesc{f1, f2}
	for i := uint32(0); i < uint32(len(cases)); i++ {
		var fn coverage.FuncDesc
		if err := dec.ReadFunc(i, &fn); err != nil {
			t.Fatalf("err reading function %d: %v", i, err)
		}
		res := cmpFuncDesc(cases[i], fn)
		if res != "" {
			t.Errorf("ReadFunc(%d): %s", i, res)
		}
	}
}

func createFuncs(i int) []coverage.FuncDesc {
	res := []coverage.FuncDesc{}
	lc := uint32(1)
	for fi := 0; fi < i+1; fi++ {
		units := []coverage.CoverableUnit{}
		for ui := 0; ui < (fi+1)*(i+1); ui++ {
			units = append(units,
				coverage.CoverableUnit{StLine: lc, StCol: lc + 1,
					EnLine: lc + 2, EnCol: lc + 3, NxStmts: lc + 4,
				})
			lc += 5
		}
		f := coverage.FuncDesc{
			Funcname: fmt.Sprintf("func_%d_%d", i, fi),
			Srcfile:  fmt.Sprintf("foo_%d.go", i),
			Units:    units,
		}
		res = append(res, f)
	}
	return res
}

func createBlob(i int) []byte {
	b := encodemeta.NewCoverageMetaDataBuilder("foopkg")
	funcs := createFuncs(i)
	for _, f := range funcs {
		b.AddFunc(f)
	}
	drws := &dummyReadWriteSeeker{}
	b.Emit(drws)
	return drws.payload
}

func createMetaDataBlobs(nb int) [][]byte {
	res := [][]byte{}
	for i := 0; i < nb; i++ {
		res = append(res, createBlob(i))
	}
	return res
}

func TestMetaDataFileWriterReader(t *testing.T) {
	d := t.TempDir()

	// Emit a meta-file...
	mfpath := filepath.Join(d, "covmeta.hash.0")
	of, err := os.OpenFile(mfpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		t.Fatalf("opening covmeta: %v", err)
	}
	t.Logf("meta-file path is %s", mfpath)
	blobs := createMetaDataBlobs(7)
	mfw := encodemeta.NewCoverageMetaFileWriter(mfpath, of)
	finalHash := [16]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	err = mfw.Write(finalHash, "myModule", blobs, coverage.CtrModeAtomic)
	if err != nil {
		t.Fatalf("writing meta-file: %v", err)
	}

	// ... then read it back in, first time without setting fileView,
	// second time setting it.
	for k := 0; k < 2; k++ {
		var fileView []byte

		inf, err := os.Open(mfpath)
		if err != nil {
			t.Fatalf("open() on meta-file: %v", err)
		}

		if k != 0 {
			// Use fileview to exercise different paths in reader.
			fi, err := os.Stat(mfpath)
			if err != nil {
				t.Fatalf("stat() on meta-file: %v", err)
			}
			fileView = make([]byte, fi.Size())
			if _, err := inf.Read(fileView); err != nil {
				t.Fatalf("read() on meta-file: %v", err)
			}
			if _, err := inf.Seek(int64(0), os.SEEK_SET); err != nil {
				t.Fatalf("seek() on meta-file: %v", err)
			}
		}

		mfr, err := decodemeta.NewCoverageMetaFileReader(inf, fileView)
		if err != nil {
			t.Fatalf("k=%d NewCoverageMetaFileReader failed with: %v", k, err)
		}
		np := mfr.NumPackages()
		if np != 7 {
			t.Fatalf("k=%d wanted 7 packages got %d", k, np)
		}
		mn := mfr.ModuleName()
		if mn != "myModule" {
			t.Fatalf("k=%d wanted mod name %q got %q", k, "myModule", mn)
		}
		md := mfr.CounterMode()
		wmd := coverage.CtrModeAtomic
		if md != wmd {
			t.Fatalf("k=%d wanted mode %d got %d", k, wmd, md)
		}

		payload := []byte{}
		for pi := 0; pi < int(np); pi++ {
			var pd *decodemeta.CoverageMetaDataDecoder
			var err error
			pd, payload, err = mfr.GetPackageDecoder(uint32(pi), payload)
			if err != nil {
				t.Fatalf("GetPackageDecoder(%d) failed with: %v", pi, err)
			}
			efuncs := createFuncs(pi)
			nf := pd.NumFuncs()
			if len(efuncs) != int(nf) {
				t.Fatalf("decoding pk %d wanted %d funcs got %d",
					pi, len(efuncs), nf)
			}
			var f coverage.FuncDesc
			for fi := 0; fi < int(nf); fi++ {
				if err := pd.ReadFunc(uint32(fi), &f); err != nil {
					t.Fatalf("ReadFunc(%d) pk %d got error %v",
						fi, pi, err)
				}
				res := cmpFuncDesc(efuncs[fi], f)
				if res != "" {
					t.Errorf("ReadFunc(%d) pk %d: %s", fi, pi, res)
				}
			}
		}
		inf.Close()
	}
}
