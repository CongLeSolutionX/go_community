// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package coverage

// NB: the "counters" object we generate for each instrumented function
// can be thought of as a struct of the following form:
//
// struct {
//     numCtrs uint32
//     pkgid uint32
//     funcid uint32
//     counterArray [numBlocks]uint32
// }
//
// where "numCtrs" is the number of blocks / coverable units within the
// function, "pkgid" is the unique index assigned to this package by
// the runtime, "funcid" is the index of this function within its containing
// packge, and "counterArray" stores the actual counters.
//
// The counter variable itself is created not as a struct but as a flat
// array of uint32's; we then use the offsets below to index into it.

const NumCtrsOffset = 0
const PkgIdOffset = 1
const FuncIdOffset = 2
const FirstCtrOffset = 3
