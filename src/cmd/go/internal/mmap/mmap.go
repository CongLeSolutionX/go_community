// Copyright 2011 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mmap

import (
	"log"
	"os"
	"runtime/debug"
)

func init() {
	// set PanicOnFault to true so that the reader can recover
	// from errors reading the mmapped file.
	debug.SetPanicOnFault(true)
}

// Data is mmap'ed read-only data from a file.
// The backing file is never closed, so Data
// remains valid for the lifetime of the process.
// Errors accessing the underlying data will result in panics.
type Data struct {
	f    *os.File
	Data []byte
}

// Mmap maps the given file into memory.
func Mmap(file string) Data {
	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	return mmapFile(f)
}
