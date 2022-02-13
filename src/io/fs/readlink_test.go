// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fs_test

import (
	. "io/fs"
	"testing"
	"testing/fstest"
)

type linkReadFS struct {
	FS
	links map[string]string
}

func (fsys linkReadFS) ReadLink(name string) (string, error) {
	if !ValidPath(name) {
		return "", &PathError{Op: "readlink", Path: name, Err: ErrInvalid}
	}
	target, ok := fsys.links[name]
	if !ok {
		return "", &PathError{Op: "readlink", Path: name, Err: ErrNotExist}
	}
	return target, nil
}

func TestReadLink(t *testing.T) {
	testFS := linkReadFS{
		FS: fstest.MapFS{
			"foo": {
				Mode: ModeSymlink | 0o777,
			},
			"bar": {
				Data: []byte("Hello, World!\n"),
				Mode: 0o644,
			},

			"dir/parentlink": {
				Mode: ModeSymlink | 0o777,
			},
			"dir/link": {
				Mode: ModeSymlink | 0o777,
			},
			"dir/file": {
				Data: []byte("Hello, World!\n"),
				Mode: 0o644,
			},
		},
		links: map[string]string{
			"foo":            "bar",
			"dir/parentlink": "../bar",
			"dir/link":       "file",
		},
	}

	check := func(fsys FS, name string, want string) {
		t.Helper()
		got, err := ReadLink(fsys, name)
		if got != want || err != nil {
			t.Errorf("ReadLink(%q) = %q, %v; want %q, <nil>", name, got, err, want)
		}
	}

	check(testFS, "foo", "bar")
	check(testFS, "dir/parentlink", "../bar")
	check(testFS, "dir/link", "file")

	// Test that ReadLink on Sub works.
	sub, err := Sub(testFS, "dir")
	if err != nil {
		t.Fatal(err)
	}

	check(sub, "link", "file")
	if got, err := ReadLink(sub, "parentlink"); err == nil {
		t.Errorf("ReadLink(\"parentlink\") = %q, %v; want \"\", <error>", got, err)
	}
}
