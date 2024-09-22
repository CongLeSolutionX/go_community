// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !linux || !amd64

package runtime

import (
	_ "unsafe"
)

//go:linkname getrandomVDSO
func getrandomVDSO(p []byte, flags uint32) (ret int, supported bool) {
	return -1, false
}

func vgrndPutState(state uintptr) {}
