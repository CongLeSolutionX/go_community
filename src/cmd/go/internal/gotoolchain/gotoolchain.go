// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gotoolchain

const ModulePath = "swtch.com/tmp/gotoolchain"

func Module(govers string) (path, vers string, ok bool) {
	return ModulePath, "v0.0.2-" + govers, true
}
