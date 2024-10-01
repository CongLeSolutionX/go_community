// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

import (
	"cmd/compile/internal/base"
	"crypto/sha256"
	"fmt"
	"math/rand"
)

// layoutshuffle shuffles basic blocks for function f,
// it still guarantees the entry block is placed at the
// beginning.
func layoutshuffle(f *Func) {
	if !base.Flag.BbShuffle {
		return
	}
	hash := sha256.Sum256([]byte(fmt.Sprint(f.Blocks)))
	r := rand.New(rand.NewSource(int64(hash[0])))
	r.Shuffle(len(f.Blocks)-1, func(i, j int) {
		f.Blocks[i+1], f.Blocks[j+1] = f.Blocks[j+1], f.Blocks[i+1]
	})
	f.laidout = true
}
