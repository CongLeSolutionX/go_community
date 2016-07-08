// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gc

import (
	"testing"
)

func TestStringEqual(t *testing.T) {
	classify := func(s string) int {
		switch s {
		case "":
			return 0
		case "a", "b", "c":
			return 1
		case "ä":
			return 2
		case "hello":
			return 5
		case "very long string":
			return 16
		}
		return -len(s)
	}

	testdata := []struct {
		s        string
		expected int
	}{
		{"", 0},
		{"a", 1},
		{"b", 1},
		{"c", 1},
		{"d", -1},
		{"ä", 2},
		{"ö", -2},
		{"hello", 5},
		{"world", -5},
		{"very long string", 16},
		{"even longer string", -18},
	}

	for _, datum := range testdata {
		actual := classify(datum.s)
		if actual != datum.expected {
			t.Errorf("Comparing %q failed: expected %d, got %d",
				datum.s, datum.expected, actual)
		}
	}
}
