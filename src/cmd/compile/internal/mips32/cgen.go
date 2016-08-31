// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mips32

import (
	"cmd/compile/internal/gc"
	"cmd/internal/obj"
	"cmd/internal/obj/mips32"
)

/*
 * generate array index into res.
 * n might be any size; res is 32-bit.
 * returns Prog* to patch to panic call.
 */
func cgenindex(n *gc.Node, res *gc.Node, bounded bool) *obj.Prog {
	if !gc.Is64(n.Type) {
		gc.Cgen(n, res)
		return nil
	}

	var tmp gc.Node
	gc.Tempname(&tmp, gc.Types[gc.TINT64])
	gc.Cgen(n, &tmp)
	var lo gc.Node
	var hi gc.Node
	split64(&tmp, &lo, &hi)
	gmove(&lo, res)
	if bounded {
		splitclean()
		return nil
	}

	var n1 gc.Node
	gc.Regalloc(&n1, gc.Types[gc.TINT32], nil)
	var n2 gc.Node
	gc.Regalloc(&n2, gc.Types[gc.TINT32], nil)
	var zero gc.Node
	gc.Nodconst(&zero, gc.Types[gc.TINT32], 0)
	gmove(&hi, &n1)
	gmove(&zero, &n2)
	gc.Regfree(&n2)
	gc.Regfree(&n1)
	splitclean()
	return ginsbranch(mips32.ABNE, nil, &n1, &n2, -1)

}

func blockcopy(n, res *gc.Node, osrc, odst, w int64) {
	// determine alignment.
	// want to avoid unaligned access, so have to use
	// smaller operations for less aligned types.
	// for example moving [4]byte must use 4 MOVB not 1 MOVW.
	align := int(n.Type.Align)

	var op obj.As
	switch align {
	default:
		gc.Fatalf("sgen: invalid alignment %d for %v", align, n.Type)

	case 1:
		op = mips32.AMOVB

	case 2:
		op = mips32.AMOVH

	case 4:
		op = mips32.AMOVW

	case 8:
		op = mips32.AMOVW
	}

	if w%int64(align) != 0 {
		gc.Fatalf("cgen: unaligned size %d (align=%d) for %v", w, align, n.Type)
	}
	c := int32(w / int64(align))

	// if we are copying forward on the stack and
	// the src and dst overlap, then reverse direction
	dir := align

	if osrc < odst && int64(odst) < int64(osrc)+w {
		dir = -dir
	}

	var dst gc.Node
	var src gc.Node
	if n.Ullman >= res.Ullman {
		gc.Agenr(n, &dst, res) // temporarily use dst
		gc.Regalloc(&src, gc.Types[gc.Tptr], nil)
		gins(mips32.AMOVW, &dst, &src)
		if res.Op == gc.ONAME {
			gc.Gvardef(res)
		}
		gc.Agen(res, &dst)
	} else {
		if res.Op == gc.ONAME {
			gc.Gvardef(res)
		}
		gc.Agenr(res, &dst, res)
		gc.Agenr(n, &src, nil)
	}

	var tmp gc.Node
	gc.Regalloc(&tmp, gc.Types[gc.Tptr], nil)

	// set up end marker
	var nend gc.Node

	// move src and dest to the end of block if necessary
	if dir < 0 {
		if c >= 4 {
			gc.Regalloc(&nend, gc.Types[gc.Tptr], nil)
			gins(mips32.AMOVW, &src, &nend)
		}

		p := gins(mips32.AADD, nil, &src)
		p.From.Type = obj.TYPE_CONST
		p.From.Offset = w

		p = gins(mips32.AADD, nil, &dst)
		p.From.Type = obj.TYPE_CONST
		p.From.Offset = w
	} else {
		p := gins(mips32.AADD, nil, &src)
		p.From.Type = obj.TYPE_CONST
		p.From.Offset = int64(-dir)

		p = gins(mips32.AADD, nil, &dst)
		p.From.Type = obj.TYPE_CONST
		p.From.Offset = int64(-dir)

		if c >= 4 {
			gc.Regalloc(&nend, gc.Types[gc.Tptr], nil)
			p := gins(mips32.AMOVW, &src, &nend)
			p.From.Type = obj.TYPE_ADDR
			p.From.Offset = w
		}
	}

	// move
	// TODO(mips32): enable duffcopy for larger copies.
	if c >= 4 {
		p := gins(op, &src, &tmp)
		p.From.Type = obj.TYPE_MEM
		p.From.Offset = int64(dir)
		ploop := p

		p = gins(mips32.AADD, nil, &src)
		p.From.Type = obj.TYPE_CONST
		p.From.Offset = int64(dir)

		p = gins(op, &tmp, &dst)
		p.To.Type = obj.TYPE_MEM
		p.To.Offset = int64(dir)

		p = gins(mips32.AADD, nil, &dst)
		p.From.Type = obj.TYPE_CONST
		p.From.Offset = int64(dir)

		gc.Patch(ginsbranch(mips32.ABNE, nil, &src, &nend, 0), ploop)
		gc.Regfree(&nend)
	} else {

		var p *obj.Prog
		for ; c > 0; c-- {
			p = gins(op, &src, &tmp)
			p.From.Type = obj.TYPE_MEM
			p.From.Offset = int64(dir)

			p = gins(mips32.AADD, nil, &src)
			p.From.Type = obj.TYPE_CONST
			p.From.Offset = int64(dir)

			p = gins(op, &tmp, &dst)
			p.To.Type = obj.TYPE_MEM
			p.To.Offset = int64(dir)

			p = gins(mips32.AADD, nil, &dst)
			p.From.Type = obj.TYPE_CONST
			p.From.Offset = int64(dir)
		}
	}

	gc.Regfree(&dst)
	gc.Regfree(&src)
	gc.Regfree(&tmp)
}
