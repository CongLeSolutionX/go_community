// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package filepath_test

import (
	"io/ioutil"
	"os"
	. "path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
)

type MatchTest struct {
	pattern, s string
	match      bool
	err        error
}

var matchTests = []MatchTest{
	{"abc", "abc", true, nil},
	{"*", "abc", true, nil},
	{"*c", "abc", true, nil},
	{"a*", "a", true, nil},
	{"a*", "abc", true, nil},
	{"a*", "ab/c", false, nil},
	{"a*/b", "abc/b", true, nil},
	{"a*/b", "a/c/b", false, nil},
	{"a*b*c*d*e*/f", "axbxcxdxe/f", true, nil},
	{"a*b*c*d*e*/f", "axbxcxdxexxx/f", true, nil},
	{"a*b*c*d*e*/f", "axbxcxdxe/xxx/f", false, nil},
	{"a*b*c*d*e*/f", "axbxcxdxexxx/fff", false, nil},
	{"a*b?c*x", "abxbbxdbxebxczzx", true, nil},
	{"a*b?c*x", "abxbbxdbxebxczzy", false, nil},
	{"ab[c]", "abc", true, nil},
	{"ab[b-d]", "abc", true, nil},
	{"ab[e-g]", "abc", false, nil},
	{"ab[^c]", "abc", false, nil},
	{"ab[^b-d]", "abc", false, nil},
	{"ab[^e-g]", "abc", true, nil},
	{"a\\*b", "a*b", true, nil},
	{"a\\*b", "ab", false, nil},
	{"a?b", "a☺b", true, nil},
	{"a[^a]b", "a☺b", true, nil},
	{"a???b", "a☺b", false, nil},
	{"a[^a][^a][^a]b", "a☺b", false, nil},
	{"[a-ζ]*", "α", true, nil},
	{"*[a-ζ]", "A", false, nil},
	{"a?b", "a/b", false, nil},
	{"a*b", "a/b", false, nil},
	{"[\\]a]", "]", true, nil},
	{"[\\-]", "-", true, nil},
	{"[x\\-]", "x", true, nil},
	{"[x\\-]", "-", true, nil},
	{"[x\\-]", "z", false, nil},
	{"[\\-x]", "x", true, nil},
	{"[\\-x]", "-", true, nil},
	{"[\\-x]", "a", false, nil},
	{"[]a]", "]", false, ErrBadPattern},
	{"[-]", "-", false, ErrBadPattern},
	{"[x-]", "x", false, ErrBadPattern},
	{"[x-]", "-", false, ErrBadPattern},
	{"[x-]", "z", false, ErrBadPattern},
	{"[-x]", "x", false, ErrBadPattern},
	{"[-x]", "-", false, ErrBadPattern},
	{"[-x]", "a", false, ErrBadPattern},
	{"\\", "a", false, ErrBadPattern},
	{"[a-b-c]", "a", false, ErrBadPattern},
	{"[", "a", false, ErrBadPattern},
	{"[^", "a", false, ErrBadPattern},
	{"[^bc", "a", false, ErrBadPattern},
	{"a[", "a", false, nil},
	{"a[", "ab", false, ErrBadPattern},
	{"*x", "xxx", true, nil},
}

func errp(e error) string {
	if e == nil {
		return "<nil>"
	}
	return e.Error()
}

func TestMatch(t *testing.T) {
	for _, tt := range matchTests {
		pattern := tt.pattern
		s := tt.s
		if runtime.GOOS == "windows" {
			if strings.Index(pattern, "\\") >= 0 {
				// no escape allowed on windows.
				continue
			}
			pattern = Clean(pattern)
			s = Clean(s)
		}
		ok, err := Match(pattern, s)
		if ok != tt.match || err != tt.err {
			t.Errorf("Match(%#q, %#q) = %v, %q want %v, %q", pattern, s, ok, errp(err), tt.match, errp(tt.err))
		}
	}
}

// contains returns true if vector contains the string s.
func contains(vector []string, s string) bool {
	for _, elem := range vector {
		if elem == s {
			return true
		}
	}
	return false
}

var globTests = []struct {
	pattern, result string
}{
	{"match.go", "match.go"},
	{"mat?h.go", "match.go"},
	{"*", "match.go"},
	{"../*/match.go", "../filepath/match.go"},
}

func TestGlob(t *testing.T) {
	for _, tt := range globTests {
		pattern := tt.pattern
		result := tt.result
		if runtime.GOOS == "windows" {
			pattern = Clean(pattern)
			result = Clean(result)
		}
		matches, err := Glob(pattern)
		if err != nil {
			t.Errorf("Glob error for %q: %s", pattern, err)
			continue
		}
		if !contains(matches, result) {
			t.Errorf("Glob(%#q) = %#v want %v", pattern, matches, result)
		}
	}
	for _, pattern := range []string{"no_match", "../*/no_match"} {
		matches, err := Glob(pattern)
		if err != nil {
			t.Errorf("Glob error for %q: %s", pattern, err)
			continue
		}
		if len(matches) != 0 {
			t.Errorf("Glob(%#q) = %#v want []", pattern, matches)
		}
	}
}

func TestGlobError(t *testing.T) {
	_, err := Glob("../../*/*/[7")
	if err == nil {
		t.Error("expected error for bad pattern; got nil")
	}
}

var globSymlinkTests = []struct {
	path, dest string
	brokenLink bool
}{
	{"test1", "link1", false},
	{"test2", "link2", true},
}

func TestGlobSymlink(t *testing.T) {
	switch runtime.GOOS {
	case "android", "nacl", "plan9":
		t.Skipf("skipping on %s", runtime.GOOS)
	case "windows":
		if !supportsSymlinks {
			t.Skipf("skipping on %s", runtime.GOOS)
		}

	}

	tmpDir, err := ioutil.TempDir("", "globsymlink")
	if err != nil {
		t.Fatal("creating temp dir:", err)
	}
	defer os.RemoveAll(tmpDir)

	for _, tt := range globSymlinkTests {
		path := Join(tmpDir, tt.path)
		dest := Join(tmpDir, tt.dest)
		f, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
		err = os.Symlink(path, dest)
		if err != nil {
			t.Fatal(err)
		}
		if tt.brokenLink {
			// Break the symlink.
			os.Remove(path)
		}
		matches, err := Glob(dest)
		if err != nil {
			t.Errorf("GlobSymlink error for %q: %s", dest, err)
		}
		if !contains(matches, dest) {
			t.Errorf("Glob(%#q) = %#v want %v", dest, matches, dest)
		}
	}
}

func TestGlobOrder(t *testing.T) {
	root, err := ioutil.TempDir("", "goglob")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)

	const (
		dirCount   = 5
		filePerDir = 20
		totalFiles = dirCount * filePerDir
	)
	createDirs(t, root, []int{dirCount}, []string{""}, filePerDir)

	files, err := Glob(Join(root, "*/*"))
	if err != nil {
		t.Fatal(err)
	}

	if got := len(files); got != totalFiles {
		t.Errorf("got %d files; want %d", got, totalFiles)
	}

	if !sort.IsSorted(sort.StringSlice(files)) {
		t.Fatal("not sorted")
	}
}

func benchmarkGlob(b *testing.B, files int, dirs ...int) {
	root, err := ioutil.TempDir("", "goglob")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(root)

	prefixes := []string{"x", "y", "z"}
	createDirs(b, root, dirs, prefixes, files)

	pattern := ""
	for range dirs {
		pattern += "*/"
	}
	pattern += "x*"
	pattern = Join(root, pattern)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Glob(pattern)
	}
}

func BenchmarkGlob_10_10(b *testing.B) {
	benchmarkGlob(b, 10, 10)
}

func BenchmarkGlob_10_10_10(b *testing.B) {
	benchmarkGlob(b, 10, 10, 10)
}

func BenchmarkGlob_5_5_5_5(b *testing.B) {
	benchmarkGlob(b, 5, 5, 5, 5)
}
func BenchmarkGlob_5_5_5_5_5(b *testing.B) {
	benchmarkGlob(b, 5, 5, 5, 5, 5)
}

func BenchmarkGlob_50_50(b *testing.B) {
	benchmarkGlob(b, 50, 50)
}

func BenchmarkGlob_100_50(b *testing.B) {
	benchmarkGlob(b, 100, 50)
}

func createDirs(tb testing.TB, dir string, dirs []int, filePrefixes []string, files int) {
	if len(dirs) == 0 {
		createFiles(tb, dir, filePrefixes, files)
		return
	}
	for i := 0; i < dirs[0]; i++ {
		dir := Join(dir, strconv.Itoa(i))
		createDirs(tb, dir, dirs[1:], filePrefixes, files)
	}
}

func createFiles(tb testing.TB, dir string, prefixes []string, count int) {
	if err := os.MkdirAll(dir, 0777); err != nil {
		tb.Fatal(err)
	}

	for _, prefix := range prefixes {
		for j := 0; j < count; j++ {
			filename := Join(dir, prefix+strconv.Itoa(j))
			if err := ioutil.WriteFile(filename, nil, 0666); err != nil {
				tb.Fatal(err)
			}
		}
	}
}
