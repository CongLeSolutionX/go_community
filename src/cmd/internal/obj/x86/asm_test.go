// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package x86

import (
	"cmd/internal/obj"
	"testing"
)

func init() {
	// Required for tests that access any of
	// opindex/ycover/reg/regrex global tables.
	var ctxt obj.Link
	instinit(&ctxt)
}

func TestRegisterListEncDec(t *testing.T) {
	tests := []struct {
		printed string
		reg0    int16
		reg1    int16
		kind    rlistKind
	}{
		{"[R10-R13]", REG_R10, REG_R13, rlistInvalid},
		{"[X0-AX]", REG_X0, REG_AX, rlistInvalid},

		{"[X0-X3]", REG_X0, REG_X3, rlistXmm},
		{"[X21-X24]", REG_X21, REG_X24, rlistXmm},

		{"[Y0-Y3]", REG_Y0, REG_Y3, rlistYmm},
		{"[Y21-Y24]", REG_Y21, REG_Y24, rlistYmm},

		{"[Z0-Z3]", REG_Z0, REG_Z3, rlistZmm},
		{"[Z21-Z24]", REG_Z21, REG_Z24, rlistZmm},
	}

	for _, test := range tests {
		enc := EncodeRegisterRange(test.reg0, test.reg1)
		dec := decodeRegisterRange(enc)

		if dec.kind != test.kind {
			t.Errorf("%s kind mismatch: have %d, want %d",
				test.printed, dec.kind, test.kind)
		}
		wantLo := regIndex(test.reg0)
		wantHi := regIndex(test.reg1)
		if dec.lo != int16(wantLo) {
			t.Errorf("%s lo mismatch: have %d, want %d",
				test.printed, dec.lo, wantLo)
		}
		if dec.hi != int16(wantHi) {
			t.Errorf("%s hi mismatch: have %d, want %d",
				test.printed, dec.hi, wantHi)
		}
		wantPrinted := "[R???-R???]"
		if test.kind != rlistInvalid {
			wantPrinted = test.printed
		}
		if rlconv(enc) != wantPrinted {
			t.Errorf("%s string mismatch: have %s, want %s",
				test.printed, rlconv(enc), wantPrinted)
		}
		if test.kind != rlistInvalid {
			if dec.Reg0() != int(test.reg0) {
				t.Errorf("%s reg0 mismatch: have %s, want %s",
					test.printed, rconv(dec.Reg0()), rconv(int(test.reg0)))
			}
			if dec.Reg1() != int(test.reg1) {
				t.Errorf("%s reg1 mismatch: have %s, want %s",
					test.printed, rconv(dec.Reg1()), rconv(int(test.reg1)))
			}
		}
	}
}

func TestRegIndex(t *testing.T) {
	tests := []struct {
		regFrom int
		regTo   int
	}{
		{REG_AL, REG_R15B},
		{REG_AX, REG_R15},
		{REG_M0, REG_M7},
		{REG_K0, REG_K7},
		{REG_X0, REG_X31},
		{REG_Y0, REG_Y31},
		{REG_Z0, REG_Z31},
	}

	for _, test := range tests {
		for index, reg := 0, test.regFrom; reg <= test.regTo; index, reg = index+1, reg+1 {
			have := regIndex(int16(reg))
			want := index
			if have != want {
				regName := rconv(int(reg))
				t.Errorf("regIndex(%s):\nhave: %d\nwant: %d",
					regName, have, want)
			}
		}
	}
}
