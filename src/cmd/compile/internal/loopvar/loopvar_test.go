// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package loopvar_test

import (
	"internal/testenv"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type testcase struct {
	lvFlag      string // ==-2, -1, 0, 1, 2
	buildExpect string // message, if any
	expectRC    int
	files       []string
}

var for_files = []string{
	"for_esc_address.go",             // address of variable
	"for_esc_closure.go",             // closure of variable
	"for_esc_minimal_closure.go",     // simple closure of variable
	"for_esc_method.go",              // method value of variable
	"for_complicated_esc_address.go", // modifies loop index in body
}

var range_files = []string{
	"range_esc_address.go",         // address of variable
	"range_esc_closure.go",         // closure of variable
	"range_esc_minimal_closure.go", // simple closure of variable
	"range_esc_method.go",          // method value of variable
}

var cases = []testcase{
	{"-2", "maybe for 3-clause loop variable leak i", 11, for_files},
	{"-1", "", 11, for_files[:1]},
	{"0", "", 0, for_files[:1]},
	{"1", "", 0, for_files[:1]},
	{"2", "transformed loop variable i ", 0, for_files},

	{"-2", "maybe range loop variable leak i", 11, range_files},
	{"-1", "", 11, range_files[:1]},
	{"0", "", 0, range_files[:1]},
	{"1", "", 0, range_files[:1]},
	{"2", "transformed loop variable i ", 0, range_files},

	{"1", "", 0, []string{"for_nested.go"}},
}

// TestFmaHash checks that the hash-test machinery works properly for a single case.
// It also runs ssa/check and gccheck to be sure that those are checked at least a
// little in each run.bash.  It does not check or run the generated code.
// The test file is however a useful example of fused-vs-cascaded multiply-add.
func TestLoopVar(t *testing.T) {
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
	tmpdir, err := os.MkdirTemp("", "x")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(tmpdir)
	output := filepath.Join(tmpdir, "foo.exe")

	for i, tc := range cases {
		for _, f := range tc.files {
			source := f
			cmd := testenv.Command(t, gocmd, "build", "-o", output, "-gcflags=-d=loopvar="+tc.lvFlag, source)
			cmd.Env = append(cmd.Env, "GOEXPERIMENT=loopvar", "HOME="+tmpdir)
			cmd.Dir = "testdata"
			t.Logf("File %s loopvar=%s expect '%s' exit code %d", f, tc.lvFlag, tc.buildExpect, tc.expectRC)
			b, e := cmd.CombinedOutput()
			if e != nil {
				t.Error(e)
			}
			if tc.buildExpect != "" {
				s := string(b)
				if !strings.Contains(s, tc.buildExpect) {
					t.Errorf("File %s test %d expected to match '%s' with \n-----\n%s\n-----", f, i, tc.buildExpect, s)
				}
			}
			// run what we just built.
			cmd = testenv.Command(t, output)
			b, e = cmd.CombinedOutput()
			if tc.expectRC != 0 {
				if e == nil {
					t.Errorf("Missing expected error, file %s, case %d", f, i)
				} else if ee, ok := (e).(*exec.ExitError); !ok || ee.ExitCode() != tc.expectRC {
					t.Error(e)
				} else {
					// okay
				}
			} else if e != nil {
				t.Error(e)
			}
		}
	}
}
