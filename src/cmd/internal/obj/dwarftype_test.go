// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package obj_test

import (
	"cmd/internal/objfile"
	"internal/testenv"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func gobuild(t *testing.T, dir string, testfile string) (string, *objfile.File) {

	src := filepath.Join(dir, "test.go")
	dst := filepath.Join(dir, "out.o")

	f, err := os.Create(src)
	if err != nil {
		t.Fatal(err)
	}

	f.WriteString(testfile)
	f.Close()

	args := []string{"build"}
	args = append(args, "-gcflags=-p=p")
	args = append(args, "-o", dst, src)

	cmd := exec.Command(testenv.GoToolPath(t), args...)
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Logf("build: %s\n", string(b))
		t.Fatal(err)
	}
	if err != nil {
		t.Fatal(err)
	}
	pkg, err := objfile.Open(dst)
	if err != nil {
		t.Fatal(err)
	}
	return src, pkg
}

var basicTypes = map[string]struct{}{
	"go.info.int8":               {},
	"go.info.*int8":              {},
	"go.info.uint8":              {},
	"go.info.*uint8":             {},
	"go.info.int16":              {},
	"go.info.*int16":             {},
	"go.info.uint16":             {},
	"go.info.*uint16":            {},
	"go.info.int32":              {},
	"go.info.*int32":             {},
	"go.info.uint32":             {},
	"go.info.*uint32":            {},
	"go.info.int64":              {},
	"go.info.*int64":             {},
	"go.info.uint64":             {},
	"go.info.*uint64":            {},
	"go.info.int":                {},
	"go.info.*int":               {},
	"go.info.uint":               {},
	"go.info.*uint":              {},
	"go.info.uintptr":            {},
	"go.info.*uintptr":           {},
	"go.info.complex64":          {},
	"go.info.*complex64":         {},
	"go.info.complex128":         {},
	"go.info.*complex128":        {},
	"go.info.float32":            {},
	"go.info.*float32":           {},
	"go.info.float64":            {},
	"go.info.*float64":           {},
	"go.info.bool":               {},
	"go.info.*bool":              {},
	"go.info.string":             {},
	"go.info.*string":            {},
	"go.info.unsafe.Pointer":     {},
	"go.info.*unsafe.Pointer":    {},
	"go.info.error":              {},
	"go.info.*error":             {},
	"go.info.func(error) string": {},
}

func checkDoNotDumpTypes(t *testing.T, sym objfile.Sym) {
	if _, ok := basicTypes[sym.Name]; ok {
		if sym.Size != 0 {
			t.Errorf("found redundant dwarf type %s in objfile", sym.Name)
		}
	}
}

func TestNamedTypes(t *testing.T) {
	testenv.MustHaveGoRun(t)
	t.Parallel()

	if runtime.GOOS == "plan9" {
		t.Skip("skipping on plan9; no DWARF symbol table in executables")
	}

	dir, err := os.MkdirTemp("", "TestNamedTypes")
	if err != nil {
		t.Fatalf("could not create directory: %v", err)
	}
	defer os.RemoveAll(dir)

	gosrc := `package dwarftype

type t1 struct {
	f1 int
}
type t2 [10]int

type t3 *int
`

	want := map[string]bool{
		"go.info.*p.t1": false,
		"go.info.p.t1":  false,
		"go.info.*p.t2": false,
		"go.info.p.t2":  false,
		"go.info.*p.t3": false,
		"go.info.p.t3":  false,
	}
	_, f := gobuild(t, dir, gosrc)
	defer f.Close()
	syms, err := f.Symbols()
	if err != nil {
		t.Fatal(err)
	}
	for _, sym := range syms {
		checkDoNotDumpTypes(t, sym)
		if _, ok := want[sym.Name]; ok {
			if sym.Size == 0 {
				t.Errorf("found zero size dwarf type %s in objfile", sym.Name)
			}
			want[sym.Name] = true
		}
	}
	for name, exist := range want {
		if !exist {
			t.Errorf("type %s must be found in objfile", name)
		}
	}
}

func TestUnamedTypes(t *testing.T) {
	testenv.MustHaveGoRun(t)
	t.Parallel()

	if runtime.GOOS == "plan9" {
		t.Skip("skipping on plan9; no DWARF symbol table in executables")
	}

	dir, err := os.MkdirTemp("", "TestUnamedTypes")
	if err != nil {
		t.Fatalf("could not create directory: %v", err)
	}
	defer os.RemoveAll(dir)

	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module dwarftype"), 0666); err != nil {
		t.Fatal(err)
	}

	gosrc := `package dwarftype
var v1 struct{
	f1 int
}
`

	want := map[string]bool{
		"go.info.struct { p.f1 int }":  false,
		"go.info.*struct { p.f1 int }": false,
	}
	_, f := gobuild(t, dir, gosrc)
	defer f.Close()
	syms, err := f.Symbols()
	if err != nil {
		t.Fatal(err)
	}
	for _, sym := range syms {
		checkDoNotDumpTypes(t, sym)
		if _, ok := want[sym.Name]; ok {
			if sym.Size == 0 {
				t.Errorf("found zero size dwarf type %s in objfile", sym.Name)
			}
			want[sym.Name] = true
		}
	}
	for name, exist := range want {
		if !exist {
			t.Errorf("type %s must be found in objfile", name)
		}
	}
}

// Make sure the basic types do not be dumped in a non-runtime package.
func TestDoNotDumpBasicTypes(t *testing.T) {
	testenv.MustHaveGoRun(t)
	t.Parallel()

	if runtime.GOOS == "plan9" {
		t.Skip("skipping on plan9; no DWARF symbol table in executables")
	}

	dir, err := os.MkdirTemp("", "TestDoNotDumpBasicTypes")
	if err != nil {
		t.Fatalf("could not create directory: %v", err)
	}
	defer os.RemoveAll(dir)

	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module dwarftype"), 0666); err != nil {
		t.Fatal(err)
	}

	gosrc := `package dwarftype

import "unsafe"

var i8 *int8
var u8 *uint8
var i16 *int16
var u16 *uint16
var i32 *int32
var u32 *uint32
var i64 *int64
var u64 *uint64
var i *int
var u *uint
var up *uintptr
var c64 *complex64
var c128 *complex128
var f32 *float32
var f64 *float64
var b *bool
var s *string
var uptr *unsafe.Pointer
var e *error

var f func(error) string

`
	_, f := gobuild(t, dir, gosrc)
	defer f.Close()
	syms, err := f.Symbols()
	if err != nil {
		t.Fatal(err)
	}
	for _, sym := range syms {
		checkDoNotDumpTypes(t, sym)
	}

}

func TestSynthesizeMapTypes(t *testing.T) {
	testenv.MustHaveGoRun(t)
	t.Parallel()

	if runtime.GOOS == "plan9" {
		t.Skip("skipping on plan9; no DWARF symbol table in executables")
	}

	dir, err := os.MkdirTemp("", "TestDoNotDumpBasicTypes")
	if err != nil {
		t.Fatalf("could not create directory: %v", err)
	}
	defer os.RemoveAll(dir)

	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module dwarftype"), 0666); err != nil {
		t.Fatal(err)
	}

	gosrc := `package dwarftype
var m map[int]int
`

	want := map[string]bool{
		"go.info.[]key<int>":       false,
		"go.info.[]val<int>":       false,
		"go.info.bucket<int,int>":  false,
		"go.info.*bucket<int,int>": false,
		"go.info.hash<int,int>":    false,
		"go.info.*hash<int,int>":   false,
	}
	_, f := gobuild(t, dir, gosrc)
	defer f.Close()
	syms, err := f.Symbols()
	if err != nil {
		t.Fatal(err)
	}
	for _, sym := range syms {
		checkDoNotDumpTypes(t, sym)
		if _, ok := want[sym.Name]; ok {
			if sym.Size == 0 {
				t.Errorf("found zero size dwarf type %s in objfile", sym.Name)
			}
			want[sym.Name] = true
		}
	}
	for name, exist := range want {
		if !exist {
			t.Errorf("type %s must be found in objfile", name)
		}
	}
}

func TestSynthesizeChanTypes(t *testing.T) {
	testenv.MustHaveGoRun(t)
	t.Parallel()

	if runtime.GOOS == "plan9" {
		t.Skip("skipping on plan9; no DWARF symbol table in executables")
	}

	dir, err := os.MkdirTemp("", "TestDoNotDumpBasicTypes")
	if err != nil {
		t.Fatalf("could not create directory: %v", err)
	}
	defer os.RemoveAll(dir)

	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module dwarftype"), 0666); err != nil {
		t.Fatal(err)
	}

	gosrc := `package dwarftype
var c chan int
`

	want := map[string]bool{
		"go.info.*chan int":   false,
		"go.info.chan int":    false,
		"go.info.sudog<int>":  false,
		"go.info.*sudog<int>": false,
		"go.info.waitq<int>":  false,
		"go.info.hchan<int>":  false,
		"go.info.*hchan<int>": false,
	}
	_, f := gobuild(t, dir, gosrc)
	defer f.Close()
	syms, err := f.Symbols()
	if err != nil {
		t.Fatal(err)
	}
	for _, sym := range syms {
		checkDoNotDumpTypes(t, sym)
		if _, ok := want[sym.Name]; ok {
			if sym.Size == 0 {
				t.Errorf("found zero size dwarf type %s in objfile", sym.Name)
			}
			want[sym.Name] = true
		}
	}
	for name, exist := range want {
		if !exist {
			t.Errorf("type %s must be found in objfile", name)
		}
	}
}
