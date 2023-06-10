// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package httputil

import (
	"internal/fakenet"
	"internal/testenv"
)

func init() {
	fakenet.SetEnabled(testenv.HasFakeNetwork())
}
