// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file is tested when running "go test -run Manual"
// without source arguments. Use for one-off debugging.

package p

func f(func(int) string) {}

func g[P, Q any](P) Q { var q Q; return q }

func _() {
	f(g)
}
