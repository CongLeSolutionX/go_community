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

	const goBuildPrefix = "\xff Go build ID: \""
	const goBuildEnd = "\"\n \xff"

	quoted := extractBetween(data[:n], goBuildPrefix, goBuildEnd)
	if len(quoted) == 0 {
		return fallbackBuildID(file, data)
	}

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

// extractBetween extracts data between start and end markers.
func extractBetween(data []byte, start, end string) []byte {
	i := bytes.Index(data, []byte(start))
	if i < 0 {
		return nil
	}

	k := bytes.Index(data[i+len(start):], []byte(end))
	if k < 0 {
		return nil
	}

	return data[i+len(start)-1 : i+len(start)+k+1]
}
