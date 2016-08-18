// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package testenv provides information about what functionality
// is available in different testing environments run by the Go team.
//
// It is an internal package because these details are specific
// to the Go team's test setup (on build.golang.org) and not
// fundamental to tests in general.
package testenv

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"syscall"
)

var once sync.Once
var winHasSymlink = true

func hasSymlink() bool {
	once.Do(func() {
		tmpdir, err := ioutil.TempDir("", "symtest")
		if err != nil {
			panic("failed to create temp directory: " + err.Error())
		}
		defer os.RemoveAll(tmpdir)

		err = os.Symlink("target", filepath.Join(tmpdir, "symlink"))
		if err != nil {
			err = err.(*os.LinkError).Err
			switch err {
			case syscall.EWINDOWS, syscall.ERROR_PRIVILEGE_NOT_HELD:
				winHasSymlink = false
			}
		}
		os.Remove("target")
	})
	return winHasSymlink
}
