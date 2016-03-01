// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin linux

package ld

import (
	"fmt"
	"os"
	"syscall"
)

// newObjReader loads an *objReader from a source file.
//
// Note that if files are mmaped, they are never unmaped. Slices onto
// files are stored in *LSym P fields, which live for the life of the
// linker process and so there is no earlier to point to unmap.
func newObjReader(file string) (*objReader, error) {
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
		return &objReader{}, nil
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

	input := &objReader{
		file: file,
		data: data,
	}
	return input, nil
}
