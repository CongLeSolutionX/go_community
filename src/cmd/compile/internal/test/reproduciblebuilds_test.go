// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package test

import (
	"bytes"
	"internal/platform"
	"internal/testenv"
	"os"
	"path/filepath"
	"runtime/debug"
	"testing"
)

func TestReproducibleBuilds(t *testing.T) {
	tests := []string{
		"issue20272.go",
		"issue27013.go",
		"issue30202.go",
	}

	testenv.MustHaveGoBuild(t)
	iters := 10
	if testing.Short() {
		iters = 4
	}
	t.Parallel()
	for _, test := range tests {
		test := test
		t.Run(test, func(t *testing.T) {
			t.Parallel()
			var want []byte
			tmp, err := os.CreateTemp("", "")
			if err != nil {
				t.Fatalf("temp file creation failed: %v", err)
			}
			defer os.Remove(tmp.Name())
			defer tmp.Close()
			for i := 0; i < iters; i++ {
				// Note: use -c 2 to expose any nondeterminism which is the result
				// of the runtime scheduler.
				out, err := testenv.Command(t, testenv.GoToolPath(t), "tool", "compile", "-p=p", "-c", "2", "-o", tmp.Name(), filepath.Join("testdata", "reproducible", test)).CombinedOutput()
				if err != nil {
					t.Fatalf("failed to compile: %v\n%s", err, out)
				}
				obj, err := os.ReadFile(tmp.Name())
				if err != nil {
					t.Fatalf("failed to read object file: %v", err)
				}
				if i == 0 {
					want = obj
				} else {
					if !bytes.Equal(want, obj) {
						t.Fatalf("builds produced different output after %d iters (%d bytes vs %d bytes)", i, len(want), len(obj))
					}
				}
			}
		})
	}
}

func readBytes(t *testing.T, fn string) []byte {
	t.Helper()

	payload, err := os.ReadFile(fn)
	if err != nil {
		t.Fatalf("failed to read executable '%s': %v", fn, err)
	}
	return payload
}

func TestIssue38068(t *testing.T) {
	testenv.MustHaveGoBuild(t)
	t.Parallel()

	// Compile a small package with and without the concurrent
	// backend, then check to make sure that the resulting archives
	// are identical.  Note: this uses "go tool compile" instead of
	// "go build" since the latter will generate different build IDs
	// if it sees different command line flags.
	scenarios := []struct {
		tag     string
		args    string
		libpath string
	}{
		{tag: "serial", args: "-c=1"},
		{tag: "concurrent", args: "-c=2"}}

	tmpdir := t.TempDir()

	src := filepath.Join("testdata", "reproducible", "issue38068.go")
	for i := range scenarios {
		s := &scenarios[i]
		s.libpath = filepath.Join(tmpdir, s.tag+".a")
		// Note: use of "-p" required in order for DWARF to be generated.
		cmd := testenv.Command(t, testenv.GoToolPath(t), "tool", "compile", "-p=issue38068", "-buildid=", s.args, "-o", s.libpath, src)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("%v: %v:\n%s", cmd.Args, err, out)
		}
	}

	b1 := readBytes(t, scenarios[0].libpath)
	b2 := readBytes(t, scenarios[1].libpath)
	if !bytes.Equal(b1, b2) {
		t.Fatalf("concurrent and serial builds produced different output")
	}
}

func TestIssue63559(t *testing.T) {
	testenv.MustHaveGoBuild(t)
	t.Parallel()

	goos := os.Getenv("GOOS")
	goarch := os.Getenv("GOARCH")

	bi, ok := debug.ReadBuildInfo()
	if !ok {
		t.Fatal("could not read build info")
	}

	var withRace bool
	for _, s := range bi.Settings {
		if s.Key == "-race" {
			withRace = true
		}
	}

	explicitBuildMode := "exe"
	if platform.DefaultPIE(goos, goarch, withRace) {
		explicitBuildMode = "pie"
	}

	// Compile a small program with and without a default buildmode
	// and check to make sure the resulting binaries are identical.
	scenarios := []struct {
		buildMode string
		binPath   string
	}{
		{buildMode: "default"},
		{buildMode: explicitBuildMode},
	}

	tempDir := t.TempDir()
	src := filepath.Join("testdata", "reproducible", "issue63559.go")
	for i := range scenarios {
		s := &scenarios[i]
		s.binPath = filepath.Join(tempDir, s.buildMode+".out")
		cmd := testenv.Command(t, testenv.GoToolPath(t), "build", "-buildmode", s.buildMode, "-o", s.binPath, src)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("%v: %v:\n%s", cmd.Args, err, out)
		}
	}

	b1 := readBytes(t, scenarios[0].binPath)
	b2 := readBytes(t, scenarios[1].binPath)
	if !bytes.Equal(b1, b2) {
		t.Fatalf("builds produced different output")
	}
}
