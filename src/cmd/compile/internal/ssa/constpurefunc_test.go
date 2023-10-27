// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa_test

import (
	"internal/testenv"
	"os"
	"path/filepath"
	// "regexp"
	"runtime"
	"testing"
)

// TestFmaHash checks that the hash-test machinery works properly for a single case.
// It also runs ssa/check and gccheck to be sure that those are checked at least a
// little in each run.bash.  It does not check or run the generated code.
// The test file is however a useful example of fused-vs-cascaded multiply-add.
func TestConstPureFunc(t *testing.T) {
	switch runtime.GOOS {
	case "linux", "darwin":
	default:
		t.Skipf("Slow test, usually avoid it, os=%s not linux or darwin", runtime.GOOS)
	}
	switch runtime.GOARCH {
	case "amd64", "arm64":
	default:
		t.Skipf("Slow test, usually avoid it, arch=%s not amd64 or arm64", runtime.GOARCH)
	}

	testenv.MustHaveGoBuild(t)
	gocmd := testenv.GoToolPath(t)
	tmpdir := t.TempDir()

	source := filepath.Join("testdata", "const_pure_func_test")
	output := filepath.Join(tmpdir, "const_pure_func_test")
	output_cmd := filepath.Join(output, "cmd")

	copy := func(s string) {
		b, e := os.ReadFile(filepath.Join(source, s))
		if e != nil {
			t.Fatalf("copy %s read failed, %v", s, e)
		}
		e = os.WriteFile(filepath.Join(output, s), b, 0750)
		if e != nil {
			t.Fatalf("copy %s write failed, %v", s, e)
		}
	}
	os.Mkdir(output, 0750)
	os.Mkdir(output_cmd, 0750)
	copy("go.mod")
	copy("test.go")
	copy(filepath.Join("cmd", "main.go"))

	cmd := testenv.Command(t, gocmd, "run", ".")
	cmd.Dir = output_cmd
	// The hash-dependence on file path name is dodged by specifying "all hashes ending in 1" plus "all hashes ending in 0"
	// i.e., all hashes.  This will print all the FMAs; this test is only interested in one of them (that should appear near the end).
	t.Logf("%v", cmd)
	t.Logf("%v", cmd.Env)
	b, e := cmd.CombinedOutput()
	if e != nil {
		t.Error(e)
	}
	got := string(b)
	expected :=
		`AddConst 3, 4 called
AddPure 14, 6 called
AddPure 14, 6 called
a=14, c=60, d=60
`
	if got != expected {
		t.Errorf("Expected '%s' but got '%s'", expected, got)
	} else {
		t.Logf(got)

	}
}
