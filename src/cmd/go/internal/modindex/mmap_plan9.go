// Copyright 2022 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package modindex

import (
	"io"
	"log"
	"os"
)

// mmapFile on plan9 doesn't mmap the file. It just reads everything.
func mmapFile(f *os.File) mmapData {
	b, err := io.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}
	return mmapData{f, b}
}
