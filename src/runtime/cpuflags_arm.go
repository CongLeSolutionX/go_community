// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	"internal/cpu"
)

// ./run.bash doesn't assemble vlob_arm.s properly with
// internal/cpu·ARM+const_ARM_HasIDIVA(SB) as an operand to MOVBU.
var hardDiv = cpu.ARM.HasIDIVA
