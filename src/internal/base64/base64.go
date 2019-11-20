// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package base64 contains functionality used by tests to more easily
// work with base64-encoded testdata.
package base64

import (
	"encoding/base64"
	"io"
	"io/ioutil"
	"os"
)

// DecodeToTempFile decodes the named file to a temporary location.
// If successful, it returns the path of the decoded file.
// The caller is responsible for ensuring that the temporary file is removed.
func DecodeToTempFile(name string) (string, error) {
	f, err := os.Open(name)
	if err != nil {
		return "", err
	}
	defer f.Close()

	tmp, err := ioutil.TempFile("", "buildid-base64-decoded-")
	if err != nil {
		return "", err
	}
	defer tmp.Close()
	if _, err := io.Copy(tmp, base64.NewDecoder(base64.StdEncoding, f)); err != nil {
		return "", err
	}
	return tmp.Name(), nil
}
