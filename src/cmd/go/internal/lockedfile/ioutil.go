// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lockedfile

// This file contains helpers that mimic the API of the standard io/ioutil
// package.

import (
	"io"
	"io/ioutil"
	"os"
)

// Read opens the named file with a read-lock and returns its contents.
func Read(name string) ([]byte, error) {
	f, err := Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return ioutil.ReadAll(f)
}

// Write opens the named file (creating it with the given permissions if needed),
// then write-locks it and overwrites it with the given content.
func Write(name string, content io.Reader, perm os.FileMode) (err error) {
	f, err := OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}

	_, err = io.Copy(f, content)
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	return err
}
