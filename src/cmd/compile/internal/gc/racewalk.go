// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gc

import (
	"cmd/compile/internal/types"
	"cmd/internal/src"
	"fmt"
)

// The instrument pass modifies the code tree for instrumentation.
//
// For flag_race it modifies the function as follows:
//
// 1. It inserts a call to racefuncenterfp at the beginning of each function.
// 2. It inserts a call to racefuncexit at the end of each function.
// 3. It inserts a call to raceread before each memory read.
// 4. It inserts a call to racewrite before each memory write.
//
// For flag_msan:
//
// 1. It inserts a call to msanread before each memory read.
// 2. It inserts a call to msanwrite before each memory write.
//
// The rewriting is not yet complete. Certain nodes are not rewritten
// but should be.

// TODO(dvyukov): do not instrument initialization as writes:
// a := make([]int, 10)

// Do not instrument the following packages at all,
// at best instrumentation would cause infinite recursion.
var omit_pkgs = []string{"runtime/internal/atomic", "runtime/internal/sys", "runtime", "runtime/race", "runtime/msan"}

// Only insert racefuncenterfp/racefuncexit into the following packages.
// Memory accesses in the packages are either uninteresting or will cause false positives.
var norace_inst_pkgs = []string{"sync", "sync/atomic"}

func ispkgin(pkgs []string) bool {
	if myimportpath != "" {
		for _, p := range pkgs {
			if myimportpath == p {
				return true
			}
		}
	}

	return false
}

func instrument(fn *Node) {
	if fn.Func.Pragma&Norace != 0 {
		return
	}

	if !flag_race || !ispkgin(norace_inst_pkgs) {
		fn.Func.SetInstrumentBody(true)
	}

	if flag_race {
		// nodpc is the PC of the caller as extracted by
		// getcallerpc. We use -widthptr(FP) for x86.
		// BUG: this will not work on arm.
		nodpc := *nodfp
		nodpc.Type = types.Types[TUINTPTR]
		nodpc.Xoffset = int64(-Widthptr)
		savedLineno := lineno
		lineno = src.NoXPos
		nd := mkcall("racefuncenter", nil, nil, &nodpc)

		fn.Func.Enter.Prepend(nd)
		nd = mkcall("racefuncexit", nil, nil)
		fn.Func.Exit.Append(nd)
		fn.Func.Dcl = append(fn.Func.Dcl, &nodpc)
		lineno = savedLineno
	}

	if Debug['W'] != 0 {
		s := fmt.Sprintf("after instrument %v", fn.Func.Nname.Sym)
		dumplist(s, fn.Nbody)
		s = fmt.Sprintf("enter %v", fn.Func.Nname.Sym)
		dumplist(s, fn.Func.Enter)
		s = fmt.Sprintf("exit %v", fn.Func.Nname.Sym)
		dumplist(s, fn.Func.Exit)
	}
}
