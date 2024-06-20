// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	"internal/cpu"
)

var memmoveBits uint8

func init() {
	// Here we assume that on modern CPUs with both FSRM and ERMS features,
	// copying data blocks of 2KB or larger using the REP MOVSB instruction
	// will be more efficient to avoid having to keep up with CPU generations.
	// Therefore, we may retain a BlockList mechanism to ensure that microarchitectures
	// that do not fit this case may appear in the future.
	isERMSNiceCPU := isIntel
	useREPMOV := isERMSNiceCPU && cpu.X86.HasERMS && cpu.X86.HasFSRM
	if cpu.X86.HasAVX {
		memmoveBits |= 0b1
	}
	if useREPMOV {
		memmoveBits |= 0b10
	}
}
