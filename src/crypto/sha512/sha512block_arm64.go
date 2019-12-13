// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sha512

func block(dig *digest, p []byte) {
	blockGeneric(dig, p)
}
