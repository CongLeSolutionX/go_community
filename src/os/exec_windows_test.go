// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os_test

import (
	. "os"
	"syscall"
	"testing"
	"unsafe"
)

func TestCommandLineToArgv(t *testing.T) {
	cmds := []string{
		`test`,
		`test a b c`,
		`test "`,
		`test ""`,
		`test """`,
		`test "" a`,
		`test "123"`,
		`test \"123\"`,
		`test \"123 456\"`,
		`test \\"`,
		`test \\\"`,
		`test \\\\\"`,
		`test \\\"x`,
		`test """"\""\\\"`,
		`"cmd line" abc`,
		`test \\\\\""x"""y z`,
		"test\tb\t\"x\ty\"",
	}

	for _, cmd := range cmds {
		var argc int32
		c, err := syscall.CommandLineToArgv(&syscall.StringToUTF16(cmd)[0], &argc)
		if err != nil {
			t.Fatal(err)
		}

		out := CommandLineToArgv(cmd)
		outwin := make([]string, len(out))

		valid := len(outwin) == len(out)
		for i := range outwin {
			outwin[i] = syscall.UTF16ToString(c[i][:])
			if i < len(out) && out[i] != outwin[i] {
				valid = false
			}
		}

		if !valid {
			t.Errorf("%#v: %#v vs %#v", cmd, out, outwin)
		}

		syscall.LocalFree(syscall.Handle(unsafe.Pointer(c)))
	}
}
