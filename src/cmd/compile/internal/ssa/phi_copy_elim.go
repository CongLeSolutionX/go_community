// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

// phiAndCopyElim combine phielim and copyelim into a single pass.
func phiAndCopyElim(f *Func) {
	for {
		change := false
		for _, b := range f.Blocks {
			for _, v := range b.Values {
				copyelimValue(v)
				change = phielimValue(v) || change
			}
		}
		if !change {
			// Update block control values.
			for _, b := range f.Blocks {
				for i, v := range b.ControlValues() {
					if v.Op == OpCopy {
						b.ReplaceControl(i, v.Args[0])
					}
				}
			}

			// Update named values.
			for _, name := range f.Names {
				values := f.NamedValues[*name]
				for i, v := range values {
					if v.Op == OpCopy {
						values[i] = v.Args[0]
					}
				}
			}
			break
		}
	}
}
