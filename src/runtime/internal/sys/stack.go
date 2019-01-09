// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sys

// AIX requires a larger stack for syscalls.
const StackGuardMultiplier = StackGuardMultiplierHost*(1-GoosAix) + 2*GoosAix
