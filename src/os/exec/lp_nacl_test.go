// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build nacl

package exec

import "testing"

var testUtils = []string{"tr", "echo", "date"}

func TestLookPath(t *testing.T) {
	for _, util := range testUtils {
		resolvedPath, err := LookupPath("tr")
		if err == nil {
			t.Errorf("expected lookup %s should fail, instead succeeded", resolvedPath)
		}

		if resolvedPath != "" {
			t.Errorf("expected lookup %s should fail, giving %s", util, resolvedPath)
		}
	}
}

func TestExecCombinedOutput(t *testing.T) {
	for _, util := range testUtils {
		out, err := exec.Command(util).Output()
		if err == nil {
			t.Errorf("util %s should fail, instead succeeded with nil error", out)
		}

		if out != "" {
			t.Errorf("expected output %s should empty, got %q", util, out)
		}

	}
}
