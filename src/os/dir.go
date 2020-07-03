// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os

// ReadDir reads the contents of the directory associated with file and
// returns a slice of up to n FileInfo values, as would be returned
// by Lstat, in directory order. Subsequent calls on the same file will yield
// further FileInfos.
//
// If n > 0, ReadDir returns at most n FileInfo structures. In this case, if
// ReadDir returns an empty slice, it will return a non-nil error
// explaining why. At the end of a directory, the error is io.EOF.
//
// If n <= 0, ReadDir returns all the FileInfo from the directory in
// a single slice. In this case, if ReadDir succeeds (reads all
// the way to the end of the directory), it returns the slice and a
// nil error. If it encounters an error before the end of the
// directory, ReadDir returns the FileInfo read until that point
// and a non-nil error.
func (f *File) ReadDir(n int) ([]FileInfo, error) {
	if f == nil {
		return nil, ErrInvalid
	}
	return f.readdir(n)
}

// Readdir is an old name for ReadDir.
func (f *File) Readdir(n int) ([]FileInfo, error) {
	return f.ReadDir(n)
}

// ReadDirNames reads the contents of the directory associated with file
// and returns a slice of up to n names of files in the directory,
// in directory order. Subsequent calls on the same file will yield
// further names.
//
// If n > 0, ReadDirNames returns at most n names. In this case, if
// ReadDirNames returns an empty slice, it will return a non-nil error
// explaining why. At the end of a directory, the error is io.EOF.
//
// If n <= 0, ReadDirNames returns all the names from the directory in
// a single slice. In this case, if ReadDirNames succeeds (reads all
// the way to the end of the directory), it returns the slice and a
// nil error. If it encounters an error before the end of the
// directory, ReadDirNames returns the names read until that point and
// a non-nil error.
func (f *File) ReadDirNames(n int) (names []string, err error) {
	if f == nil {
		return nil, ErrInvalid
	}
	return f.readdirnames(n)
}

// Readdirnames is an old name for ReadDirNames.
func (f *File) Readdirnames(n int) ([]string, error) {
	return f.ReadDirNames(n)
}
