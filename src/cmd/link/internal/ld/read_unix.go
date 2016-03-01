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

func mmap(f *os.File, size int64) []byte {
	data, err := syscall.Mmap(int(f.Fd()), 0, int(size), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_PRIVATE)
	if err != nil {
		panic(fmt.Sprintf("ld: mmap of %q failed: %v", f.Name(), err))
	}
	return data
}
