// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

TEXT svetest(SB),$0
    ZBFDOT Z0.H, Z0.H, Z0.S                          // 00806064
    ZBFDOT Z11.H, Z12.H, Z10.S                       // 6a816c64
    ZBFDOT Z31.H, Z31.H, Z31.S                       // ff837f64
    
    RET
