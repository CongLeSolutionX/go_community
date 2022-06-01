// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

import "cmd/internal/src"

func hardwiredZero(f *Func) {
	if !f.Config.hasZeroReg {
		return
	}

	var zero *Value
	for _, b := range f.Blocks {
		for _, v := range b.Values {
			switch v.Op {
			case OpRISCV64MOVDconst, OpLOONG64MOVVconst:
				// ok
			default:
				continue
			}
			if v.AuxInt == 0 {
				if zero == nil {
					zero = f.Entry.NewValue0(src.NoXPos, OpHardwiredZero, f.Config.Types.Int)
				}
				v.copyOf(zero)
			}
		}
	}
}
