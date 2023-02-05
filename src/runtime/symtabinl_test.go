// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	"internal/abi"
	"runtime/internal/sys"
)

func XTestInlineUnwinder(t T) {
	pc1 := abi.FuncPCABIInternal(tiwTest)
	f := findfunc(pc1)
	if !f.valid() {
		t.Fatalf("failed to resolve tiwTest at PC %#x", pc1)
	}

	want := map[string]int{
		"tiwInlined1:3 tiwTest:10":               0,
		"tiwInlined1:3 tiwInlined2:6 tiwTest:11": 0,
		"tiwInlined2:7 tiwTest:11":               0,
		"tiwTest:12":                             0,
	}
	wantStart := map[string]int{
		"tiwInlined1": 2,
		"tiwInlined2": 5,
		"tiwTest":     9,
	}

	// Iterate over the PCs in tiwTest and walk the inline stack for each.
	prevStack := "x"
	var cache pcvalueCache
pcLoop:
	for pc := pc1; pc < pc1+1024; {
		stack := ""
		for u := newInlineUnwinder(f, pc, &cache); u.valid(); u.next() {
			file, line := u.fileLine()
			if file == "?" {
				// This is a good indicator we're past the end of the function.
				//
				// TODO: If we ever have function end information, use that.
				break pcLoop
			}

			const wantFile = "symtabinl_test.go"
			if len(file) < len(wantFile) || file[len(file)-len(wantFile):] != wantFile {
				t.Errorf("tiwTest+%#x: want file ...%s, got %s", pc-pc1, wantFile, file)
			}

			sf := u.srcFunc()

			name := sf.name()
			const namePrefix = "runtime."
			if hasPrefix(name, namePrefix) {
				name = name[len(namePrefix):]
			}
			if !hasPrefix(name, "tiw") {
				t.Errorf("tiwTest+%#x: unexpected function %s", pc-pc1, name)
			}

			start := int(sf.startLine) - tiwStart
			if start != wantStart[name] {
				t.Errorf("tiwTest+%#x: want startLine %d, got %d", pc-pc1, wantStart[name], start)
			}
			if sf.funcID != funcID_normal {
				t.Errorf("tiwTest+%#x: bad funcID %v", pc-pc1, sf.funcID)
			}

			stack = FmtSprintf("%s %s:%d", stack, name, line-tiwStart)
		}
		stack = stack[1:]

		if stack != prevStack {
			prevStack = stack

			t.Logf("tiwTest+%#x: %s", pc-pc1, stack)

			if _, ok := want[stack]; ok {
				want[stack]++
			}
		}

		pc += sys.PCQuantum
		// Check if we're still in tiwTest
		if findfunc(pc) != f {
			break
		}
	}

	// Check that we got all the stacks we wanted.
	for stack, count := range want {
		if count == 0 {
			t.Errorf("missing stack %s", stack)
		}
	}
}

func lineNumber() int {
	_, _, line, _ := Caller(1)
	return line // return 0 for error
}

var tiwStart = lineNumber() // +0
var tiw1, tiw2, tiw3 int    // +1
func tiwInlined1() { // +2
	tiw1++ // +3
} // +4
func tiwInlined2() { // +5
	tiwInlined1() // +6
	tiw2++        // +7
} // +8
func tiwTest() { // +9
	tiwInlined1() // +10
	tiwInlined2() // +11
	tiw3++        // +12
} // +13
