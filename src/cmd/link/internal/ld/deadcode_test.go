// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ld

import (
	"bytes"
	"internal/testenv"
	"path/filepath"
	"testing"
)

func TestDeadcode(t *testing.T) {
	testenv.MustHaveGoBuild(t)
	t.Parallel()

	tmpdir := t.TempDir()

	mk := func(s string, a ...string) []string {
		if s == "" {
			return []string{}
		} else {
			return append([]string{s}, a...)
		}
	}

	type ss = []string

	tests := []struct {
		src      string
		pos, neg []string // positive and negative patterns
	}{
		{"reflectcall", ss{""}, mk("main.T.M")},
		{"typedesc", mk(""), mk("type:main.T")},
		{"ifacemethod", mk(""), mk("main.T.M")},
		{"ifacemethod2", mk("main.T.M"), mk("")},
		{"ifacemethod3", mk("main.S.M"), mk("")},
		{"ifacemethod4", mk(""), mk("main.T.M")},
		{"globalmap", mk("main.small", "main.effect"), mk("main.large")},
	}
	for _, test := range tests {
		test := test
		t.Run(test.src, func(t *testing.T) {
			t.Parallel()
			src := filepath.Join("testdata", "deadcode", test.src+".go")
			exe := filepath.Join(tmpdir, test.src+".exe")
			cmd := testenv.Command(t, testenv.GoToolPath(t), "build", "-ldflags=-dumpdep", "-o", exe, src)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("%v: %v:\n%s", cmd.Args, err, out)
			}
			for _, pos := range test.pos {
				if !bytes.Contains(out, []byte(pos+"\n")) {
					t.Errorf("%s should be reachable. Output:\n%s", pos, out)
				}
			}
			for _, neg := range test.neg {
				if bytes.Contains(out, []byte(neg+"\n")) {
					t.Errorf("%s should not be reachable. Output:\n%s", neg, out)
				}
			}
		})
	}
}
