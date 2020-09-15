// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package minitypes

import (
	"go/token"
	itypes "internal/types"
)

var Universe *Scope

func init() {
	Universe = NewScope(nil, token.NoPos, token.NoPos, "")
	itypes.BuildUniverse(wrapScope(Universe))
}
