// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build checknewoldreassignment

package ir

import (
	"cmd/compile/internal/base"
	"cmd/internal/src"
	"fmt"
	"path/filepath"
	"strings"
)

func checkStaticValueResult(n Node, newres Node) {
	oldres := StaticValue(n)
	if oldres != newres {
		base.Fatalf("%s: new/old static value disagreement on %v:\nnew=%v\nold=%v", fmtFullPos(n.Pos()), n, newres, oldres)
	}
}

func checkReassignedResult(n Node, newres bool) {
	origres := Reassigned(n)
	if newres != origres {
		base.Fatalf("%s: new/old reassigned disagreement on %v (class %s) newres=%v oldres=%v", fmtFullPos(n.Pos()), n, n.Class.String(), newres, oldres)
	}
}

func fmtFullPos(p src.XPos) string {
	var sb strings.Builder
	sep := ""
	base.Ctxt.AllPos(p, func(pos src.Pos) {
		fmt.Fprintf(&sb, sep)
		sep = "|"
		file := filepath.Base(pos.Filename())
		fmt.Fprintf(&sb, "%s:%d:%d", file, pos.Line(), pos.Col())
	})
	return sb.String()
}
