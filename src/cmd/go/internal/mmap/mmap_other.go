// Copyright 2022 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build plan9 || solaris

package mmap

import (
	"io"
	"log"
	"os"
)

// mmapFile on other systems doesn't mmap the file. It just reads everything.
func mmapFile(f *os.File) Data {
	b, err := io.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}
	return Data{f, b}
}
