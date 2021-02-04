// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package arm64

// This order should be strictly consistent to that in a.out.go
// TODO: generate these strings automatically.
var cnames7 = []string{
	"NONE",
	"REG",
	"RSP",
	"FREG",
	"VREG",
	"PAIR",
	"SHIFT",
	"EXTREG",
	"SPR",
	"SPOP",
	"COND",
	"ARNG",
	"ELEM",
	"LIST",
	"ZCON",
	"LCON",
	"VCON",
	"FCON",
	"MOVCONZ",
	"MOVCONN",
	"VCONADDR",
	"LACON",
	"SBRA",
	"ZOREG",
	"LOREG",
	"VOREG",
	"ROFF",
	"ADDR",
	"GOTADDR",
	"TLS_LE",
	"TLS_IE",
	"GOK",
	"TEXTSIZE",
	"NCLASS",
}
