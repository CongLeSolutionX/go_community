// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package arm64

import "cmd/internal/obj"

// This file contains the encoding implementation of the argument type.

// encodeArgs encodes the argument arg of p.
func (c *ctxt7) encodeArg(p *obj.Prog, arg *obj.Addr, atyp arg) uint32 {
	// TODO: enable more types.
	c.ctxt.Diag("unimplemented argument type %v: %v\n", atyp, p)
	return 0
}

// encodeOpcode encodes the opcode. The opcode of some special
// instructions affects the encoding, for most instructions
// this function just return 0.
// TODO enable this function when necessary.
/*
func (c *ctxt7) encodeOpcode(a obj.As) uint32 {
	switch a {
	default:
		return 0
	}
}
*/
