// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package inlheur

import (
	"fmt"
	"strings"
)

// TODO: convert the following two functions to a single generic
// function when bootstrap go compiler supports generics.

func paramFlagSliceToSB(sb *strings.Builder, sl []ParamPropBits, prefix string) {
	var sb2 strings.Builder
	foundnz := false
	fmt.Fprintf(&sb2, "%sRecvrParamFlags:\n", prefix)
	for i, e := range sl {
		if e != 0 {
			foundnz = true
		}
		fmt.Fprintf(&sb2, "%s  %d: %s\n", prefix, i, e.String())
	}
	if foundnz {
		sb.WriteString(sb2.String())
	}
}

func returnFlagSliceToSB(sb *strings.Builder, sl []ReturnPropBits, prefix string) {
	var sb2 strings.Builder
	foundnz := false
	fmt.Fprintf(&sb2, "%sReturnFlags:\n", prefix)
	for i, e := range sl {
		if e != 0 {
			foundnz = true
		}
		fmt.Fprintf(&sb2, "%s  %d: %s\n", prefix, i, e.String())
	}
	if foundnz {
		sb.WriteString(sb2.String())
	}
}

func (fp *FuncProps) String() string {
	return fp.ToString("")
}

func (fp *FuncProps) ToString(prefix string) string {
	var sb strings.Builder
	if fp.Flags != 0 {
		fmt.Fprintf(&sb, "%sFlags: %s\n", prefix, fp.Flags)
	}
	paramFlagSliceToSB(&sb, fp.RecvrParamFlags, prefix)
	returnFlagSliceToSB(&sb, fp.ReturnFlags, prefix)
	return sb.String()
}
