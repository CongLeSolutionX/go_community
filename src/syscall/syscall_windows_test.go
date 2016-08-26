// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syscall_test

import (
	"fmt"
	"internal/syscall/windows"
	"internal/testenv"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"unicode/utf16"
	"unsafe"
)

var supportJunctionLinks = true

func init() {
	b, _ := exec.Command("cmd", "/c", "mklink", "/?").Output()
	if !strings.Contains(string(b), " /J ") {
		supportJunctionLinks = false
	}
}

func TestWin32finddata(t *testing.T) {
	dir, err := ioutil.TempDir("", "go-build")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "long_name.and_extension")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create %v: %v", path, err)
	}
	f.Close()

	type X struct {
		fd  syscall.Win32finddata
		got byte
		pad [10]byte // to protect ourselves

	}
	var want byte = 2 // it is unlikely to have this character in the filename
	x := X{got: want}

	pathp, _ := syscall.UTF16PtrFromString(path)
	h, err := syscall.FindFirstFile(pathp, &(x.fd))
	if err != nil {
		t.Fatalf("FindFirstFile failed: %v", err)
	}
	err = syscall.FindClose(h)
	if err != nil {
		t.Fatalf("FindClose failed: %v", err)
	}

	if x.got != want {
		t.Fatalf("memory corruption: want=%d got=%d", want, x.got)
	}
}

func abort(funcname string, err error) {
	panic(funcname + " failed: " + err.Error())
}

func ExampleLoadLibrary() {
	h, err := syscall.LoadLibrary("kernel32.dll")
	if err != nil {
		abort("LoadLibrary", err)
	}
	defer syscall.FreeLibrary(h)
	proc, err := syscall.GetProcAddress(h, "GetVersion")
	if err != nil {
		abort("GetProcAddress", err)
	}
	r, _, _ := syscall.Syscall(uintptr(proc), 0, 0, 0, 0)
	major := byte(r)
	minor := uint8(r >> 8)
	build := uint16(r >> 16)
	print("windows version ", major, ".", minor, " (Build ", build, ")\n")
}

func runWithEnablePrivilege(privName string, fn func()) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	err := windows.ImpersonateSelf(windows.SecurityImpersonation)
	if err != nil {
		return err
	}
	defer windows.RevertToSelf()

	th, err := windows.GetCurrentThread()
	if err != nil {
		return err
	}

	var t syscall.Token

	err = windows.OpenThreadToken(th, syscall.TOKEN_ADJUST_PRIVILEGES|syscall.TOKEN_QUERY, false, &t)
	if err != nil {
		return err
	}
	defer syscall.CloseHandle(syscall.Handle(t))

	tp := windows.TOKEN_PRIVILEGES{
		PrivilegeCount: 1,
	}

	p, err := syscall.UTF16PtrFromString(privName)
	if err != nil {
		return err
	}

	err = windows.LookupPrivilegeValue(nil, p, &tp.Privileges[0].Luid)
	if err != nil {
		return err
	}

	tp.Privileges[0].Attributes = windows.SE_PRIVILEGE_ENABLED

	err = windows.AdjustTokenPrivileges(t, false, &tp, 0, nil, nil)
	if err != nil {
		return err
	}

	fn()

	return nil
}

func createReparsePoint(name string, rdbuf []byte, isDir bool) error {
	if isDir {
		err := os.Mkdir(name, 0777)
		if err != nil {
			return err
		}
	} else {
		f, err := os.Create(name)
		if err != nil {
			return err
		}
		err = f.Close()
		if err != nil {
			return err
		}
	}

	namep, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return err
	}

	fd, err := syscall.CreateFile(namep, syscall.GENERIC_WRITE, 0, nil, syscall.OPEN_EXISTING, syscall.FILE_FLAG_OPEN_REPARSE_POINT|syscall.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		return err
	}
	defer syscall.CloseHandle(fd)

	var bytesReturned uint32

	return syscall.DeviceIoControl(fd, windows.FSCTL_SET_REPARSE_POINT, &rdbuf[0], uint32(len(rdbuf)), nil, 0, &bytesReturned, nil)
}

var readlinkSymlinkTests = []struct {
	SubstituteName string
	PrintName      string
	PathBuffer     string

	Want string
}{
	// relative paths
	{
		"target",
		"target",
		"{{subName}}{{printName}}",

		"target",
	},
	{
		"target",
		"target",
		"{{subName}}\x00{{printName}}",

		"target",
	},
	{
		"target",
		"target",
		"{{subName}}\x00\x00{{printName}}",

		"target",
	},
	{
		"target",
		"target",
		"\x00{{subName}}\x00{{printName}}\x00",

		"target",
	},
	{
		"target",
		"target",
		"abc{{subName}}defg{{printName}}hijklmn",

		"target",
	},
	{
		"target",
		"target",
		"{{printName}}{{subName}}",

		"target",
	},
	{
		"target",
		"target",
		"{{printName}}\x00{{subName}}",

		"target",
	},
	{
		"target",
		"target",
		"{{printName}}\x00\x00{{subName}}",

		"target",
	},
	{
		"target",
		"target",
		"\x00{{printName}}\x00{{subName}}\x00",

		"target",
	},
	{
		"target",
		"target",
		"abc{{printName}}defg{{subName}}hijklmn",

		"target",
	},

	// absolute paths
	{
		`\??\{{tmp}}\target`,
		`{{tmp}}\target`,
		"{{subName}}{{printName}}",

		`{{tmp}}\target`,
	},
	{
		`\??\{{tmp}}\target`,
		`{{tmp}}\target`,
		"{{printName}}{{subName}}",

		`{{tmp}}\target`,
	},

	// TODO https://github.com/golang/go/issues/16145
	// paths without printName
	// {
	// 	`target`,
	// 	``,
	// 	"{{subName}}{{printName}}",

	// 	`target`,
	// },
	// {
	// 	`\??\{{tmp}}\target`,
	// 	``,
	// 	"{{subName}}{{printName}}",

	// 	`{{tmp}}\target`,
	// },
}

func makeSymlinkReparseDataBuffer(subName, printName, pathBuffer string) ([]byte, error) {
	rdbuf := make([]byte, syscall.MAXIMUM_REPARSE_DATA_BUFFER_SIZE)

	rdb := (*windows.SymbolicLinkReparseDataBuffer)(unsafe.Pointer(&rdbuf[0]))

	subIndex := strings.Index(pathBuffer, "{{subName}}")
	printIndex := strings.Index(pathBuffer, "{{printName}}")

	switch {
	case subIndex < printIndex && subIndex != -1:
		pathBuffer = strings.Replace(pathBuffer, "{{subName}}", subName, 1)
		printIndex = strings.Index(pathBuffer, "{{printName}}")
		pathBuffer = strings.Replace(pathBuffer, "{{printName}}", printName, 1)
	case printIndex < subIndex && printIndex != -1:
		pathBuffer = strings.Replace(pathBuffer, "{{printName}}", printName, 1)
		subIndex = strings.Index(pathBuffer, "{{subName}}")
		pathBuffer = strings.Replace(pathBuffer, "{{subName}}", subName, 1)
	default:
		return nil, fmt.Errorf("unsupported symlink format: %s", pathBuffer)
	}

	rs := []rune(pathBuffer)

	if len(rs) != len(pathBuffer) {
		return nil, fmt.Errorf("multi bytes chars are not supported on this test: %s", pathBuffer)
	}

	ws := utf16.Encode(rs)

	rdb.ReparseTag = syscall.IO_REPARSE_TAG_SYMLINK
	rdb.ReparseDataLength = uint16(12 + len(pathBuffer)*2)
	rdb.SubstituteNameOffset = uint16(subIndex * 2)
	rdb.SubstituteNameLength = uint16(len(subName) * 2)
	rdb.PrintNameOffset = uint16(printIndex * 2)
	rdb.PrintNameLength = uint16(len(printName) * 2)

	if !strings.HasPrefix(subName, `\??\`) {
		rdb.Flags = windows.SYMLINK_FLAG_RELATIVE
	}

	copy((*(*[0xffff]uint16)(unsafe.Pointer(&rdb.PathBuffer[0])))[:], ws)

	return rdbuf[:8+rdb.ReparseDataLength], nil
}

func TestReadlinkSymlink(t *testing.T) {
	testenv.MustHaveSymlink(t)

	tmp, err := ioutil.TempDir("", "TestReadlinkSymlink")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	err = os.Chdir(tmp)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(wd)

	// test symlinks to the file
	err = runWithEnablePrivilege(windows.SE_CREATE_SYMBOLIC_LINK_NAME, func() {
		f, err := os.Create("target")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove("target")

		_, err = f.Write([]byte("abcdefghijklmn"))
		if err != nil {
			t.Fatal(err)
		}

		err = f.Close()
		if err != nil {
			t.Fatal(err)
		}

		for _, test := range readlinkSymlinkTests {
			subName := strings.Replace(test.SubstituteName, "{{tmp}}", tmp, 1)
			printName := strings.Replace(test.PrintName, "{{tmp}}", tmp, 1)
			want := strings.Replace(test.Want, "{{tmp}}", tmp, 1)

			rdbuf, err := makeSymlinkReparseDataBuffer(subName, printName, test.PathBuffer)
			if err != nil {
				t.Errorf("cannot make the symlink reparse data buffer: %v\n", err)

				continue
			}

			{
				err = createReparsePoint("link", rdbuf, false)
				if err != nil {
					t.Errorf("cannot create the reparse point: %v\n", err)

					goto clean
				}

				bs, err := ioutil.ReadFile("link")
				if err != nil {
					t.Errorf("cannot read from the link: %v\n", err)

					goto clean
				}

				if string(bs) != "abcdefghijklmn" {
					t.Errorf("contents are mismatch: expected %q, got %q\n", "abcdefghijklmn", string(bs))

					goto clean
				}

				readlink, err := os.Readlink("link")
				if err != nil {
					t.Errorf("cannot readlink from the link: %v\n", err)

					goto clean
				}

				if readlink != want {
					t.Errorf("readlink is mismatch: expected %q, got %q\n", want, readlink)

					goto clean
				}
			}

		clean:
			os.Remove("link")
		}
	})
	if err != nil {
		t.Fatal(err)
	}

	// test symlinks to the directory
	err = runWithEnablePrivilege(windows.SE_CREATE_SYMBOLIC_LINK_NAME, func() {
		err := os.Mkdir("target", 0777)
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll("target")

		files := []string{"abc", "def", "ghi"}

		for _, file := range files {
			f, err := os.Create(filepath.Join("target", file))
			if err != nil {
				t.Fatal(err)
			}
			err = f.Close()
			if err != nil {
				t.Fatal(err)
			}
		}

		for _, test := range readlinkSymlinkTests {
			subName := strings.Replace(test.SubstituteName, "{{tmp}}", tmp, 1)
			printName := strings.Replace(test.PrintName, "{{tmp}}", tmp, 1)
			want := strings.Replace(test.Want, "{{tmp}}", tmp, 1)

			rdbuf, err := makeSymlinkReparseDataBuffer(subName, printName, test.PathBuffer)
			if err != nil {
				t.Errorf("cannot make the symlink reparse data buffer: %v\n", err)

				continue
			}

			{
				err = createReparsePoint("link", rdbuf, true)
				if err != nil {
					t.Errorf("cannot create the reparse point: %v\n", err)

					goto clean
				}

				fis, err := ioutil.ReadDir("link")
				if err != nil {
					t.Errorf("cannot readdir from the link: %v\n", err)

					goto clean
				}

				fnames := make([]string, len(fis))
				for i, fi := range fis {
					fnames[i] = fi.Name()
				}

				if len(fnames) != len(files) {
					t.Errorf("filenames are mismatch: expected %v, got %v\n", files, fnames)

					goto clean
				}

				for i := range fnames {
					if files[i] != fnames[i] {
						t.Errorf("filenames are mismatch: expected %v, got %v\n", files, fnames)

						goto clean
					}
				}

				readlink, err := os.Readlink("link")
				if err != nil {
					t.Errorf("cannot readlink from the link: %v\n", err)

					goto clean
				}

				if readlink != want {
					t.Errorf("readlink is mismatch: expected %q, got %q\n", want, readlink)

					goto clean
				}
			}

		clean:
			os.Remove("link")
		}
	})
	if err != nil {
		t.Fatal(err)
	}
}

// Suprisingly enough, the mount point allows only one format.
// Other formats are rejected with ERROR_INVALID_REPARSE_DATA_BUFFER.
// This is an undocumented behavior.
var readlinkJunctionTests = []struct {
	SubstituteName string
	PrintName      string
	PathBuffer     string

	Want string
}{
	// absolute paths
	{
		`\??\{{tmp}}\target`,
		`{{tmp}}\target`,
		"{{subName}}\x00{{printName}}\x00",

		`{{tmp}}\target`,
	},

	// TODO https://github.com/golang/go/issues/16145
	// absolute paths without printName
	// {
	// 	`\??\{{tmp}}\target`,
	// 	``,
	//  "{{subName}}\x00{{printName}}\x00",

	// 	`{{tmp}}\target`,
	// },
}

func makeMountPointReparseDataBuffer(subName, printName, pathBuffer string) ([]byte, error) {
	rdbuf := make([]byte, syscall.MAXIMUM_REPARSE_DATA_BUFFER_SIZE)

	rdb := (*windows.MountPointReparseDataBuffer)(unsafe.Pointer(&rdbuf[0]))

	subIndex := strings.Index(pathBuffer, "{{subName}}")
	printIndex := strings.Index(pathBuffer, "{{printName}}")

	switch {
	case subIndex < printIndex && subIndex != -1:
		pathBuffer = strings.Replace(pathBuffer, "{{subName}}", subName, 1)
		printIndex = strings.Index(pathBuffer, "{{printName}}")
		pathBuffer = strings.Replace(pathBuffer, "{{printName}}", printName, 1)
	case printIndex < subIndex && printIndex != -1:
		pathBuffer = strings.Replace(pathBuffer, "{{printName}}", printName, 1)
		subIndex = strings.Index(pathBuffer, "{{subName}}")
		pathBuffer = strings.Replace(pathBuffer, "{{subName}}", subName, 1)
	default:
		return nil, fmt.Errorf("unsupported mount point format: %s", pathBuffer)
	}

	rs := []rune(pathBuffer)

	if len(rs) != len(pathBuffer) {
		return nil, fmt.Errorf("multi bytes chars are not supported on this test: %s", pathBuffer)
	}

	ws := utf16.Encode(rs)

	rdb.ReparseTag = windows.IO_REPARSE_TAG_MOUNT_POINT
	rdb.ReparseDataLength = uint16(8 + len(pathBuffer)*2)
	rdb.SubstituteNameOffset = uint16(subIndex * 2)
	rdb.SubstituteNameLength = uint16(len(subName) * 2)
	rdb.PrintNameOffset = uint16(printIndex * 2)
	rdb.PrintNameLength = uint16(len(printName) * 2)

	if !strings.HasPrefix(subName, `\??\`) {
		return nil, fmt.Errorf("unsupported mount point format: %s", pathBuffer)
	}

	copy((*(*[0xffff]uint16)(unsafe.Pointer(&rdb.PathBuffer[0])))[:], ws)

	return rdbuf[:8+rdb.ReparseDataLength], nil
}

func TestReadlinkJunction(t *testing.T) {
	if !supportJunctionLinks {
		t.Skip("skipping because junction links are not supported")
	}

	tmp, err := ioutil.TempDir("", "TestReadlinkJunction")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	err = os.Chdir(tmp)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(wd)

	// test junctions to the directory
	func() {
		err := os.Mkdir("target", 0777)
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll("target")

		files := []string{"abc", "def", "ghi"}

		for _, file := range files {
			f, err := os.Create(filepath.Join("target", file))
			if err != nil {
				t.Fatal(err)
			}
			err = f.Close()
			if err != nil {
				t.Fatal(err)
			}
		}

		for _, test := range readlinkJunctionTests {
			subName := strings.Replace(test.SubstituteName, "{{tmp}}", tmp, 1)
			printName := strings.Replace(test.PrintName, "{{tmp}}", tmp, 1)
			want := strings.Replace(test.Want, "{{tmp}}", tmp, 1)

			rdbuf, err := makeMountPointReparseDataBuffer(subName, printName, test.PathBuffer)
			if err != nil {
				t.Errorf("cannot make the mount point reparse data buffer: %v\n", err)

				continue
			}

			{
				err = createReparsePoint("link", rdbuf, true)
				if err != nil {
					t.Errorf("cannot create the reparse point: %v\n", err)

					goto clean
				}

				fis, err := ioutil.ReadDir("link")
				if err != nil {
					t.Errorf("cannot readdir from the link: %v\n", err)

					goto clean
				}

				fnames := make([]string, len(fis))
				for i, fi := range fis {
					fnames[i] = fi.Name()
				}

				if len(fnames) != len(files) {
					t.Errorf("filenames are mismatch: expected %v, got %v\n", files, fnames)

					goto clean
				}

				for i := range fnames {
					if files[i] != fnames[i] {
						t.Errorf("filenames are mismatch: expected %v, got %v\n", files, fnames)

						goto clean
					}
				}

				readlink, err := os.Readlink("link")
				if err != nil {
					t.Errorf("cannot readlink from the link: %v\n", err)

					goto clean
				}

				if readlink != want {
					t.Errorf("readlink is mismatch: expected %q, got %q\n", want, readlink)

					goto clean
				}
			}

		clean:
			os.Remove("link")
		}
	}()
}
