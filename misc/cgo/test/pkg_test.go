// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cgotest

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestCrossPackageTests compiles and runs tests that depend on imports of other
// local packages, using source code stored in the testdata directory.
//
// The tests in the misc directory tree do not have a valid import path in
// GOPATH mode, so they previously used relative imports. However, relative
// imports do not work in module mode. In order to make the test work in both
// modes, we synthesize a GOPATH in which the module paths are equivalent, and
// run the tests as a subprocess.
//
// If and when we no longer support these tests in GOPATH mode, we can remove
// this shim and move the tests currently located in testdata back into the
// parent directory.
func TestCrossPackageTests(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	GOPATH, err := ioutil.TempDir("", "cgotest")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(GOPATH)

	modRoot := filepath.Join(GOPATH, "src", "cgotest")
	if err := os.MkdirAll(modRoot, 0777); err != nil {
		log.Fatal(err)
	}
	if err := ioutil.WriteFile(filepath.Join(modRoot, "go.mod"), []byte("module cgotest\n"), 0666); err != nil {
		log.Fatal(err)
	}
	err = filepath.Walk("testdata", func(path string, info os.FileInfo, err error) error {
		if err != nil || path == "testdata" {
			return err
		}

		dstPath := modRoot + strings.TrimPrefix(path, "testdata")
		if info.IsDir() {
			return os.Mkdir(dstPath, 0777)
		}

		data, err := ioutil.ReadFile(filepath.Join(cwd, path))
		if err != nil {
			return err
		}
		return ioutil.WriteFile(dstPath, data, info.Mode())
	})
	if err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command("go", "test")
	if testing.Verbose() {
		cmd.Args = append(cmd.Args, "-v")
	}
	if testing.Short() {
		cmd.Args = append(cmd.Args, "-short")
	}
	cmd.Dir = modRoot
	cmd.Env = append(os.Environ(), "GOPATH="+GOPATH)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Logf("%s:\n%s", strings.Join(cmd.Args, " "), out)
	} else {
		t.Fatalf("%s: %s\n%s", strings.Join(cmd.Args, " "), err, out)
	}
}
