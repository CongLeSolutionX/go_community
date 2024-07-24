// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Export for testing.

package secret

func GetStack() (uintptr, uintptr) {
	return getStack()
}
