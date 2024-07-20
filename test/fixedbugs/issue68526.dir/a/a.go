// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build goexperiment.aliastypeparams

package a

type G[S any] struct{ F S }

type A[T any] = G[T]
