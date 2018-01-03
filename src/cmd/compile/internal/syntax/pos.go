// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syntax

import "fmt"

const LineMax = 1<<32 - 1

type Pos struct {
	base      *PosBase
	line, col uint32
}

func MakePos(base *PosBase, line, col uint) Pos { return Pos{base, sat32(line), sat32(col)} }

func (pos Pos) IsKnown() bool  { return pos.line > 0 }
func (pos Pos) Base() *PosBase { return pos.base }
func (pos Pos) Line() uint     { return uint(pos.line) }
func (pos Pos) Col() uint      { return uint(pos.col) }
func (pos Pos) String() string { return fmt.Sprintf("%s:%d:%d", pos.base.Filename(), pos.line, pos.col) }

type PosBase struct {
	pos      Pos
	filename string
	line     uint
}

func NewFileBase(filename string) *PosBase                     { return &PosBase{filename: filename} }
func NewLineBase(pos Pos, filename string, line uint) *PosBase { return &PosBase{pos, filename, line} }

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
