// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	"internal/cpu"
)

var (
	useAVXmemmove bool
	useERMS       bool
)

func init() {
	// Here we assume that on modern CPUs with both FSRM and ERMS features,
	// copying data blocks of 2KB or larger using the REP MOVSB instruction
	// will be more efficient to avoid having to keep up with CPU generations.
	// Therefore, we may retain a BlockList mechanism to ensure that microarchitectures
	// that do not fit this case may appear in the future.
	isERMSNiceCPU := isIntel
	useERMS = isERMSNiceCPU && cpu.X86.HasERMS && cpu.X86.HasFSRM
	useAVXmemmove = cpu.X86.HasAVX
}
