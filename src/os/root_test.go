package os_test

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
)

// testMaybeRooted calls f in two subtests,
// one with a Root and one with a nil r.
func testMaybeRooted(t *testing.T, f func(t *testing.T, r *os.Root)) {
	t.Run("NoRoot", func(t *testing.T) {
		t.Chdir(t.TempDir())
		f(t, nil)
	})
	t.Run("InRoot", func(t *testing.T) {
		t.Chdir(t.TempDir())
		r, err := os.OpenRoot(".")
		if err != nil {
			t.Fatal(err)
		}
		defer r.Close()
		f(t, r)
	})
}

// makefs creates a test filesystem layout and returns the path to its root.
//
// Each entry in the slice is a file, directory, or symbolic link to create:
//
//   - "d/": directory d
//   - "f": file f with contents f
//   - "a => b": symlink a with target b
//
// The directory containing the filesystem is always named ROOT.
// $ABS is replaced with the absolute path of the directory containing the filesystem.
//
// Parent directories are automatically created as needed.
//
// makefs calls t.Skip if the layout contains features not supported by the current GOOS.
func makefs(t *testing.T, fs []string) string {
	root := path.Join(t.TempDir(), "ROOT")
	if err := os.Mkdir(root, 0777); err != nil {
		t.Fatal(err)
	}
	for _, ent := range fs {
		ent = strings.ReplaceAll(ent, "$ABS", root)
		base, link, isLink := strings.Cut(ent, " => ")
		if isLink {
			if runtime.GOOS == "wasip1" && path.IsAbs(link) {
				t.Skip("absolute link targets not supported on " + runtime.GOOS)
			}
			if runtime.GOOS == "plan9" {
				t.Skip("symlinks not supported on " + runtime.GOOS)
			}
			ent = base
		}
		if err := os.MkdirAll(path.Join(root, path.Dir(base)), 0777); err != nil {
			t.Fatal(err)
		}
		if isLink {
			if err := os.Symlink(link, path.Join(root, base)); err != nil {
				t.Fatal(err)
			}
		} else if strings.HasSuffix(ent, "/") {
			if err := os.MkdirAll(path.Join(root, ent), 0777); err != nil {
				t.Fatal(err)
			}
		} else {
			if err := os.WriteFile(path.Join(root, ent), []byte(ent), 0666); err != nil {
				t.Fatal(err)
			}
		}
	}
	return root
}

// A rootTest is a test case for os.Root.
type rootTest struct {
	name string

	// fs is the test filesystem layout. See makefs above.
	fs []string

	// open is the filename to access in the test.
	open string

	// target is the filename that we expect to be accessed, after resolving all symlinks.
	// For test cases where the operation fails due to an escaping path such as ../ROOT/x,
	// the target is the filename that should not have been opened.
	target string

	// wantError is true if accessing the file should fail.
	wantError bool

	// alwaysFails is true if the open operation is expected to fail
	// even when using non-openat operations.
	//
	// This lets us check that tests that are expected to fail because (for example)
	// a path escapes the directory root will succeed when the escaping checks are not
	// performed.
	alwaysFails bool
}

// run sets up the test filesystem layout, os.OpenDirs the root, and calls f.
func (test *rootTest) run(t *testing.T, f func(t *testing.T, target string, d *os.Root)) {
	t.Run(test.name, func(t *testing.T) {
		root := makefs(t, test.fs)
		d, err := os.OpenRoot(root)
		if err != nil {
			t.Fatal(err)
		}
		defer d.Close()
		target := test.target
		if test.target != "" {
			target = filepath.Join(root, test.target)
		}
		f(t, target, d)
	})
}

// checkErr checks the error result of a test, verifying that it succeeded or failed as expected.
func (test *rootTest) checkErr(t *testing.T, err error, format string, args ...any) bool {
	t.Helper()
	if test.wantError {
		if err == nil {
			op := fmt.Sprintf(format, args...)
			t.Fatalf("%v = nil; want error", op)
		}
		return true
	} else {
		if err != nil {
			op := fmt.Sprintf(format, args...)
			t.Fatalf("%v = %v; want success", op, err)
		}
		return false
	}
}

var rootTestCases = []rootTest{{
	name:   "plain path",
	fs:     []string{},
	open:   "target",
	target: "target",
}, {
	name: "path in directory",
	fs: []string{
		"a/b/c/",
	},
	open:   "a/b/c/target",
	target: "a/b/c/target",
}, {
	name: "symlink",
	fs: []string{
		"link => target",
	},
	open:   "link",
	target: "target",
}, {
	name: "symlink chain",
	fs: []string{
		"link => a/b/c/target",
		"a/b => e",
		"a/e => ../f",
		"f => g/h/i",
		"g/h/i => ..",
		"g/c/",
	},
	open:   "link",
	target: "g/c/target",
}, {
	name: "path with dot",
	fs: []string{
		"a/b/",
	},
	open:   "./a/./b/./target",
	target: "a/b/target",
}, {
	name: "path with dotdot",
	fs: []string{
		"a/b/",
	},
	open:   "a/../a/b/../../a/b/../b/target",
	target: "a/b/target",
}, {
	name: "dotdot no symlink",
	fs: []string{
		"a/",
	},
	open:   "a/../target",
	target: "target",
}, {
	name: "dotdot after symlink",
	fs: []string{
		"a => b/c",
		"b/c/",
	},
	open: "a/../target",
	target: func() string {
		if runtime.GOOS == "windows" {
			// On Windows, the path is cleaned before symlink resolution.
			return "target"
		}
		return "b/target"
	}(),
}, {
	name: "dotdot before symlink",
	fs: []string{
		"a => b/c",
		"b/c/",
	},
	open:   "b/../a/target",
	target: "b/c/target",
}, {
	name:        "directory does not exist",
	fs:          []string{},
	open:        "a/file",
	wantError:   true,
	alwaysFails: true,
}, {
	name:        "empty path",
	fs:          []string{},
	open:        "",
	wantError:   true,
	alwaysFails: true,
}, {
	name: "symlink cycle",
	fs: []string{
		"a => a",
	},
	open:        "a",
	wantError:   true,
	alwaysFails: true,
}, {
	name:      "path escapes",
	fs:        []string{},
	open:      "../ROOT/target",
	target:    "target",
	wantError: true,
}, {
	name: "long path escapes",
	fs: []string{
		"a/",
	},
	open:      "a/../../ROOT/target",
	target:    "target",
	wantError: true,
}, {
	name: "absolute symlink",
	fs: []string{
		"link => $ABS/target",
	},
	open:      "link",
	target:    "target",
	wantError: true,
}, {
	name: "relative symlink",
	fs: []string{
		"link => ../ROOT/target",
	},
	open:      "link",
	target:    "target",
	wantError: true,
}, {
	name: "symlink chain escapes",
	fs: []string{
		"link => a/b/c/target",
		"a/b => e",
		"a/e => ../../ROOT",
		"c/",
	},
	open:      "link",
	target:    "c/target",
	wantError: true,
}}

func TestRootOpen_File(t *testing.T) {
	want := []byte("target")
	for _, test := range rootTestCases {
		test.run(t, func(t *testing.T, target string, root *os.Root) {
			if target != "" {
				if err := os.WriteFile(target, want, 0666); err != nil {
					t.Fatal(err)
				}
			}
			f, err := root.Open(test.open)
			if test.checkErr(t, err, "root.Open(%q)", test.open) {
				return
			}
			defer f.Close()
			got, err := io.ReadAll(f)
			if err != nil || !bytes.Equal(got, want) {
				t.Errorf(`Dir.Open(%q): read content %q, %v; want %q`, test.open, string(got), err, string(want))
			}
		})
	}
}

func TestRootOpen_Directory(t *testing.T) {
	for _, test := range rootTestCases {
		test.run(t, func(t *testing.T, target string, root *os.Root) {
			if target != "" {
				if err := os.Mkdir(target, 0777); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(target+"/found", nil, 0666); err != nil {
					t.Fatal(err)
				}
			}
			f, err := root.Open(test.open)
			if test.checkErr(t, err, "root.Open(%q)", test.open) {
				return
			}
			defer f.Close()
			got, err := f.Readdirnames(-1)
			if err != nil {
				t.Errorf(`Dir.Open(%q).Readdirnames: %v`, test.open, err)
			}
			if want := []string{"found"}; !slices.Equal(got, want) {
				t.Errorf(`Dir.Open(%q).Readdirnames: %q, want %q`, test.open, got, want)
			}
		})
	}
}

func TestRootCreate(t *testing.T) {
	want := []byte("target")
	for _, test := range rootTestCases {
		test.run(t, func(t *testing.T, target string, root *os.Root) {
			f, err := root.Create(test.open)
			if test.checkErr(t, err, "root.Create(%q)", test.open) {
				return
			}
			if _, err := f.Write(want); err != nil {
				t.Fatal(err)
			}
			f.Close()
			got, err := os.ReadFile(target)
			if err != nil {
				t.Fatalf(`reading file created with root.Create(%q): %v`, test.open, err)
			}
			if !bytes.Equal(got, want) {
				t.Fatalf(`reading file created with root.Create(%q): got %q; want %q`, test.open, got, want)
			}
		})
	}
}

func TestRootMkdir(t *testing.T) {
	for _, test := range rootTestCases {
		test.run(t, func(t *testing.T, target string, root *os.Root) {
			if !test.wantError {
				fi, err := os.Lstat(filepath.Join(root.Name(), test.open))
				if err == nil && fi.Mode().Type() == fs.ModeSymlink {
					// This case is trying to mkdir("some symlink"),
					// which is an error.
					test.wantError = true
				}
			}

			err := root.Mkdir(test.open, 0777)
			if test.checkErr(t, err, "root.Create(%q)", test.open) {
				return
			}
			fi, err := os.Lstat(target)
			if err != nil {
				t.Fatalf(`stat file created with Root.Mkdir(%q): %v`, test.open, err)
			}
			if !fi.IsDir() {
				t.Fatalf(`stat file created with Root.Mkdir(%q): not a directory`, test.open)
			}
		})
	}
}

func TestRootOpenRoot(t *testing.T) {
	for _, test := range rootTestCases {
		test.run(t, func(t *testing.T, target string, root *os.Root) {
			if target != "" {
				if err := os.Mkdir(target, 0777); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(target+"/f", nil, 0666); err != nil {
					t.Fatal(err)
				}
			}
			rr, err := root.OpenRoot(test.open)
			if test.checkErr(t, err, "root.OpenRoot(%q)", test.open) {
				return
			}
			defer rr.Close()
			f, err := rr.Open("f")
			if err != nil {
				t.Fatalf(`root.OpenRoot(%q).Open("f") = %v`, test.open, err)
			}
			f.Close()
		})
	}
}

// A rootConsistencyTest is a test case comparing os.Root behavior with
// the corresponding non-Root function.
//
// These tests verify that, for example, Root.Open("file/./") and os.Open("file/./")
// have the same result, although the specific result may vary by platform.
type rootConsistencyTest struct {
	name string

	// fs is the test filesystem layout. See makefs above.
	fs []string

	// open is the filename to access in the test.
	open string
}

var rootConsistencyTestCases = []rootConsistencyTest{{
	name: "dir slash dot",
	fs: []string{
		"target/file",
	},
	open: "target/.",
}, {
	name: "dot",
	fs: []string{
		"file",
	},
	open: ".",
}, {
	name: "file slash dot",
	fs: []string{
		"target",
	},
	open: "target/.",
}, {
	name: "dir slash",
	fs: []string{
		"target/file",
	},
	open: "target/",
}, {
	name: "dot slash",
	fs: []string{
		"file",
	},
	open: "./",
}, {
	name: "file slash",
	fs: []string{
		"target",
	},
	open: "target/",
}, {
	name: "symlink slash",
	fs: []string{
		"target/file",
		"link => target",
	},
	open: "link/",
}, {
	name: "unresolved symlink",
	fs: []string{
		"link => target",
	},
	open: "link",
}, {
	name: "resolved symlink",
	fs: []string{
		"link => target",
		"target",
	},
	open: "link",
}, {
	name: "dotdot in path after symlink",
	fs: []string{
		"a => b/c",
		"b/c/",
		"b/target",
	},
	open: "a/../target",
}}

func (test rootConsistencyTest) run(t *testing.T, f func(t *testing.T, path string, r *os.Root) (string, error)) {
	if runtime.GOOS == "wasip1" {
		// On wasip, non-Root functions clean paths before opening them,
		// resulting in inconsistent behavior.
		// https://go.dev/issue/69509
		t.Skip("#69509: inconsistent results on wasip1")
	}

	t.Run(test.name, func(t *testing.T) {
		dir1 := makefs(t, test.fs)
		dir2 := makefs(t, test.fs)

		r, err := os.OpenRoot(dir1)
		if err != nil {
			t.Fatal(err)
		}
		defer r.Close()

		res1, err1 := f(t, test.open, r)
		res2, err2 := f(t, dir2+"/"+test.open, nil)

		if res1 != res2 || ((err1 == nil) != (err2 == nil)) {
			t.Errorf("with root:    res=%v", res1)
			t.Errorf("              err=%v", err1)
			t.Errorf("without root: res=%v", res2)
			t.Errorf("              err=%v", err2)
			t.Errorf("want consistent results, got mismatch")
		}
	})
}

func TestRootConsistencyOpen(t *testing.T) {
	for _, test := range rootConsistencyTestCases {
		test.run(t, func(t *testing.T, path string, r *os.Root) (string, error) {
			var f *os.File
			var err error
			if r == nil {
				f, err = os.Open(path)
			} else {
				f, err = r.Open(path)
			}
			if err != nil {
				return "", err
			}
			defer f.Close()
			fi, err := f.Stat()
			if err == nil && !fi.IsDir() {
				b, err := io.ReadAll(f)
				return string(b), err
			} else {
				names, err := f.Readdirnames(-1)
				slices.Sort(names)
				return fmt.Sprintf("%q", names), err
			}
		})
	}
}

func TestRootConsistencyCreate(t *testing.T) {
	for _, test := range rootConsistencyTestCases {
		test.run(t, func(t *testing.T, path string, r *os.Root) (string, error) {
			var f *os.File
			var err error
			if r == nil {
				f, err = os.Create(path)
			} else {
				f, err = r.Create(path)
			}
			if err == nil {
				f.Write([]byte("file contents"))
				f.Close()
			}
			return "", err
		})
	}
}

func TestRootConsistencyMkdir(t *testing.T) {
	for _, test := range rootConsistencyTestCases {
		test.run(t, func(t *testing.T, path string, r *os.Root) (string, error) {
			var err error
			if r == nil {
				err = os.Mkdir(path, 0777)
			} else {
				err = r.Mkdir(path, 0777)
			}
			return "", err
		})
	}
}

func TestRootRenameAfterOpen(t *testing.T) {
	switch runtime.GOOS {
	case "windows":
		t.Skip("renaming open files not supported on " + runtime.GOOS)
	case "js", "plan9":
		t.Skip("openat not supported on " + runtime.GOOS)
	}

	dir := t.TempDir()

	// Create directory "a" and open it.
	if err := os.Mkdir(filepath.Join(dir, "a"), 0777); err != nil {
		t.Fatal(err)
	}
	dirf, err := os.OpenRoot(filepath.Join(dir, "a"))
	if err != nil {
		t.Fatal(err)
	}
	defer dirf.Close()

	// Rename "a" => "b", and create "b/f".
	if err := os.Rename(filepath.Join(dir, "a"), filepath.Join(dir, "b")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b/f"), []byte("hello"), 0666); err != nil {
		t.Fatal(err)
	}

	// Open "f", and confirm that we see it.
	f, err := dirf.OpenFile("f", os.O_RDONLY, 0)
	if err != nil {
		t.Fatalf("reading file after renaming parent: %v", err)
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(b), "hello"; got != want {
		t.Fatalf("file contents: %q, want %q", got, want)
	}

	// f.Name reflects the original path we opened the directory under (".../a"), not "b".
	if got, want := f.Name(), dirf.Name()+string(os.PathSeparator)+"f"; got != want {
		t.Errorf("f.Name() = %q, want %q", got, want)
	}
}

func TestRootNonPermissionMode(t *testing.T) {
	r, err := os.OpenRoot(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.OpenFile("file", os.O_RDWR|os.O_CREATE, 01777); err == nil {
		t.Errorf("r.OpenFile(file, O_RDWR|O_CREATE, 01777) succeeded; want error")
	}
	if err := r.Mkdir("file", 01777); err == nil {
		t.Errorf("r.Mkdir(file, 01777) succeeded; want error")
	}
}
