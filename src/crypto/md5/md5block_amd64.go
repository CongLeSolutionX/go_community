// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package md5

import "internal/cpu"

var useAVX512 = cpu.X86.HasAVX512F && cpu.X86.HasAVX512VL
