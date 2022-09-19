// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package coverage

import (
	"internal/goexperiment"
	"path/filepath"
	"testing"
	_ "unsafe"
)

//go:linkname testing_testGoCoverDir testing.testGoCoverDir
func testing_testGoCoverDir() string

func TestTestSupport(t *testing.T) {
	if !goexperiment.CoverageRedesign {
		return
	}
	if testing.CoverMode() == "" {
		return
	}
	t.Logf("testing.testGoCoverDir() returns %s\n",
		testing_testGoCoverDir())

	textfile := filepath.Join(t.TempDir(), "file.txt")
	err := processCoverTestDir(testing_testGoCoverDir(), textfile,
		testing.CoverMode(), "")
	if err != nil {
		t.Fatalf("bad: %v", err)
	}
}
