<<<<<<< HEAD   (c83a43 [dev.go2go] go/*: merge parser and types changes from dev.ty)
=======
// UNREVIEWED
>>>>>>> BRANCH (dc122c [dev.typeparams] test: exclude a failing test again (fix 32b)
// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types2

import "testing"

func TestStripAnnotations(t *testing.T) {
	for _, test := range []struct {
		in, want string
	}{
		{"", ""},
		{"   ", "   "},
		{"foo", "foo"},
		{"foo₀", "foo"},
		{"foo(T₀)", "foo(T)"},
		{"#foo(T₀)", "foo(T)"},
	} {
		got := stripAnnotations(test.in)
		if got != test.want {
			t.Errorf("%q: got %q; want %q", test.in, got, test.want)
		}
	}
}
