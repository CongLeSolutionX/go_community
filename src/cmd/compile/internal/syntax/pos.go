// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syntax

import "fmt"

const LineMax = 1<<32 - 1

// A Pos represents an absolute (line, col) source position
// with a reference to the containing file or most recent
// line directive.
// Pos values are intentionally light-weight so that they
// can be created without too much concern about space use.
type Pos struct {
	base      *PosBase
	line, col uint32
}

// MakePos returns a new Pos for the given PosBase, line and column.
func MakePos(base *PosBase, line, col uint) Pos { return Pos{base, sat32(line), sat32(col)} }

func (pos Pos) IsKnown() bool  { return pos.line > 0 }
func (pos Pos) Base() *PosBase { return pos.base }
func (pos Pos) Line() uint     { return uint(pos.line) }
func (pos Pos) Col() uint      { return uint(pos.col) }
func (pos Pos) String() string { return fmt.Sprintf("%s:%d:%d", pos.base.Filename(), pos.line, pos.col) }

// A PosBase represents a source file (name) or line directive within a source file.
type PosBase struct {
	pos       Pos
	filename  string
	line, col uint
}

func NewFileBase(filename string) *PosBase {
	return &PosBase{filename: filename}
}

func NewLineBase(pos Pos, filename string, line, col uint) *PosBase {
	return &PosBase{pos, filename, line, col}
}

func (base *PosBase) Pos() (_ Pos) {
	if base == nil {
		return
	}
	return base.pos
}

func (base *PosBase) Filename() string {
	if base == nil {
		return ""
	}
	return base.filename
}

func (base *PosBase) Line() uint {
	if base == nil {
		return 0
	}
	return base.line
}

func sat32(x uint) uint32 {
	if x > LineMax {
		return LineMax
	}
	return uint32(x)
}
