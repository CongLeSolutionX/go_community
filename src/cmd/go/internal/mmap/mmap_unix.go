// Copyright 2011 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build unix && !solaris

package mmap

import (
	"log"
	"os"
	"syscall"
)

func mmapFile(f *os.File) Data {
	st, err := f.Stat()
	if err != nil {
		log.Fatal(err)
	}
	size := st.Size()
	pagesize := int64(os.Getpagesize())
	if int64(int(size+(pagesize-1))) != size+(pagesize-1) {
		log.Fatalf("%s: too large for mmap", f.Name())
	}
	n := int(size)
	if n == 0 {
		return Data{f, nil}
	}
	data, err := syscall.Mmap(int(f.Fd()), 0, (n+int(pagesize-1))&^int(pagesize-1), syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		log.Fatalf("mmap %s: %v", f.Name(), err)
	}
	return Data{f, data[:n]}
}
