// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ir

type Call struct {
	defaultNode
	nodeFieldOp
	nodeFieldLeft
	nodeFieldRight
	nodeFieldType
	nodeFieldIsDDD
	nodeFieldList
	nodeFieldRlist
	nodeFieldNoInline
	nodeFieldHasCall
	nodeFieldOpt
	nodeFieldNonNil
	nodeFieldNinit
	nodeFieldTransient
	nodeFieldNbody
	nodeFieldBounded
}

func (c *Call) RawCopy() INode { copy := *c; return &copy }

func (c *Call) SetOp(op Op) {
	switch op {
	case OCALL, OCALLMETH, OCALLINTER:
		c.op = op
	default:
		panic("Call SetOp " + op.String())
	}
}
