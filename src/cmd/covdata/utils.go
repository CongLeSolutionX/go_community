// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"io"
	"os"
)

type mdWriteSeeker struct {
	payload []byte
	off     int64
}

func (d *mdWriteSeeker) Write(p []byte) (n int, err error) {
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

func (d *mdWriteSeeker) Seek(offset int64, whence int) (int64, error) {
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

func copyFile(inpath, outpath string) {
	inf, err := os.Open(inpath)
	if err != nil {
		fatal("opening input meta-data file %s: %v", inpath, err)
	}
	defer inf.Close()

	fi, err := inf.Stat()
	if err != nil {
		fatal("accessing input meta-data file %s: %v", inpath, err)
	}

	outf, err := os.OpenFile(outpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fi.Mode())
	if err != nil {
		fatal("opening output meta-data file %s: %v", outpath, err)
	}

	_, err = io.Copy(outf, inf)
	outf.Close()
	if err != nil {
		fatal("writing output meta-data file %s: %v", outpath, err)
	}
}
