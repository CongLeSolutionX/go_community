// asmcheck

// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package codegen

// Check that the method wrapper with pointer receivers (in both the wrapper and the method) uses tail call.

// 386,amd64,arm,arm64,loong,loong64,mipsx,mips64x,ppc64x,s390x:`JMP\tcommand-line-arguments\.\(\*Foo\)\.GetVals\(SB\)`
// riscv64:`JAL\tX0, command-line-arguments\.\(\*Foo\)\.GetVals\(SB\)`
func (f *Foo) GetVals() [2]int { return [2]int{f.Val, f.Val + 1} }

type Foo struct{ Val int }

type Bar struct {
	int64
	*Foo // needs a method wrapper
	string
}

var i any

func init() {
	i = Bar{1, nil, "first"}
}
