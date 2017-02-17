// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package obj

import "cmd/internal/src"

type InlTree struct {
	nodes []InlinedCall
}

type InlinedCall struct {
	Parent int32
	Pos    src.XPos
	Func   *LSym // function that was inlined
}

func (tree *InlTree) Add(parent int32, pos src.XPos, func_ *LSym) int32 {
	r := len(tree.nodes)
	call := InlinedCall{
		Parent: parent,
		Pos:    pos,
		Func:   func_,
	}
	tree.nodes = append(tree.nodes, call)
	return int32(r)
}

func (ctxt *Link) OutermostLine(xpos src.XPos) string {
	pos := ctxt.PosTable.Pos(xpos)
	ix := pos.Base().InlIndex()
	if ix == -1 {
		return pos.String()
	}

	var outerxpos = xpos
	for ix != -1 {
		call := ctxt.InlTree.nodes[ix]
		ix = call.Parent
		outerxpos = call.Pos
	}
	return ctxt.PosTable.Pos(outerxpos).String()
}
