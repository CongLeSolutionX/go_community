// +build !nacl
// run

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
)

func main() {
	dir, err := ioutil.TempDir("", "issue8983")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	setupTempDirsAndWriteFiles(dir)

	cmd := exec.Command("go", "run", filepath.Join(dir, "main.go"))
	output, err := cmd.CombinedOutput()
	if err == nil {
		log.Fatal("unexpected success")
	}

	orig := make([]byte, len(output))
	copy(orig, output)
	wantPatterns := []string{
		".+.go:13:22: cannot use\n",
		"\t+a1.NewCloser\\(\\) \\(type a1.Closer\\) \\(package .+a1\\)\n",
		"as\n",
		"\t+type Closer \\(package main\\)\n",
		"in return argument\n",
		".+.go:17:22: cannot use\n",
		"\ta2.NilCloser\\(\\) \\(type a2.Closer\\) \\(package .+a2\\)\n",
		"as\n",
		"\ttype Closer \\(package main\\)\n",
		"in return argument\n",
		".+.go:23:19: cannot use\n",
		"\t+errors.New\\(\"x\"\\) \\(type error\\) \\(package builtin\\)\n",
		"as\n",
		"\ttype error \\(package main\\)\n",
		"in return argument\n",

		// Moved down here to avoid distracting from the important matches.
		"# command-line-arguments\n",
	}

	for _, pat := range wantPatterns {
		reg := regexp.MustCompile(pat)
		match := reg.Find(output)
		if len(match) == 0 {
			log.Fatalf("Failed to match pattern: %q\ninput: %s\norig: %s\n",
				pat, output, orig)
		}
		index := bytes.Index(output, match)
		output = append(output[:index], output[index+len(match):]...)
	}
	if len(output) != 0 {
		log.Fatalf("Unmatched content:\n%q", output)
	}
}

func setupTempDirsAndWriteFiles(dir string) {
	// Setup phase
	for layout, body := range setup {
		fullDir := filepath.Join(dir, layout.dir)
		if err := os.MkdirAll(fullDir, 0755); err != nil {
			log.Fatalf("MdkdirAll: %q err: %v", fullDir, err)
		}
		fullPath := filepath.Join(fullDir, layout.path)
		f, err := os.Create(fullPath)
		if err != nil {
			log.Fatalf("Create file: %q err: %v", fullPath, err)
		}
		fmt.Fprintf(f, body)
		f.Close()
	}
}

type layout struct {
	dir  string
	path string
}

var setup = map[*layout]string{
	{".", "main.go"}: `package main

import (
  "errors"

  "./a1"
  "./a2"
)

type Closer int

func newOne1() Closer {
  return a1.NewCloser()
}

func newOne2() Closer {
  return a2.NilCloser()
}

type error int

func F() error {
	return errors.New("x")
}

func main() {
}`,
	{"a1", "a1.go"}: `package a1

type Closer interface {
  Close() error
}

func NewCloser() Closer {
  return nil
}`,
	{"a2", "a2.go"}: `package a2

type Closer interface {
  Close() error
}

func NilCloser() Closer {
  return nil
}`,
}
