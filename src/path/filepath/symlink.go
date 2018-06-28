// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package filepath

import (
	"errors"
	"os"
	"runtime"
)

// isRoot returns true if path is root of file system
// (`/` on unix and `/`, `\`, `c:\` or `c:/` on windows).
func isRoot(path string) bool {
	if runtime.GOOS != "windows" {
		return path == "/"
	}
	switch len(path) {
	case 1:
		return os.IsPathSeparator(path[0])
	case 3:
		return path[1] == ':' && os.IsPathSeparator(path[2])
	}
	return false
}

// isDriveLetter returns true if path is Windows drive letter (like "c:").
func isDriveLetter(path string) bool {
	if runtime.GOOS != "windows" {
		return false
	}
	return len(path) == 2 && path[1] == ':'
}

func walkLink(path string, linksWalked *int) (newpath string, err error) {
	for {
		if *linksWalked > 255 {
			return "", errors.New("EvalSymlinks: too many links")
		}
		fi, err := os.Lstat(path)
		if err != nil {
			return "", err
		}
		if fi.Mode()&os.ModeSymlink == 0 {
			return path, nil
		}
		newpath, err = os.Readlink(path)
		if err != nil {
			return "", err
		}
		*linksWalked++
		if IsAbs(newpath) || os.IsPathSeparator(newpath[0]) {
			path = newpath
		} else {
			path = Join(path, "..", newpath)
		}
	}
}

func walkLinks(path string, linksWalked *int) (string, error) {
	switch dir, file := Split(path); {
	case dir == "":
		newpath, err := walkLink(file, linksWalked)
		return newpath, err
	case file == "":
		if isDriveLetter(dir) {
			return dir, nil
		}
		if os.IsPathSeparator(dir[len(dir)-1]) {
			if isRoot(dir) {
				return dir, nil
			}
			return walkLinks(dir[:len(dir)-1], linksWalked)
		}
		newpath, err := walkLink(dir, linksWalked)
		return newpath, err
	default:
		newdir, err := walkLinks(dir, linksWalked)
		if err != nil {
			return "", err
		}
		newpath, err := walkLink(Join(newdir, file), linksWalked)
		if err != nil {
			return "", err
		}
		return newpath, nil
	}
}

func walkSymlinks(path string) (string, error) {
	if path == "" {
		return path, nil
	}
	var linksWalked int // to protect against cycles
	for {
		i := linksWalked
		newpath, err := walkLinks(path, &linksWalked)
		if err != nil {
			return "", err
		}
		if runtime.GOOS == "windows" {
			// walkLinks(".", ...) always returns "." on unix.
			// But on windows it returns symlink target, if current
			// directory is a symlink. Stop the walk, if symlink
			// target is not absolute path, and return "."
			// to the caller (just like unix does).
			// Same for "C:.".
			if path[volumeNameLen(path):] == "." && !IsAbs(newpath) {
				return path, nil
			}
		}
		if i == linksWalked {
			return Clean(newpath), nil
		}
		path = newpath
	}
}
