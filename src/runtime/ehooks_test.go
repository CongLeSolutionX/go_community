// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime_test

import (
	"os/exec"
	"strings"
	"testing"
)

func TestExitHooks(t *testing.T) {
	scenarios := []struct {
		mode     string
		expected string
		musthave string
	}{
		{mode: "simple",
			expected: `bar
foo
`,
			musthave: "",
		},
		{mode: "goodexit",
			expected: `orange
apple
`,
			musthave: "",
		},
		{mode: "badexit",
			expected: `blub
blix
`,
			musthave: "",
		},
		{mode: "panics",
			expected: "",
			musthave: "fatal error: internal error: exit hook invoked panic",
		},
		{mode: "callsexit",
			expected: "",
			musthave: "fatal error: internal error: exit hook invoked exit",
		},
	}

	exe, err := buildTestProg(t, "testexithooks", "")
	if err != nil {
		t.Fatal(err)
	}

	for _, s := range scenarios {
		cmd := exec.Command(exe, []string{"-mode", s.mode}...)
		out, _ := cmd.CombinedOutput()
		outs := string(out)
		if s.expected != "" {
			if s.expected != outs {
				t.Logf("raw output: %q", outs)
				t.Errorf("failed mode %s: wanted %q got %q",
					s.mode, s.expected, outs)
			}
		} else if s.musthave != "" {
			if !strings.Contains(outs, s.musthave) {
				t.Logf("raw output: %q", outs)
				t.Errorf("failed mode %s: output does not contain %q",
					s.mode, s.musthave)
			}
		} else {
			panic("badly written scenario")
		}
	}
}
