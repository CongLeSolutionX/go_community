// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

// A genericType implements methods to access its type parameters.
type genericType interface {
	Type
	TypeParams() *TypeParamList
	SetTypeParams(tparams []*TypeParam)
}
