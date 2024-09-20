// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package arm64

import (
	"testing"
)

func TestARM64SVERegister(t *testing.T) {
	r := NewSVERegister(REG_Z2, EXT_NONE)

	r.SetExt(EXT_B)
	if r.Ext() != EXT_B {
		t.Fail()
		t.Logf("incorrect extension: %b, expected: %b", r.Ext(), EXT_B)
	}
	if r.Format() != REG_Z|EXT_B {
		t.Fail()
		t.Logf("incorrect format: %b expected: %b", r.Format(), REG_Z|EXT_B)
	}
	if getType(r.Format()) != REG_Z {
		t.Fail()
		t.Logf("incorrect format type: %b expected: %b", getType(r.Format()), REG_Z)
	}
	if getExt(r.Format()) != EXT_B {
		t.Fail()
		t.Logf("incorrect format ext: %b expected: %b", getExt(r.Format()), EXT_B)
	}
	if r.Number() != 2 {
		t.Fail()
		t.Logf("incorrect register number %d != %d", r.Number(), 2)
	}
}

func TestARM64IsSVERegister(t *testing.T) {
	if IsSVERegister(REG_R10) {
		t.Fail()
	}
	r := NewSVERegister(REG_Z14, EXT_D)
	if !IsSVERegister(r.ToInt16()) {
		t.Fail()
	}
	r2 := NewSVERegister(REG_R14, EXT_NONE)
	if !IsSVERegister(r2.ToInt16()) {
		t.Fail()
	}
}
