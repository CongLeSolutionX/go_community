// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package filepath

import (
	"strings"
	"syscall"
)

// normVolumeName is like VolumeName, but makes drive letter upper case.
// result of EvalSymlinks must be unique, so we have
// EvalSymlinks(`c:\a`) == EvalSymlinks(`C:\a`).
func normVolumeName(path string) string {
	volume := VolumeName(path)

	if len(volume) > 2 { // isUNC
		return volume
	}

	return strings.ToUpper(volume)
}

// normBase returns the last element of path with correct case.
func normBase(path string) (string, error) {
	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return "", err
	}

	var data syscall.Win32finddata

	h, err := syscall.FindFirstFile(p, &data)
	if err != nil {
		return "", err
	}
	syscall.FindClose(h)

	return syscall.UTF16ToString(data.FileName[:]), nil
}

// toNorm returns the normalized path that is guranteed to be unique.
// It should accepts following formats:
//   * UNC paths                              (e.g \\server\share\foo\bar)
//   * absolute paths                         (e.g C:\foo\bar)
//   * relative paths begin with drive letter (e.g C:foo\bar, C:..\foo\bar, C:.., C:.)
//   * relative paths begin with '\'          (e.g \foo\bar)
//   * relative paths begin without '\'       (e.g foo\bar, ..\foo\bar, .., .)
// The normalization should be done without breaking the given format.
// If two paths A and B are indicating same file with same format, toNorm(A) should be equal to toNorm(B).
func toNorm(path string, normBase func(string) (string, error)) (string, error) {
	if path == "" {
		return path, nil
	}

	path = Clean(path)

	volume := normVolumeName(path)
	path = path[len(volume):]

	// skip special cases
	if path == "." || path == `\` {
		return volume + path, nil
	}

	var normPath string

	for {
		i := strings.LastIndexByte(path, Separator)
		if path[i+1:] == ".." {
			normPath = path + `\` + normPath

			break
		}

		name, err := normBase(volume + path)
		if err != nil {
			return "", err
		}

		normPath = name + `\` + normPath

		if i == -1 {
			break
		}
		if i == 0 { // `\Go` or `C:\Go`
			normPath = `\` + normPath

			break
		}

		path = path[:i]
	}

	normPath = normPath[:len(normPath)-1] // remove trailing '\'

	return volume + normPath, nil
}

func evalSymlinks(path string) (string, error) {
	path, err := walkSymlinks(path)
	if err != nil {
		return "", err
	}
	return toNorm(path, normBase)
}
