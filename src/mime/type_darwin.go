// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin
// +build darwin

package mime

import (
	"bufio"
	"os"
	"strings"
)

func init() {
	osInitMime = initMimeDarwin
}

var typeFiles = []string{
	"/etc/apache2/mime.types",
}

func loadMimeFile(filename string) {
	f, err := os.Open(filename)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) <= 1 || fields[0][0] == '#' {
			continue
		}
		mimeType := fields[0]
		for _, ext := range fields[1:] {
			if ext[0] == '#' {
				break
			}
			setExtensionType("."+ext, mimeType)
		}
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
}

func initMimeDarwin() {
	for _, filename := range typeFiles {
		loadMimeFile(filename)
	}
}

func initMimeForTests() map[string]string {
	typeFiles = []string{"testdata/test.types.darwin"}
	return map[string]string{
		".T1":  "application/test",
		".t2":  "text/test; charset=utf-8",
		".png": "image/png",
	}
}
