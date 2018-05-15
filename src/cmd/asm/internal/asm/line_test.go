// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package asm

import (
	"cmd/asm/internal/lex"
	"strings"
	"testing"
)

type badLineTest struct {
	input string
	error string
}

func TestAMD64BadLineParser(t *testing.T) {
	testBadLineParser(t, "amd64", []badLineTest{
		// Test AVX512 suffixes.
		{"VADDPD.A X0, X1, X2", `parse suffix: unknown "A"`},
		{"VADDPD.A.A X0, X1, X2", `parse suffix: unknown "A"; duplicate "A"`},
		{"VADDPD.A.A.A X0, X1, X2", `parse suffix: unknown "A"; duplicate "A"`},
		{"VADDPD.A.B X0, X1, X2", `parse suffix: unknown "A"; unknown "B"`},
		{"VADDPD.Z.A X0, X1, X2", `parse suffix: Z suffix should be the last; unknown "A"`},
		{"VADDPD.Z.Z X0, X1, X2", `parse suffix: Z suffix should be the last; duplicate "Z"`},
		{"VADDPD.SAE.BCST X0, X1, X2", `parse suffix: can't combine rounding/SAE and broadcast`},
		{"VADDPD.BCST.SAE X0, X1, X2", `parse suffix: can't combine rounding/SAE and broadcast`},
		{"VADDPD.BCST.Z.SAE X0, X1, X2", `parse suffix: Z suffix should be the last; can't combine rounding/SAE and broadcast`},
		{"VADDPD.SAE.SAE X0, X1, X2", `parse suffix: duplicate "SAE"`},
		{"VADDPD.RZ_SAE.SAE X0, X1, X2", `parse suffix: bad suffix combination`},
	})
}

func testBadLineParser(t *testing.T, goarch string, tests []badLineTest) {
	for i, test := range tests {
		architecture, ctxt := setArch(goarch)
		tokenizer := lex.NewTokenizer("", strings.NewReader(test.input+"\n"), nil)
		parser := NewParser(ctxt, architecture, tokenizer)

		err := tryParse(t, func() {
			parser.start(lex.Tokenize(test.input))
			parser.line()
		})

		switch {
		case err == nil:
			t.Errorf("#%d: %q: want error %q; have none", i, test.input, test.error)
		case !strings.Contains(err.Error(), test.error):
			t.Errorf("#%d: %q: want error %q; have %q", i, test.input, test.error, err)
		}
	}
}
