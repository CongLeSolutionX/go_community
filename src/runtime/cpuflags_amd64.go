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
	// Let's remove stepping and reserved fields
	processor := processorVersionInfo & 0x0FFF3FF0

	isIntelERMSGoodCPU := isIntel &&
		processor == 0x206A0 || // Sandy Bridge (Client)
		processor == 0x206D0 || // Sandy Bridge (Server)
		processor == 0x306A0 || // Ivy Bridge (Client)
		processor == 0x306E0 || // Ivy Bridge (Server)
		processor == 0x606A0 || // Ice Lake (Server) SP
		processor == 0x606C0 || // Ice Lake (Server) DE
		processor == 0x806F0 // Sapphire Rapids
	useERMS = isIntelERMSGoodCPU && cpu.X86.HasERMS
	useAVXmemmove = cpu.X86.HasAVX
}
