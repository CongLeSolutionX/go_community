// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package renameio writes files atomically by renaming temporary files.
package renameio

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

// WriteFile is like ioutil.WriteFile, but first writes data to an arbitrary
// file “near” filename, then renames it atomically to the final location.
//
// That ensures that the final location, if it exists, is always a complete file.
func WriteFile(filename string, data []byte) (err error) {
	f, err := ioutil.TempFile(filepath.Dir(filename), filepath.Base(filename)+".tmp-")
	if err != nil {
		return err
	}
	defer func() {
		// Only call os.Remove on f.Name() if we failed to rename it: otherwise,
		// some other process may have created a new file with the same name after
		// that.
		if err != nil {
			f.Close()
			os.Remove(f.Name())
		}
	}()

	if _, err = f.Write(data); err != nil {
		return err
	}
	// Sync the file before renaming it: otherwise, after a crash the reader may
	// observe a 0-length file instead of the actual contents.
	// See https://golang.org/issue/22397#issuecomment-380831736.
	if err = f.Sync(); err != nil {
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), filename)
}
