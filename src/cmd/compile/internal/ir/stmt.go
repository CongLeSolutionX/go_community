// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ir

import "cmd/compile/internal/types"

type ContinueStmt struct {
	defaultNode
	Label *types.Sym // label
}

func (*ContinueStmt) Op() Op                  { return OCONTINUE }
func (n *ContinueStmt) Sym() *types.Sym       { return n.Label }
func (n *ContinueStmt) SetSym(sym *types.Sym) { n.Label = sym }
func (n *ContinueStmt) RawCopy() INode {
	m := *n
	return &m
}
