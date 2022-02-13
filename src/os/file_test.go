// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os_test

import (
	"internal/testenv"
	"io/fs"
	. "os"
	"path/filepath"
	"testing"
)

func TestDirFSReadLink(t *testing.T) {
	testenv.MustHaveSymlink(t)

	root := t.TempDir()
	subdir := filepath.Join(root, "dir")
	if err := Mkdir(subdir, 0o777); err != nil {
		t.Fatal(err)
	}
	links := map[string]string{
		filepath.Join(root, "parent-link"):        filepath.Join("..", "foo"),
		filepath.Join(root, "sneaky-parent-link"): filepath.Join("dir", "..", "..", "foo"),
		filepath.Join(root, "abs-link"):           filepath.Join(root, "foo"),
		filepath.Join(root, "rel-link"):           "foo",
		filepath.Join(root, "rel-sub-link"):       filepath.Join("dir", "foo"),
		filepath.Join(subdir, "parent-link"):      filepath.Join("..", "foo"),
	}
	for newname, oldname := range links {
		if err := Symlink(oldname, newname); err != nil {
			t.Fatal(err)
		}
	}

	fsys := DirFS(root)
	want := map[string]string{
		"rel-link":           "foo",
		"rel-sub-link":       "dir/foo",
		"dir/parent-link":    "../foo",
		"parent-link":        "",
		"sneaky-parent-link": "",
		"abs-parent-link":    "",
	}
	for name, want := range want {
		got, err := fs.ReadLink(fsys, name)
		if got != want || (err != nil) != (want == "") {
			wantErr := "<nil>"
			if want == "" {
				wantErr = "<error>"
			}
			t.Errorf("fs.ReadLink(fsys, %q) = %q, %v; want %q, %s", name, got, err, want, wantErr)
		}
	}
}

func TestDirFSLstat(t *testing.T) {
	testenv.MustHaveSymlink(t)

	root := t.TempDir()
	subdir := filepath.Join(root, "dir")
	if err := Mkdir(subdir, 0o777); err != nil {
		t.Fatal(err)
	}
	if err := Symlink("dir", filepath.Join(root, "link")); err != nil {
		t.Fatal(err)
	}

	fsys := DirFS(root)
	want := map[string]fs.FileMode{
		"link": fs.ModeSymlink,
		"dir":  fs.ModeDir,
	}
	for name, want := range want {
		info, err := fs.Lstat(fsys, name)
		var got fs.FileMode
		if info != nil {
			got = info.Mode().Type()
		}
		if got != want || err != nil {
			t.Errorf("fs.Lstat(fsys, %q).Mode().Type() = %v, %v; want %v, <nil>", name, got, err, want)
		}
	}
}
