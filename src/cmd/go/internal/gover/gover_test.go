// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gover

import "testing"

var compareTests = []struct {
	x   string
	y   string
	out int
}{
	{"", "", 0},
	{"x", "x", 0},
	{"", "x", 0},
	{"1.5", "1.6", -1},
	{"1.5", "1.10", -1},
	{"1.6", "1.6.1", -1},
	{"1.19", "1.19.1", -1},
	{"1.19rc1", "1.19", -1},
	{"1.19rc1", "1.19.1", -1},
	{"1.19rc1", "1.19rc2", -1},
	{"1.19.0", "1.19.1", -1},
	{"1.19rc1", "1.19.0", -1},
	{"1.19alpha3", "1.19beta2", -1},
	{"1.19beta2", "1.19rc1", -1},

	// Syntax we don't ever plan to use, but just in case we do.
	{"1.19.0-rc.1", "1.19.0-rc.2", -1},
	{"1.19.0-rc.1", "1.19.0", -1},
	{"1.19.0-alpha.3", "1.19.0-beta.2", -1},
	{"1.19.0-beta.2", "1.19.0-rc.1", -1},
}

func TestCompare(t *testing.T) {
	for _, tt := range compareTests {
		out := Compare(tt.x, tt.y)
		if out != tt.out {
			t.Errorf("Compare(%q, %q) = %d, want %d", tt.x, tt.y, out, tt.out)
		}
		out = Compare(tt.y, tt.x)
		if out != -tt.out {
			t.Errorf("Compare(%q, %q) = %d, want %d", tt.y, tt.x, out, -tt.out)
		}
	}
}

var toSemverTests = []struct {
	in  string
	out string
}{
	{"1", "v1.0.0"},
	{"1.2", "v1.2.0"},
	{"1.2.0", "v1.2.0"},
	{"1.2.3", "v1.2.3"},
	{"1.2alpha1", "v1.2.0-alpha.1"},
	{"1.2beta1", "v1.2.0-beta.1"},
	{"1.2rc1", "v1.2.0-rc.1"},
	{"bad", ""},

	// Syntax we don't ever plan to use, but just in case we do.
	{"1.19.0-rc.1", "v1.19.0-rc.1"},
}

func TestToSemver(t *testing.T) {
	for _, tt := range toSemverTests {
		if out := ToSemver(tt.in); out != tt.out {
			t.Errorf("ToSemver(%q) = %q, want %q", tt.in, out, tt.out)
		}
	}
}

var fromSemverTests = []struct {
	in  string
	out string
}{
	{"v1.0.0", "1"},
	{"v1.2.0", "1.2"},
	{"v1.21.0", "1.21.0"},
	{"v1.2.3", "1.2.3"},
	{"v1.2.0-alpha.1", "1.2alpha1"},
	{"v1.2.0-beta.1", "1.2beta1"},
	{"v1.2.0-rc.1", "1.2rc1"},
	{"bad", ""},
}

func TestFromSemver(t *testing.T) {
	for _, tt := range fromSemverTests {
		if out := FromSemver(tt.in); out != tt.out {
			t.Errorf("FromSemver(%q) = %q, want %q", tt.in, out, tt.out)
		}
	}
}

var majorTests = []struct {
	in  string
	out string
}{
	{"1.2.3", "1.2"},
	{"1.2", "1.2"},
	{"1", "1"},
}

func TestMajor(t *testing.T) {
	for _, tt := range majorTests {
		if out := Major(tt.in); out != tt.out {
			t.Errorf("Major(%q) = %q, want %q", tt.in, out, tt.out)
		}
	}
}

var prevTests = []struct {
	in  string
	out string
}{
	{"1.3rc4", "1.2"},
	{"1.3.5", "1.2"},
	{"1.3", "1.2"},
	{"1", "1"},
}

func TestPrev(t *testing.T) {
	for _, tt := range prevTests {
		if out := Prev(tt.in); out != tt.out {
			t.Errorf("Prev(%q) = %q, want %q", tt.in, out, tt.out)
		}
	}
}
