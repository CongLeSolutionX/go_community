// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package renameio writes files atomically by renaming temporary files.
package renameio

import (
	"bytes"
	"io"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"

	"cmd/go/internal/robustio"
)

// WriteFile is like ioutil.WriteFile, but first writes data to an arbitrary
// file in the same directory as filename, then renames it atomically to the
// final name. The write is not complete until the wait func returns.
//
// That ensures that the final location, if it exists, is always a complete file.
func WriteFile(filename string, data []byte, perm fs.FileMode) (wait func(), err error) {

	f, err := tempFile(filepath.Dir(filename), filepath.Base(filename), perm)
	if err != nil {
		return nil, err
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

	if _, err := io.Copy(f, bytes.NewReader(data)); err != nil {
		return nil, err
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		// Sync the file before renaming it: otherwise, after a crash the reader may
		// observe a 0-length file instead of the actual contents.
		// See https://golang.org/issue/22397#issuecomment-380831736.
		// If an error occurs, try to clean up the temporary file. If the file
		// is not written to the cache, it will be written the next time it's fetched.
		if err := f.Sync(); err != nil {
			// Try to clean up.
			f.Close()
			os.Remove(f.Name())
			return
		}
		if err := f.Close(); err != nil {
			os.Remove(f.Name())
			return
		}

		robustio.Rename(f.Name(), filename)
	}()
	return func() { <-done }, nil
}

// tempFile creates a new temporary file with given permission bits.
func tempFile(dir, prefix string, perm fs.FileMode) (f *os.File, err error) {
	for i := 0; i < 10000; i++ {
		name := filepath.Join(dir, prefix+strconv.Itoa(rand.Intn(1000000000))+".tmp")
		f, err = os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_EXCL, perm)
		if os.IsExist(err) {
			continue
		}
		break
	}
	return
}
