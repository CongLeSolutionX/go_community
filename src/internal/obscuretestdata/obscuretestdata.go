// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package obscuretestdata contains functionality used by tests to more easily
// work with testdata that must be obscured primarily due to
// golang.org/issue/34986.
package obscuretestdata

import (
	"encoding/base64"
	"io"
	"io/ioutil"
	"os"
)

// DecodeToTempFile decodes the named file to a temporary location.
// If successful, it returns the path of the decoded file.
// The caller is responsible for ensuring that the temporary file is removed.
func DecodeToTempFile(name string) (path string, err error) {
	f, err := os.Open(name)
	if err != nil {
		return "", err
	}
	defer func() {
		if cerr := f.Close(); err == nil && cerr != nil {
			err = cerr
		}
	}()

	tmp, err := ioutil.TempFile("", "obscuretestdata-decoded-")
	if err != nil {
		return "", err
	}
	defer func() {
		if cerr := tmp.Close(); err == nil && cerr != nil {
			err = cerr
		}
	}()
	if _, err := io.Copy(tmp, base64.NewDecoder(base64.StdEncoding, f)); err != nil {
		return "", err
	}
	return tmp.Name(), nil
}
