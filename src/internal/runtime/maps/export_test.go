// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package maps

import (
	"internal/abi"
)

const DebugLog = debugLog

var AlignUpPow2 = alignUpPow2

func (t *table) Type() *abi.SwissMapType {
	return t.typ
}
