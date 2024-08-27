// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !plan9

package os_test

import (
	. "os"
	"runtime"
	"strings"
	"testing"
)

func TestGetwdDeep(t *testing.T) {
	testGetwdDeep(t, false)
}

func TestGetwdDeepWithPWDSet(t *testing.T) {
	testGetwdDeep(t, true)
}

// testGetwdDeep checks that os.Getwd is able to return paths
// longer than syscall.PathMax (with or without PWD set).
func testGetwdDeep(t *testing.T, setPWD bool) {
	switch runtime.GOOS {
	case "js":
		t.Skip("not supported on js: TempDir RemoveAll cleanup fails on long paths")
	case "wasip1":
		t.Skip("not supported on wasip1: Chdir won't work if wd is too deep")
	case "windows":
		t.Skip("not supported on windows: Mkdir won't work if wd is too deep")
	}
	dir := t.TempDir()
	t.Chdir(dir)

	if setPWD {
		t.Setenv("PWD", dir)
	} else {
		// When testing os.Getwd, setting PWD to empty string
		// is the same as unsetting it, but the latter would
		// be more complicated since we don't have t.Unsetenv.
		t.Setenv("PWD", "")
	}

	name := strings.Repeat("a", 200)
	for {
		if err := Mkdir(name, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := Chdir(name); err != nil {
			t.Fatal(err)
		}
		if setPWD {
			dir += "/" + name
			if err := Setenv("PWD", dir); err != nil {
				t.Fatal(err)
			}
			t.Logf(" $PWD len: %d", len(dir))
		}

		wd, err := Getwd()
		t.Logf("Getwd len: %d %q", len(wd), wd)
		if err != nil {
			t.Fatal(err)
		}
		// Ideally the success criterion should be len(wd) > syscall.PathMax,
		// but the latter is not public for some platforms, so use Stat(wd).
		// When it fails with ENAMETOOLONG, it means:
		//  - wd is longer than PathMax;
		//  - Getwd used the slow fallback code.
		if _, err := Stat(wd); err != nil {
			t.Logf("success with len(wd)=%d, got error %v from stat", len(wd), err)
			break
		}
	}
}
