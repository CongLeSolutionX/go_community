// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package arm64

import (
	"cmd/internal/obj"
	"log"
)

// This file contains the encoding of the each element type.

// encodeElm encodes an element. It is worth noting that the encoding of an element
// is also the checking process of an instruction, returns false if the check fails.
func (c *ctxt7) encodeElm(p *obj.Prog, bin uint32, ag *obj.Addr, instIdx, oprIdx, elmIdx int, checks map[uint32]uint32) (uint32, bool) {
	ai := instTab[instIdx].args[oprIdx]
	e := ai.elms[elmIdx]
	enc := uint32(0)
	switch e {
	default:
		log.Fatalf("unimplemented element type %s: %v\n", e, p)
	}
	return enc, true
}

// encodeArgs encodes the argument ag of p.
func (c *ctxt7) encodeArg(p *obj.Prog, bin uint32, ag *obj.Addr, instIdx, oprIdx int, checks map[uint32]uint32) (uint32, bool) {
	ai := instTab[instIdx].args[oprIdx]
	enc := uint32(0)
	for i := range ai.elms {
		if v, ok := c.encodeElm(p, bin, ag, instIdx, oprIdx, i, checks); !ok {
			return 0, false
		} else {
			enc |= v
		}
	}
	return enc, true
}
