// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

import (
	"fmt"
	"strings"

	"cmd/internal/src"
)

// columns reports positions that appear in distinct basic blocks.
func columns(f *Func) {
	if f.pass.debug == 0 {
		return
	}

	m := make(map[src.XPos]map[*Block]struct{})

	for j := 0; j < len(f.Blocks); j++ {
		b := f.Blocks[j]

		for _, v := range b.Values {
			pos := v.Pos.WithNotStmt() // Don't care about stmt bit.

			set := m[pos]
			if set == nil {
				set = make(map[*Block]struct{})
			}
			set[b] = struct{}{}
			m[pos] = set
		}
	}

	for pos, set := range m {
		if len(set) < 2 {
			continue
		}

		var sb strings.Builder
		for b := range set {
			fmt.Fprintf(&sb, "%v ", b)
		}

		fmt.Printf("%s: %v: appears in %s\n", f.Name, f.Config.ctxt.InnermostPos(pos), sb.String())
	}
}
