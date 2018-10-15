// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package modfetch

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"cmd/go/internal/modfetch/codehost"
	"cmd/go/internal/module"
	"cmd/go/internal/str"
)

func Unzip(dir, zipfile, prefix string) error {
	maxSize := int64(codehost.MaxZipFile)

	// We may have created dir previously and even marked it read-only, only to
	// later fail due to a hash mismatch. Start by cleaning out the directory.
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			os.Chmod(dir, 0777)
		}
		return nil
	})
	if !os.IsNotExist(err) {
		if err := os.RemoveAll(dir); err != nil {
			return err
		}
	}
	if err := os.MkdirAll(dir, 0777); err != nil {
		return err
	}

	f, err := os.Open(zipfile)
	if err != nil {
		return err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return err
	}

	z, err := zip.NewReader(f, info.Size())
	if err != nil {
		return fmt.Errorf("unzip %v: %s", zipfile, err)
	}

	foldPath := make(map[string]string)
	var checkFold func(string) error
	checkFold = func(name string) error {
		fold := str.ToFold(name)
		if foldPath[fold] == name {
			return nil
		}
		dir := path.Dir(name)
		if dir != "." {
			if err := checkFold(dir); err != nil {
				return err
			}
		}
		if foldPath[fold] == "" {
			foldPath[fold] = name
			return nil
		}
		other := foldPath[fold]
		return fmt.Errorf("unzip %v: case-insensitive file name collision: %q and %q", zipfile, other, name)
	}

	// Check total size, valid file names.
	var size int64
	for _, zf := range z.File {
		if !str.HasPathPrefix(zf.Name, prefix) {
			return fmt.Errorf("unzip %v: unexpected file name %s", zipfile, zf.Name)
		}
		if zf.Name == prefix || strings.HasSuffix(zf.Name, "/") {
			continue
		}
		name := zf.Name[len(prefix)+1:]
		if err := module.CheckFilePath(name); err != nil {
			return fmt.Errorf("unzip %v: %v", zipfile, err)
		}
		if err := checkFold(name); err != nil {
			return err
		}
		if path.Clean(zf.Name) != zf.Name || strings.HasPrefix(zf.Name[len(prefix)+1:], "/") {
			return fmt.Errorf("unzip %v: invalid file name %s", zipfile, zf.Name)
		}
		s := int64(zf.UncompressedSize64)
		if s < 0 || maxSize-size < s {
			return fmt.Errorf("unzip %v: content too large", zipfile)
		}
		size += s
	}

	// Unzip, enforcing sizes checked earlier.
	dirs := map[string]bool{dir: true}
	for _, zf := range z.File {
		if zf.Name == prefix || strings.HasSuffix(zf.Name, "/") {
			continue
		}
		name := zf.Name[len(prefix):]
		dst := filepath.Join(dir, name)
		parent := filepath.Dir(dst)
		for parent != dir {
			dirs[parent] = true
			parent = filepath.Dir(parent)
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0777); err != nil {
			return err
		}
		// Since we removed the existing contents above, the only way dst can exist
		// here is if our locking somehow failed and two concurrent 'go' commands
		// are unzipping into the same directory simultaneously. If that happens,
		// they'll still be fine as long as they overwrite the file with exactly the
		// same data, so don't set O_TRUNC.
		w, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE, 0444)
		if err != nil {
			return fmt.Errorf("unzip %v: %v", zipfile, err)
		}
		r, err := zf.Open()
		if err != nil {
			w.Close()
			return fmt.Errorf("unzip %v: %v", zipfile, err)
		}
		lr := &io.LimitedReader{R: r, N: int64(zf.UncompressedSize64) + 1}
		_, err = io.Copy(w, lr)
		r.Close()
		if err != nil {
			w.Close()
			return fmt.Errorf("unzip %v: %v", zipfile, err)
		}
		if err := w.Close(); err != nil {
			return fmt.Errorf("unzip %v: %v", zipfile, err)
		}
		if lr.N <= 0 {
			return fmt.Errorf("unzip %v: content too large", zipfile)
		}
	}

	// Mark directories unwritable, best effort.
	var dirlist []string
	for dir := range dirs {
		dirlist = append(dirlist, dir)
	}
	// Sort reversed to chmod children before parents.
	sort.Sort(sort.Reverse(sort.StringSlice(dirlist)))
	makeReadonly(dirlist...)

	return nil
}

func makeReadonly(paths ...string) (firstErr error) {
	for _, path := range paths {
		mode := os.FileMode(0444)
		if fi, err := os.Stat(path); err == nil {
			mode = fi.Mode() &^ 0222 // Remove write bits from existing permissions.
		}
		if err := os.Chmod(path, mode); firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
