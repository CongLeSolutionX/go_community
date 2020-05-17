// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gc_test

import (
	"bytes"
	"internal/testenv"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
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
			tmp, err := ioutil.TempFile("", "")
			if err != nil {
				t.Fatalf("temp file creation failed: %v", err)
			}
			defer os.Remove(tmp.Name())
			defer tmp.Close()
			for i := 0; i < iters; i++ {
				// Note: use -c 2 to expose any nondeterminism which is the result
				// of the runtime scheduler.
				out, err := exec.Command(testenv.GoToolPath(t), "tool", "compile", "-c", "2", "-o", tmp.Name(), filepath.Join("testdata", "reproducible", test)).CombinedOutput()
				if err != nil {
					t.Fatalf("failed to compile: %v\n%s", err, out)
				}
				obj, err := ioutil.ReadFile(tmp.Name())
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

func TestIssue38068(t *testing.T) {
	testenv.MustHaveGoBuild(t)
	t.Parallel()

	// Compile a small test case using two scenarios. In the first scenario,
	// set "-d compilelater" so as to force the same behavior you get with
	// the concurrent back end (delaying backend compilation until all
	// functions have been added to a list).  In the second scenario, turn
	// off concurrent compilation completely, meaning that each function
	// will be sent through the backend right away. Dwarf compression is
	// turned off for the link just out of paranoia (in case we're on a
	// platform where external linking is the default, and the external
	// linker's dwarf compression is non-deterministic).
	scenarios := []struct {
		tag     string
		env     string
		args    string
		exepath string
	}{
		{tag: "concurrent", env: "GO19CONCURRENTCOMPILATION=0", args: "-d=compilelater"},
		{tag: "serial", env: "GO19CONCURRENTCOMPILATION=0", args: ""}}

	tmpdir, err := ioutil.TempDir("", "TestIssue38068")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	src := filepath.Join("testdata", "reproducible", "issue38068.go")
	for i := range scenarios {
		s := &scenarios[i]
		s.exepath = filepath.Join(tmpdir, s.tag+".exe")
		cmd := exec.Command(testenv.GoToolPath(t), "build", "-trimpath", "-ldflags=-compressdwarf=0 -buildid=", "-gcflags="+s.args, "-o", s.exepath, src)
		if s.env != "" {
			cmd.Env = append(os.Environ(), s.env)
		}
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("%v: %v:\n%s", cmd.Args, err, out)
		}
	}

	readBytes := func(fn string) []byte {
		payload, err := ioutil.ReadFile(fn)
		if err != nil {
			t.Fatalf("failed to read executable '%s': %v", fn, err)
		}
		return payload
	}

	b1 := readBytes(scenarios[0].exepath)
	b2 := readBytes(scenarios[1].exepath)
	if !bytes.Equal(b1, b2) {
		t.Fatalf("concurrent and serial builds produced different output")
	}
}
