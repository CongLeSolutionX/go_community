// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build aix || dragonfly || freebsd || (js && wasm) || linux || netbsd || openbsd || solaris
// +build aix dragonfly freebsd js,wasm linux netbsd openbsd solaris

package mime

import (
	"bufio"
	"os"
	"strings"
)

func init() {
	osInitMime = initMimeUnix
}

var typeFiles = []string{
	"/usr/local/share/mime/globs",
	"/usr/share/mime/globs", // Fallback for unix systems that don't use /usr/local.
}

func loadMimeFile(filename string) {
	f, err := os.Open(filename)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		if scanner.Text() == "" || scanner.Text()[0] == '#' {
			continue
		}

		// Each line should be of format: mimetype:*.ext
		fields := strings.Split(scanner.Text(), ":")
		if fields[1][0] != '*' {
			continue // We only support getting mimetypes for extensions, not for filenames.
		}

		setExtensionType(fields[1][1:], fields[0])
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
}

func initMimeUnix() {
	for _, filename := range typeFiles {
		loadMimeFile(filename)
	}
}

func initMimeForTests() map[string]string {
	typeFiles = []string{"testdata/test.types"}
	return map[string]string{
		".T1":  "application/test",
		".t2":  "text/test; charset=utf-8",
		".png": "image/png",
	}
}
