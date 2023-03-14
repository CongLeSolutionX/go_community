// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// FIXED_FRAME defines the size of the fixed part of a stack frame. A stack
// frame looks like this:
//
// +---------------------+
// | local variable area |
// +---------------------+
// | argument area       |
// +---------------------+ <- R1+FIXED_FRAME
// | fixed area          |
// +---------------------+ <- R1
//
// So a function that sets up a stack frame at all uses as least FIXED_FRAME
// bytes of stack. This mostly affects assembly that calls other functions
// with arguments (the arguments should be stored at FIXED_FRAME+0(R1),
// FIXED_FRAME+8(R1) etc) and some other low-level places.
//
// The reason for using a constant is to make supporting PIC easier (although
// we only support PIC on ppc64le which has a minimum 32 bytes of stack frame,
// and currently always use that much, PIC on ppc64 would need to use 48).

#define FIXED_FRAME 32

// ELFv1 is used by linux/ppc64.
#ifdef GOOS_linux
#ifdef GOARCH_ppc64
#define GO_PPC64X_ELFV1
#endif
#endif

// ELFv2 is used by linux/ppc64le.
#ifdef GOOS_linux
#ifdef GOARCH_ppc64le
#define GO_PPC64X_ELFV2
#endif
#endif

// ELFv2 is used by openbsd/ppc64.
#ifdef GOOS_openbsd
#ifdef GOARCH_ppc64
#define GO_PPC64X_ELFV2
#endif
#endif

// XCOFF is used by aix/ppc64.
#ifdef GOOS_aix
#ifdef GOARCH_ppc64
#define GO_PPC64X_XCOFF
#endif
#endif

// Both ELFv1 and XCOFF use function descriptors.
#ifdef GO_PPC64X_ELFV1
#define GO_PPC64X_HAS_FUNCDESC
#endif
#ifdef GO_PPC64X_XCOFF
#define GO_PPC64X_HAS_FUNCDESC
#endif
