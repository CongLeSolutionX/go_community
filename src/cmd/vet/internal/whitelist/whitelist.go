// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package whitelist defines exceptions for the vet tool.
package whitelist

// UnkeyedListeral contains frozen struct types from standard packages.
var UnkeyedLiteral = map[string]bool{
	"cmd/compile/internal/gc.Branch":           true,
	"cmd/compile/internal/gc.FloatingEQNEJump": true,
	"cmd/compile/internal/ssa.ExternSymbol":    true,
	"cmd/compile/internal/ssa.LocalSlot":       true,

	"image/color.Alpha16": true,
	"image/color.Alpha":   true,
	"image/color.CMYK":    true,
	"image/color.Gray16":  true,
	"image/color.Gray":    true,
	"image/color.NRGBA64": true,
	"image/color.NRGBA":   true,
	"image/color.NYCbCrA": true,
	"image/color.RGBA64":  true,
	"image/color.RGBA":    true,
	"image/color.YCbCr":   true,
	"image.Point":         true,
	"image.Rectangle":     true,
	"image.Uniform":       true,

	"unicode.Range16": true,
}
