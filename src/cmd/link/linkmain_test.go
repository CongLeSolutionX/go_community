// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"internal/testenv"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestInvalidMainDecl(t *testing.T) {
	testenv.MustHaveGoBuild(t)

	src := filepath.Join("testdata", "invalidmaindecl.go")
	dst := filepath.Join("testdata", "invalidmaindecl")

	cmd := exec.Command(testenv.GoToolPath(t), "build", "-o", dst, src)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("build succeeded unexpectedly")
	}
	if !strings.Contains(string(out), "main.main must be a function") {
		t.Errorf("Failing build produced unexpected output: %s (%v)", out, err)
	}
}
