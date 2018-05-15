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
