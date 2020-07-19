// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fs

import (
	"errors"
	pathpkg "path"
)

// SkipDir is used as a return value from WalkFuncs to indicate that
// the directory named in the call is to be skipped. It is not returned
// as an error by any function.
var SkipDir = errors.New("skip this directory")

// WalkDirFunc is the type of the function called for each file or directory
// visited by WalkDir. The path argument contains the argument to WalkDir as a
// prefix; that is, if WalkDir is called with "dir", which is a directory
// containing the file "a", the walk function will be called with argument
// "dir/a". The info argument is the FileInfo for the named path.
//
// If there was a problem walking to the file or directory named by path, the
// incoming error will describe the problem and the function can decide how
// to handle that error (and WalkDir will not descend into that directory). In the
// case of an error, the info argument will be nil. If an error is returned,
// processing stops. The sole exception is when the function returns the special
// value SkipDir. If the function returns SkipDir when invoked on a directory,
// WalkDir skips the directory's contents entirely. If the function returns SkipDir
// when invoked on a non-directory file, WalkDir skips the remaining files in the
// containing directory.
type WalkDirFunc func(path string, entry DirEntry, err error) error

// walk recursively descends path, calling fn.
func walk(fsys FS, path string, entry DirEntry, fn WalkDirFunc) error {
	if !entry.IsDir() {
		return fn(path, entry, nil)
	}

	dirs, err := ReadDir(fsys, path)
	err1 := fn(path, entry, err)
	// If err != nil, walk can't walk into this directory.
	// err1 != nil means fn want walk to skip this directory or stop walking.
	// Therefore, if one of err and err1 isn't nil, walk will return.
	if err != nil || err1 != nil {
		// The caller's behavior is controlled by the return value, which is decided
		// by fn. fn may ignore err and return nil.
		// If fn returns SkipDir, it will be handled by the caller.
		// So walk should return whatever fn returns.
		return err1
	}

	for _, entry := range dirs {
		filename := pathpkg.Join(path, entry.Name())
		err = walk(fsys, filename, entry, fn)
		if err != nil {
			if !entry.IsDir() || err != SkipDir {
				return err
			}
		}
	}
	return nil
}

// WalkDir walks the file tree rooted at root, calling fn for each file or
// directory in the tree, including root. All errors that arise visiting files
// and directories are filtered by fn. The files are walked in lexical
// order, which makes the output deterministic but means that for very
// large directories Walk can be inefficient.
func WalkDir(fsys FS, root string, fn WalkDirFunc) error {
	info, err := Stat(fsys, root)
	if err != nil {
		err = fn(root, nil, err)
	} else {
		err = walk(fsys, root, &statDirEntry{info}, fn)
	}
	if err == SkipDir {
		return nil
	}
	return err
}

type statDirEntry struct {
	info FileInfo
}

func (d *statDirEntry) Name() string            { return d.info.Name() }
func (d *statDirEntry) IsDir() bool             { return d.info.IsDir() }
func (d *statDirEntry) Type() FileMode          { return d.info.Mode().Type() }
func (d *statDirEntry) Info() (FileInfo, error) { return d.info, nil }
