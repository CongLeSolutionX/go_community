// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This program generates a test to verify that a program can be
// successfully linked even when there are very large text
// sections present.

package main

import (
	"bytes"
	"fmt"
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
	const FN = 4
	tmpdir, err := ioutil.TempDir("", "invalidmaindecl")

	defer os.RemoveAll(tmpdir)

	fmt.Fprintf(&w, "package main\n")
	fmt.Fprintf(&w, "\nvar main = func() { }\n")

	err = ioutil.WriteFile(tmpdir+"/invalidmaindecl.go", w.Bytes(), 0666)
	if err != nil {
		t.Fatalf("can't write output: %v\n", err)
	}

	// Build and run with internal linking.
	os.Chdir(tmpdir)
	cmd := exec.Command(testenv.GoToolPath(t), "build", "-o", "invalidmaindecl")
	out, err := cmd.CombinedOutput()
	if err == nil || !strings.Contains(string(out), "main.main should be declared as a function") {
		t.Fatalf("Build failed for invalidmaindecl program: %v, output: %s", err, out)
	}
}
