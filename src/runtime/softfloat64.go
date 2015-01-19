// Copyright 2010 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Software IEEE754 64-bit floating point.
// Only referred to (and thus linked in) by arm port
// and by tests in this directory.

package runtime

import (
	_fp "runtime/internal/fp"
)

func fcmp64(f, g uint64) (cmp int32, isnan bool) {
	fs, fm, _, fi, fn := _fp.Funpack64(f)
	gs, gm, _, gi, gn := _fp.Funpack64(g)

	switch {
	case fn, gn: // flag NaN
		return 0, true

	case !fi && !gi && fm == 0 && gm == 0: // ±0 == ±0
		return 0, false

	case fs > gs: // f < 0, g > 0
		return -1, false

	case fs < gs: // f > 0, g < 0
		return +1, false

	// Same sign, not NaN.
	// Can compare encodings directly now.
	// Reverse for sign.
	case fs == 0 && f < g, fs != 0 && f > g:
		return -1, false

	case fs == 0 && f > g, fs != 0 && f < g:
		return +1, false
	}

	// f == g
	return 0, false
}
func fsub64c(f, g uint64, ret *uint64)              { *ret = _fp.Fsub64(f, g) }
func fmul64c(f, g uint64, ret *uint64)              { *ret = _fp.Fmul64(f, g) }
func fdiv64c(f, g uint64, ret *uint64)              { *ret = _fp.Fdiv64(f, g) }
func f32to64c(f uint32, ret *uint64)                { *ret = _fp.F32to64(f) }
func f64to32c(f uint64, ret *uint32)                { *ret = _fp.F64to32(f) }
func fcmp64c(f, g uint64, ret *int32, retnan *bool) { *ret, *retnan = fcmp64(f, g) }
func fintto64c(val int64, ret *uint64)              { *ret = _fp.Fintto64(val) }
func f64tointc(f uint64, ret *int64, retok *bool)   { *ret, *retok = _fp.F64toint(f) }
