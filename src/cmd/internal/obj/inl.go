// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package obj

import "cmd/internal/src"

// InlTree is a collection of inlined calls. The Parent field of an
// InlinedCall is used to point to another call in the tree via its
// index in the nodes slice.
//
// The compiler maintains a global inlining tree and adds a node to it
// every time a function is inlined. For example, suppose f() calls g()
// and g has two calls to h(), and that f, g, and h are inlineable.
// Assuming the global tree starts empty, recursively inlining a call
// to f would produce the following tree:
//
//   []InlinedCall{
//     {Parent: -1, Func: "f", Pos: ...},
//     {Parent:  0, Func: "g", Pos: ...},
//     {Parent:  1, Func: "h", Pos: ...},
//     {Parent:  1, Func: "h", Pos: ...},
//   }
//
// Eventually, the compiler extracts a per-function inlining tree from
// the global inlining tree (see pcln.go).
type InlTree struct {
	nodes []InlinedCall
}

// InlinedCall is a node in an InlTree.
type InlinedCall struct {
	Parent int      // index of the parent in the InlTree or -1 if outermost call
	Pos    src.XPos // position of the inlined call
	Func   *LSym    // function that was inlined
}

// Add adds a new call to the tree, returning its index.
func (tree *InlTree) Add(parent int, pos src.XPos, func_ *LSym) int {
	r := len(tree.nodes)
	call := InlinedCall{
		Parent: parent,
		Pos:    pos,
		Func:   func_,
	}
	tree.nodes = append(tree.nodes, call)
	return r
}

// OutermostPos returns the outermost position corresponding to xpos.
func (ctxt *Link) OutermostPos(xpos src.XPos) src.Pos {
	pos := ctxt.PosTable.Pos(xpos)
	ix := pos.Base().InliningIndex()
	if ix == -1 {
		return pos
	}

	outerxpos := xpos
	for ix >= 0 {
		call := ctxt.InlTree.nodes[ix]
		ix = call.Parent
		outerxpos = call.Pos
	}
	return ctxt.PosTable.Pos(outerxpos)
}
