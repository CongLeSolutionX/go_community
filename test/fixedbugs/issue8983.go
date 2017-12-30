// +build !nacl
// run

// Copyright 2017 The Go Authors. All rights reserved.
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

	setItUp(dir)

	cmd := exec.Command("go", "run", filepath.Join(dir, "main.go"))
	output, err := cmd.CombinedOutput()
	if err == nil {
		log.Fatal("unexpected success")
	}

orig := make([]byte, len(output))
copy(orig, output)
	wantPatterns := []string{
		"13:22: cannot use",
		"\ta1.NewCloser\\(\\) \\(type a1.Closer\\) \\(package .+a1\\)",
		"as",
		"\ttype Closer \\(package main\\)$",
		"in return argument$",
		"17:22: cannot use$",
		"\ta2.NilCloser\\(\\) \\(type a2.Closer\\) \\(package .+a2\\)$",
		"as",
		"\ttype Closer \\(package main\\)",
		"in return argument",
		"cannot use",
		`errors.New\\("x"\\) \\(type error\\) \\(package go\\.builtin\\)`,
		"as",
		"\ttype error \\(package main\\)",
	}

	for _, pat := range wantPatterns {
		reg := regexp.MustCompile(pat)
		match := reg.Find(output)
		if len(match) == 0 {
			log.Fatalf("failed to match pattern: %q\ninput: %s\norig: %s\n", pat, output, orig)
		}
		index := bytes.Index(output, match)
		output = bytes.Join([][]byte{output[:index], output[index+len(match):]}, nil)
	}
}

func setItUp(dir string) {
	// Setup phase
	for layout, body := range setup {
		fullDir := filepath.Join(dir, layout.dir)
		if err := os.MkdirAll(fullDir, 0755); err != nil {
			log.Fatalf("%q mkdir %v", fullDir, err)
		}
		fullPath := filepath.Join(fullDir, layout.path)
		f, err := os.Create(fullPath)
		if err != nil {
			log.Fatalf("%q create: %v", fullPath, err)
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
