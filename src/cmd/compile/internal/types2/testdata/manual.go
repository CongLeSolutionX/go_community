// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file is tested when running "go test -run Manual"
// without source arguments. Use for one-off debugging.

package p

func f[T any](func(int), T) {}

func g[T any](T) {}

// TODO(gri) make this work for one-argument functions
func _() {
	f(g, 0)
}
