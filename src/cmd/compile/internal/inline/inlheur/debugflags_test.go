// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package inlheur

import (
	"internal/testenv"
	"path/filepath"
	"testing"
)

func TestSelectedDebugFlags(t *testing.T) {
	td := t.TempDir()
	testenv.MustHaveGoBuild(t)
	if testing.Short() {
		t.Skip("no need to test in short mode")
	}
	outpath := filepath.Join(td, "example.a")
	srcpath := filepath.Join("testdata", "dumpscores.go")
	run := []string{testenv.GoToolPath(t), "tool", "compile", "-p", "example",
		"-d=inlscoreadj=passConstToNestedIfAdj:-999/passInlinableFuncToNestedIndCallAdj:32", "-o", outpath, srcpath}
	out, err := testenv.Command(t, run[0], run[1:]...).CombinedOutput()
	if err != nil {
		t.Logf("run: %+v\n", run)
		t.Logf("out: %s\n", out)
		t.Fatalf("problems compiling with -d=inlscoreadj")
	}
}
