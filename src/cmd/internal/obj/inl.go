// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package obj

import "cmd/internal/src"

// InlTree is a collection of inlined calls. The calls form a tree using the
// Parent field of InlinedCall. For example, suppose f() calls g() and g
// has two calls to h(). If f, g, and h are inlineable, a call to f would
// produce this inlining tree:
//
//   []InlinedCall{
//     {Parent: -1, Func: "f", Pos: ...},
//     {Parent:  0, Func: "g", Pos: ...},
//     {Parent:  1, Func: "h", Pos: ...},
//     {Parent:  1, Func: "h", Pos: ...},
//   }
//
type InlTree struct {
	nodes []InlinedCall
}

// InlinedCall is a node in an InlTree.
type InlinedCall struct {
	Parent int32    // index of the parent in the InlTree or -1 if outermost call
	Pos    src.XPos // position of the inlined call
	Func   *LSym    // function that was inlined
}

// Add unconditionally adds a new call to the tree, returning its index.
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

// OutermostPos returns the outermost position corresponding to xpos.
func (ctxt *Link) OutermostPos(xpos src.XPos) src.Pos {
	pos := ctxt.PosTable.Pos(xpos)
	ix := pos.Base().InlIndex()
	if ix == -1 {
		return pos
	}

	outerxpos := xpos
	for ix != -1 {
		call := ctxt.InlTree.nodes[ix]
		ix = call.Parent
		outerxpos = call.Pos
	}
	return ctxt.PosTable.Pos(outerxpos)
}
