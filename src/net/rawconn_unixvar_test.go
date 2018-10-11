// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build aix solaris

package net

import (
	"errors"
	"os"
)

func newSyscallConnFile() (string, *os.File, error) {
	return "", nil, errors.New("not supported")
}
