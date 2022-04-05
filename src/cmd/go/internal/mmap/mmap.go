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
	debug.SetPanicOnFault(true)
}

// An mmapData is mmap'ed read-only data from a file.
type Data struct {
	File *os.File
	Data []byte
}

// mmap maps the given file into memory.
func Mmap(file string) Data {
	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	return mmapFile(f)
}
