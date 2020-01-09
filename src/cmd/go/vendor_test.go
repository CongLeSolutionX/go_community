// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Tests for vendoring semantics.

package main_test

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestVendorGOPATH(t *testing.T) {
	tg := testgo(t)
	defer tg.cleanup()
	changeVolume := func(s string, f func(s string) string) string {
		vol := filepath.VolumeName(s)
		return f(vol) + s[len(vol):]
	}
	gopath := changeVolume(filepath.Join(tg.pwd(), "testdata"), strings.ToLower)
	tg.setenv("GOPATH", gopath)
	cd := changeVolume(filepath.Join(tg.pwd(), "testdata/src/vend/hello"), strings.ToUpper)
	tg.cd(cd)
	tg.run("run", "hello.go")
	tg.grepStdout("hello, world", "missing hello world output")
}
