// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os_test

import (
	"io/fs"
	. "os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDirFSReadLink(t *testing.T) {
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
			if runtime.GOOS == "windows" {
				// Windows permits symlinks under certain conditions, but not always.
				// If the symlink operation fails, skip it.
				t.Skipf("Could not create symlink on %s: %v", runtime.GOOS, err)
			}
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
