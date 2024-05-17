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
	isERMSNiceCPU := isIntel
	useERMS = isERMSNiceCPU && cpu.X86.HasERMS
	useAVXmemmove = cpu.X86.HasAVX
}
