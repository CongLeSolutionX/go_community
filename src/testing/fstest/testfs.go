// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fstest

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"sort"
	"strings"
	"testing/iotest"
)

// TestFS tests a file system implementation.
// It walks the entire tree of files in fsys,
// opening and checking that each file behaves correctly.
// TestFS considers it a misbehavior if the file system walk
// does not find the files in the expect list.
//
// If TestFS finds any misbehaviors, it returns an error reporting all of them.
// The error text spans multiple lines, one per detected misbehavior.
//
// Typical usage inside a test is:
//
//	if err := fstest.TestFS(myFS, "file/that/should/be/present"); err != nil {
//		t.Fatal(err)
//	}
//
func TestFS(fsys fs.FS, expect ...string) error {
	t := fsTester{fsys: fsys}
	t.checkDir(".")
	t.checkOpen(".")
	found := make(map[string]bool)
	for _, dir := range t.dirs {
		found[dir] = true
	}
	for _, file := range t.files {
		found[file] = true
	}
	for _, name := range expect {
		if !found[name] {
			t.errorf("%s: expected but did not visit", name)
		}
	}
	if len(t.errText) == 0 {
		return nil
	}
	return errors.New("TestFS found errors:\n" + string(t.errText))
}

// An fsTester holds state for running the test.
type fsTester struct {
	fsys    fs.FS
	errText []byte
	dirs    []string
	files   []string
}

// errorf adds an error line to errText.
func (t *fsTester) errorf(format string, args ...interface{}) {
	if len(t.errText) > 0 {
		t.errText = append(t.errText, '\n')
	}
	t.errText = append(t.errText, fmt.Sprintf(format, args...)...)
}

func (t *fsTester) openDir(dir string) fs.ReadDirFile {
	f, err := t.fsys.Open(dir)
	if err != nil {
		t.errorf("%s: Open: %v", dir, err)
		return nil
	}
	d, ok := f.(fs.ReadDirFile)
	if !ok {
		f.Close()
		t.errorf("%s: Open returned File type %T, not a io.ReadDirFile", dir, f)
		return nil
	}
	return d
}

// checkDir checks the directory dir, which is expected to exist
// (it is either the root or was found in a directory listing with IsDir true).
func (t *fsTester) checkDir(dir string) {
	// Read entire directory.
	t.dirs = append(t.dirs, dir)
	d := t.openDir(dir)
	if d == nil {
		return
	}
	list, err := d.ReadDir(-1)
	if err != nil {
		d.Close()
		t.errorf("%s: ReadDir(-1): %v", dir, err)
		return
	}

	// Check all children.
	var prefix string
	if dir == "." {
		prefix = ""
	} else {
		prefix = dir + "/"
	}
	for _, info := range list {
		if strings.Contains(info.Name(), "/") {
			t.errorf("%s: ReadDir: child name contains slash: %s", dir, info.Name())
			continue
		}
		path := prefix + info.Name()
		t.checkStat(path, info)
		t.checkOpen(path)
		if info.IsDir() {
			t.checkDir(path)
		} else {
			t.checkFile(path)
		}
	}

	// Check ReadDir(-1) at EOF.
	list2, err := d.ReadDir(-1)
	if len(list2) > 0 || err != nil {
		d.Close()
		t.errorf("%s: ReadDir(-1) at EOF = %d entries, %v, wanted 0 entries, nil", dir, len(list2), err)
		return
	}

	// Check ReadDir(1) at EOF (different results).
	list2, err = d.ReadDir(1)
	if len(list2) > 0 || err != io.EOF {
		d.Close()
		t.errorf("%s: ReadDir(1) at EOF = %d entries, %v, wanted 0 entries, EOF", dir, len(list2), err)
		return
	}

	// Check that close does not report an error.
	if err := d.Close(); err != nil {
		t.errorf("%s: Close: %v", dir, err)
	}

	// Check that closing twice doesn't crash.
	// The return value doesn't matter.
	d.Close()

	// Reopen directory, read a second time, make sure contents match.
	if d = t.openDir(dir); d == nil {
		return
	}
	defer d.Close()
	list2, err = d.ReadDir(-1)
	if err != nil {
		t.errorf("%s: second Open+ReadDir(-1): %v", dir, err)
		return
	}
	t.checkDirList(dir, "first Open+ReadDir(-1) vs second Open+ReadDir(-1)", list, list2)

	// Reopen directory, read a third time in pieces, make sure contents match.
	if d = t.openDir(dir); d == nil {
		return
	}
	defer d.Close()
	list2 = nil
	for {
		n := 1
		if len(list2) > 0 {
			n = 2
		}
		frag, err := d.ReadDir(n)
		if len(frag) > n {
			t.errorf("%s: third Open: ReadDir(%d) after %d: %d entries (too many)", dir, n, len(list2), len(frag))
			return
		}
		list2 = append(list2, frag...)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.errorf("%s: third Open: ReadDir(%d) after %d: %v", dir, n, len(list2), err)
			return
		}
		if n == 0 {
			t.errorf("%s: third Open: ReadDir(%d) after %d: 0 entries but nil error", dir, n, len(list2))
			return
		}
	}
	t.checkDirList(dir, "first Open+ReadDir(-1) vs third Open+ReadDir(1,2) loop", list, list2)
}

// formatInfo formats an fs.FileInfo into a string for error messages and comparison.
func formatInfo(info fs.FileInfo) string {
	return fmt.Sprintf("%s Size=%d IsDir=%v Mode=%v ModTime=%v", info.Name(), info.Size(), info.IsDir(), info.Mode(), info.ModTime())
}

// checkStat checks that a direct stat of path matches info,
// which was found in the parent's directory listing.
func (t *fsTester) checkStat(path string, info fs.FileInfo) {
	file, err := t.fsys.Open(path)
	if err != nil {
		t.errorf("%s: Open: %v", path, err)
		return
	}
	info2, err := file.Stat()
	file.Close()
	if err != nil {
		t.errorf("%s: Stat: %v", path, err)
		return
	}
	finfo := formatInfo(info)
	finfo2 := formatInfo(info2)
	if finfo2 != finfo {
		t.errorf("%s: file.Stat() = %v\n\twant %v", path, finfo2, finfo)
	}
}

// checkDirList checks that two directory lists contain the same files and file info.
// The order of the lists need not match.
func (t *fsTester) checkDirList(dir, desc string, list1, list2 []fs.FileInfo) {
	old := make(map[string]fs.FileInfo)
	checkMode := func(info fs.FileInfo) {
		if info.IsDir() != (info.Mode()&fs.ModeDir != 0) {
			if info.IsDir() {
				t.errorf("%s: ReadDir returned %s with IsDir() = true, Mode() & ModeDir = 0", dir, info.Name())
			} else {
				t.errorf("%s: ReadDir returned %s with IsDir() = false, Mode() & ModeDir = ModeDir", dir, info.Name())
			}
		}
	}

	for _, info1 := range list1 {
		old[info1.Name()] = info1
		checkMode(info1)
	}

	var diffs []string
	for _, info2 := range list2 {
		info1 := old[info2.Name()]
		if info1 == nil {
			checkMode(info2)
			diffs = append(diffs, "+ "+formatInfo(info2))
			continue
		}
		if formatInfo(info1) != formatInfo(info2) {
			diffs = append(diffs, "- "+formatInfo(info1), "+ "+formatInfo(info2))
		}
		delete(old, info2.Name())
	}
	for _, info1 := range old {
		diffs = append(diffs, "- "+formatInfo(info1))
	}

	if len(diffs) == 0 {
		return
	}

	sort.Slice(diffs, func(i, j int) bool {
		fi := strings.Fields(diffs[i])
		fj := strings.Fields(diffs[j])
		// sort by name (i < j) and then +/- (j < i, because + < -)
		return fi[1]+" "+fj[0] < fj[1]+" "+fi[0]
	})

	t.errorf("%s: diff %s:\n\t%s", dir, desc, strings.Join(diffs, "\n\t"))
}

// checkFile checks that basic file reading works correctly.
func (t *fsTester) checkFile(file string) {
	t.files = append(t.files, file)

	// Read entire file.
	f, err := t.fsys.Open(file)
	if err != nil {
		t.errorf("%s: Open: %v", file, err)
		return
	}

	data, err := ioutil.ReadAll(f)
	if err != nil {
		f.Close()
		t.errorf("%s: Open+ReadAll: %v", file, err)
		return
	}

	if err := f.Close(); err != nil {
		t.errorf("%s: Close: %v", file, err)
	}

	// Check that closing twice doesn't crash.
	// The return value doesn't matter.
	f.Close()

	// Use iotest.TestReader to check small reads, Seek, ReadAt.
	f, err = t.fsys.Open(file)
	if err != nil {
		t.errorf("%s: second Open: %v", file, err)
		return
	}
	defer f.Close()
	if err := iotest.TestReader(f, data); err != nil {
		t.errorf("%s: failed TestReader:\n\t%s", file, strings.ReplaceAll(err.Error(), "\n", "\n\t"))
	}
}

func (t *fsTester) checkFileRead(file, desc string, data1, data2 []byte) {
	if string(data1) != string(data2) {
		t.errorf("%s: %s: different data returned\n\t%q\n\t%q", file, desc, data1, data2)
		return
	}
}

// checkOpen checks that various invalid forms of file's name cannot be opened.
func (t *fsTester) checkOpen(file string) {
	bad := []string{
		"/" + file,
		file + "/.",
	}
	if file == "." {
		bad = append(bad, "/")
	}
	if i := strings.Index(file, "/"); i >= 0 {
		bad = append(bad,
			file[:i]+"//"+file[i+1:],
			file[:i]+"/./"+file[i+1:],
			file[:i]+`\`+file[i+1:],
			file[:i]+"/../"+file,
		)
	}
	if i := strings.LastIndex(file, "/"); i >= 0 {
		bad = append(bad,
			file[:i]+"//"+file[i+1:],
			file[:i]+"/./"+file[i+1:],
			file[:i]+`\`+file[i+1:],
			file+"/../"+file[i+1:],
		)
	}

	for _, b := range bad {
		if f, err := t.fsys.Open(b); err == nil {
			f.Close()
			t.errorf("%s: Open(%s) succeeded, want error", file, b)
		}
	}
}
