// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pprof

import (
	"bytes"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

// peBuildID returns a build ID from the binary.
func peBuildID(file string) string {
	// We'll try to find the build ID in the first 32 kB of the binary.
	const readSize = 32 * 1024
	data := make([]byte, readSize)

	f, err := os.Open(file)
	if err != nil {
		return ""
	}
	defer f.Close()

	var n int
	n, err = io.ReadFull(f, data)
	if err == io.ErrUnexpectedEOF {
		err = nil
	}
	if err != nil {
		return fallbackBuildID(file, data)
	}
	data = data[:n]

	const goBuildPrefix = "\xff Go build ID: \""
	const goBuildEnd = "\"\n \xff"

	i := bytes.Index(data, []byte(goBuildPrefix))
	if i < 0 {
		return fallbackBuildID(file, data)
	}

	k := bytes.Index(data[i+len(goBuildPrefix):], []byte(goBuildEnd))
	if k < 0 {
		return fallbackBuildID(file, data)
	}

	quoted := data[i+len(goBuildPrefix)-1 : i+len(goBuildPrefix)+k+1]
	id, err := strconv.Unquote(string(quoted))
	if err != nil {
		return fallbackBuildID(file, data)
	}

	return id
}

// fallbackBuildID computes an hash of the filename and 32 kB of the binary.
func fallbackBuildID(file string, data []byte) string {
	// using crc32, because it's already a dependency from pprof gzip.
	nameHash := crc32.ChecksumIEEE([]byte(filepath.Base(file)))
	dataHash := crc32.ChecksumIEEE(data)
	return fmt.Sprintf("%04x%04x", nameHash, dataHash)
}
