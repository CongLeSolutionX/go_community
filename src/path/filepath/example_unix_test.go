// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !windows,!plan9

package filepath_test

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func ExampleSplitList() {
	fmt.Println("On Unix:", filepath.SplitList("/a/b/c:/usr/bin"))
	// Output:
	// On Unix: [/a/b/c /usr/bin]
}

func ExampleRel() {
	paths := []string{
		"/a/b/c",
		"/b/c",
		"./b/c",
	}
	base := "/a"

	fmt.Println("On Unix:")
	for _, p := range paths {
		rel, err := filepath.Rel(base, p)
		fmt.Printf("%q: %q %v\n", p, rel, err)
	}

	// Output:
	// On Unix:
	// "/a/b/c": "b/c" <nil>
	// "/b/c": "../b/c" <nil>
	// "./b/c": "" Rel: can't make ./b/c relative to /a
}

func ExampleSplit() {
	paths := []string{
		"/home/arnie/amelia.jpg",
		"/mnt/photos/",
		"rabbit.jpg",
		"/usr/local//go",
	}
	fmt.Println("On Unix:")
	for _, p := range paths {
		dir, file := filepath.Split(p)
		fmt.Printf("input: %q\n\tdir: %q\n\tfile: %q\n", p, dir, file)
	}
	// Output:
	// On Unix:
	// input: "/home/arnie/amelia.jpg"
	// 	dir: "/home/arnie/"
	// 	file: "amelia.jpg"
	// input: "/mnt/photos/"
	// 	dir: "/mnt/photos/"
	// 	file: ""
	// input: "rabbit.jpg"
	// 	dir: ""
	// 	file: "rabbit.jpg"
	// input: "/usr/local//go"
	// 	dir: "/usr/local//"
	// 	file: "go"
}

func ExampleJoin() {
	fmt.Println("On Unix:")
	fmt.Println(filepath.Join("a", "b", "c"))
	fmt.Println(filepath.Join("a", "b/c"))
	fmt.Println(filepath.Join("a/b", "c"))
	fmt.Println(filepath.Join("a/b", "/c"))
	// Output:
	// On Unix:
	// a/b/c
	// a/b/c
	// a/b/c
	// a/b/c
}
func ExampleIsAbs() {
	paths := []string{
		"",
		"/",
		"/usr/bin/gcc",
		"..",
		"/a/../bb",
		".",
		"./",
		"lala",
	}
	fmt.Println("On Unix:")
	for _, p := range paths {
		isAbs := filepath.IsAbs(p)
		fmt.Printf("input: %q\n\tisAbs: %t\n", p, isAbs)
	}
	// Output:
	// On Unix:
	// input: ""
	// 	isAbs: false
	// input: "/"
	// 	isAbs: true
	// input: "/usr/bin/gcc"
	// 	isAbs: true
	// input: ".."
	// 	isAbs: false
	// input: "/a/../bb"
	// 	isAbs: true
	// input: "."
	// 	isAbs: false
	// input: "./"
	// 	isAbs: false
	// input: "lala"
	// 	isAbs: false
}

func ExampleBase() {
	paths := []string{
		"",
		".",
		"/.",
		"/",
		"////",
		"x/",
		"abc",
		"abc/def",
		"a/b/.x",
		"a/b/c.",
		"a/b/c.x",
	}
	fmt.Println("On Unix:")
	for _, p := range paths {
		base := filepath.Base(p)
		fmt.Printf("input: %q\n\tbase: %q\n", p, base)
	}
	// Output:
	// On Unix:
	// input: ""
	// 	base: "."
	// input: "."
	// 	base: "."
	// input: "/."
	// 	base: "."
	// input: "/"
	// 	base: "/"
	// input: "////"
	// 	base: "/"
	// input: "x/"
	// 	base: "x"
	// input: "abc"
	// 	base: "abc"
	// input: "abc/def"
	// 	base: "def"
	// input: "a/b/.x"
	// 	base: ".x"
	// input: "a/b/c."
	// 	base: "c."
	// input: "a/b/c.x"
	// 	base: "c.x"
}

func ExampleClean() {
	paths := []string{
		"",
		".",
		"../..",
		"a/b/c",
		"abc/def/",
		"./",
		"../../",
		"abc//def///ghi//",
		"abc/def/ghi/../jkl",
		"/abc/def/../..",
		"abc//./../def",
		"abc/../../././../def",
	}
	fmt.Println("On Unix:")
	for _, p := range paths {
		clean := filepath.Clean(p)
		fmt.Printf("input: %q\n\tclean: %q\n", p, clean)
	}
	// Output:
	// On Unix:
	// input: ""
	// 	clean: "."
	// input: "."
	// 	clean: "."
	// input: "../.."
	// 	clean: "../.."
	// input: "a/b/c"
	// 	clean: "a/b/c"
	// input: "abc/def/"
	// 	clean: "abc/def"
	// input: "./"
	// 	clean: "."
	// input: "../../"
	// 	clean: "../.."
	// input: "abc//def///ghi//"
	// 	clean: "abc/def/ghi"
	// input: "abc/def/ghi/../jkl"
	// 	clean: "abc/def/jkl"
	// input: "/abc/def/../.."
	// 	clean: "/"
	// input: "abc//./../def"
	// 	clean: "def"
	// input: "abc/../../././../def"
	// 	clean: "../../def"
}

func ExampleDir() {
	paths := []string{
		"",
		".",
		"/.",
		"/",
		"////",
		"/foo",
		"x/",
		"abc",
		"abc/def",
		"a/b/.x",
		"a/b/c.",
		"a/b/c.x",
	}
	fmt.Println("On Unix:")
	for _, p := range paths {
		dir := filepath.Dir(p)
		fmt.Printf("input: %q\n\tdir: %q\n", p, dir)
	}
	// Output:
	// On Unix:
	// input: ""
	// 	dir: "."
	// input: "."
	// 	dir: "."
	// input: "/."
	// 	dir: "/"
	// input: "/"
	// 	dir: "/"
	// input: "////"
	// 	dir: "/"
	// input: "/foo"
	// 	dir: "/"
	// input: "x/"
	// 	dir: "x"
	// input: "abc"
	// 	dir: "."
	// input: "abc/def"
	// 	dir: "abc"
	// input: "a/b/.x"
	// 	dir: "a/b"
	// input: "a/b/c."
	// 	dir: "a/b"
	// input: "a/b/c.x"
	// 	dir: "a/b"
}

func ExampleWalk() {
	dir := "dir/to/walk"
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		fmt.Printf("visited file: %q in rootdir: %q\n", path, dir)
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}
