// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package profile

import (
	"strings"
	"testing"
)

func TestParseError(t *testing.T) {
	testcases := []string{
		"",
		"garbage text",
		"\x1f\x8b", // truncated gzip header
		"\x1f\x8b\x08\x08\xbe\xe9\x20\x58\x00\x03\x65\x6d\x70\x74\x79\x00\x03\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00", // empty gzipped file
	}

	for i, input := range testcases {
		_, err := Parse(strings.NewReader(input))
		if err == nil {
			t.Errorf("got nil, want error for input #%d", i)
		}
	}
}
