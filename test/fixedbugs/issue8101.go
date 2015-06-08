// compile

// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 8101: cmd/5g: ICE regfree: reg R0 not allocated

package p

var c0 []chan int
var s0 []int
var s1 []**int
var i = 0
var boom = cap(c0[s0[len(c0[**(s1)[s0[s0[**s1[s0[s0[s0[i]]+1]-1]]]]])]+1])
