// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This program generates a test to verify that a program can be
// successfully linked even when there are very large text
// sections present.

package main

import (
	"bytes"
	"internal/testenv"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestInvalidMainDecl(t *testing.T) {
	testenv.MustHaveGoBuild(t)

	var w bytes.Buffer
	tmpdir, err := ioutil.TempDir("", "invalidmaindecl")

	defer os.RemoveAll(tmpdir)
	const invalidMainDecl = `
	package main

	var main = func() {}
	`

	err = ioutil.WriteFile(tmpdir+"/invalidmaindecl.go", string(invalidMainDecl), 0666)
	if err != nil {
		t.Fatalf("can't write output: %v\n", err)
	}

	// Build and run with internal linking.
	err = os.Chdir(tmpdir)
	if err != nil {
		t.Fatalf("can't change directory: %v\n", err)
	}

	cmd := exec.Command(testenv.GoToolPath(t), "build", "-o", "invalidmaindecl")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Error("build succeeded unexpectedly")
	}
	if !strings.Contains(string(out), "main.main must be a function") {
		t.Errorf("Build failed for invalidmaindecl program: %v, output: %s", err, out)
	}
}
