// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build unix

package os_test

import (
	. "os"
	"strings"
	"syscall"
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
		if len(wd) > syscall.PathMax*2 { // Success.
			break
		}
	}
}
