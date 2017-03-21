// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syntax

import "fmt"

// A Pos is a light-weight (line, col) pair representing a source position.
type Pos uint32

const (
	lineBits = 24
	colBits  = 32 - lineBits
	lineMask = 1<<lineBits - 1
	colMask  = 1<<colBits - 1
)

func MakePos(line, col uint) Pos {
	return Pos(saturate(line, lineMask)<<colBits | saturate(col, colMask))
}
func (p Pos) Line() uint     { return uint(p >> colBits) }
func (p Pos) Col() uint      { return uint(p & colMask) }
func (p Pos) IsKnown() bool  { return p != 0 }
func (p Pos) String() string { return fmt.Sprintf("%d:%d", p.Line(), p.Col()) }

func saturate(x, max uint) uint {
	if x > max {
		return max
	}
	return x
}
