// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package arm64

import (
	"strconv"
	"strings"
)

func (inst *inst) GoSyntax(code uint32) string {
	text := inst.goOp.String()
	var args []string
	for i, arg := range inst.args {
		args = append(args, arg.GoSyntax(inst, code, i))
	}
	if args != nil {
		text += " " + strings.Join(args, ", ")
	}
	return text
}

func (inst *inst) GNUSyntax(code uint32) string {
	text := inst.armOp.String()
	var args []string
	for i, arg := range inst.gnuOrderArgs() {
		args = append(args, arg.GNUSyntax(inst, code, i))
	}
	if args != nil {
		text += " " + strings.Join(args, ", ")
	}
	return text
}

func (inst *inst) gnuOrderArgs() []arg {
	var args []arg
	return args
}

func (arg arg) GoSyntax(inst *inst, code uint32, idx int) string {
	text := ""
	switch arg.aType {
	case AC_REG:
		switch arg.elms[0] {
		case sa_wa__Ra:
			ra := (code >> 10) & 0x1f
			if ra == 31 {
				text = "ZR"
			} else {
				text = "R" + strconv.Itoa(int(ra))
			}
		}
	}
	return text
}

func (arg arg) GNUSyntax(inst *inst, code uint32, idx int) string {
	text := ""
	switch arg.aType {
	case AC_REG:
		switch arg.elms[0] {
		case sa_wa__Ra:
			ra := (code >> 10) & 0x1f
			if ra == 31 {
				text = "wzr"
			} else {
				text = "w" + strconv.Itoa(int(ra))
			}
		}
	}
	return text
}
