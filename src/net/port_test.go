// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import (
	"testing"
)

var parsePortTests = []struct {
	service     string
	port        int
	needsLookup bool
}{
	{"", 0, false},
	{"42", 42, false},

	{"123456789", 123456789, false},
	{"-123456789", -123456789, false},
	{"-1", -1, false},
	{"0", 0, false},

	{"abc", 0, true},
	{"9pfs", 0, true},
	{"123badport", 0, true},
	{"bad123port", 0, true},
	{"badport123", 0, true},
	{"123456789badport", 0, true},
}

func TestParsePort(t *testing.T) {
	for _, tt := range parsePortTests {
		if port, needsLookup := parsePort(tt.service); port != tt.port || needsLookup != tt.needsLookup {
			t.Errorf("parsePort(%q) = %d, %t; want %d, %t", tt.service, port, needsLookup, tt.port, tt.needsLookup)
		}
	}
}
