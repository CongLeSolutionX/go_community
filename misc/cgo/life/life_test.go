// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package life_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestTestRun(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	GOPATH, err := ioutil.TempDir("", "cgolife")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(GOPATH)
	os.Setenv("GOPATH", GOPATH)

	modRoot := filepath.Join(GOPATH, "src", "cgolife")
	if err := os.MkdirAll(modRoot, 0777); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(filepath.Join(modRoot, "go.mod"), []byte("module cgolife\n"), 0666); err != nil {
		t.Fatal(err)
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
		t.Fatal(err)
	}

	out, err := exec.Command("go", "env", "GOROOT").Output()
	if err != nil {
		t.Fatal(err)
	}
	GOROOT := string(bytes.TrimSpace(out))

	cmd := exec.Command("go", "run", filepath.Join(GOROOT, "test", "run.go"), "-", ".")
	cmd.Dir = modRoot
	out, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s: %s\n%s", strings.Join(cmd.Args, " "), err, out)
	}
	t.Logf("%s:\n%s", strings.Join(cmd.Args, " "), out)
}
