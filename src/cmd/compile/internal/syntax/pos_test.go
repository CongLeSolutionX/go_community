// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syntax

import (
	"fmt"
	"testing"
)

func TestMakePos(t *testing.T) {
	for _, test := range []struct {
		line, col uint
		str       string
	}{
		{0, 0, "0:0"},
		{1, 0, "1:0"},
		{0, 1, "0:1"},
		{1, 1, "1:1"},
		{10, 1, "10:1"},
		{100, 10, "100:10"},
		{lineMask, colMask, fmt.Sprintf("%d:%d", lineMask, colMask)},
		{lineMask + 1, colMask + 1, fmt.Sprintf("%d:%d", lineMask, colMask)},
	} {
		p := MakePos(test.line, test.col)
		if l := saturate(test.line, lineMask); p.Line() != l {
			t.Errorf("%d:%d: line = %d; want %d", test.line, test.col, p.Line(), l)
		}
		if c := saturate(test.col, colMask); p.Col() != c {
			t.Errorf("%d:%d: col = %d; want %d", test.line, test.col, p.Col(), c)
		}
		if p.String() != test.str {
			t.Errorf("%d:%d: got %q; want %q", test.line, test.col, p.String(), test.str)
		}
	}
}
