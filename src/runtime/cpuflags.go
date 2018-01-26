// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

// Information about what x86 CPU features are available.
// Set on startup in asm_{386,amd64,amd64p32}.s.
// Packages outside the runtime should not use these
// as they are not an external api.
var (
	x86_processorVersionInfo uint32
	x86_isIntel              bool
	x86_lfenceBeforeRdtsc    bool
	x86_hasAES               bool
	x86_hasAVX               bool
	x86_hasAVX2              bool
	x86_hasBMI1              bool
	x86_hasBMI2              bool
	x86_hasERMS              bool
	x86_hasOSXSAVE           bool
	x86_hasPOPCNT            bool
	x86_hasSSE2              bool
	x86_hasSSE41             bool
	x86_hasSSE42             bool
	x86_hasSSSE3             bool
)
