// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package doc

import (
	"go/parser"
	"go/token"
	"reflect"
	"strings"
	"testing"
)

func TestImportGroupStarts(t *testing.T) {
	for _, test := range []struct {
		name string
		in   string
		want []token.Pos
	}{
		{
			name: "one group",
			in: `package p
import (
	"a"
	"b"
	"c"
	"d"
)
`,
			want: []token.Pos{21},
		},
		{
			name: "several groups",
			in: `package p
import (
	"a"

	"b"
	"c"

	"d"
)
`,
			want: []token.Pos{21, 27, 38},
		},
		{
			name: "extra space",
			in: `package p
import (
	"a"


	"b"
	"c"


	"d"
)
`,
			want: []token.Pos{21, 28, 40},
		},
		{
			name: "line comment",
			in: `package p
import (
	"a" // comment
	"b" // comment

	"c"
)`,
			want: []token.Pos{21, 54},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", strings.NewReader(test.in), parser.ParseComments)
			if err != nil {
				t.Fatal(err)
			}
			got := findImportGroupStarts(file.Imports)
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("got %v, want %v", got, test.want)
			}
		})
	}

}
