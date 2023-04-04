// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pprof

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPEBuildID(t *testing.T) {
	tests := []struct {
		data, expected string
	}{
		{
			data:     "abcdefg\xff Go build ID: \"ABCDE\"\n \xffhijklm",
			expected: "ABCDE",
		},
		{
			data:     "abcdefghijklm",
			expected: "533efedecad144f7",
		},
	}

	for _, test := range tests {
		path := filepath.Join(t.TempDir(), "pe")
		err := os.WriteFile(path, []byte(test.data), 0644)
		if err != nil {
			t.Fatal(err)
		}
		id := peBuildID(path)
		if id != test.expected {
			t.Fatalf("got %q expected %q", id, test.expected)
		}
	}
}
