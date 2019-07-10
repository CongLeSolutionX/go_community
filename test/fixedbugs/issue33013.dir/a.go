// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package a

import "context"

type Service uint64
type ServiceDesc struct {
	X int
	uc
}

type uc interface {
	f() context.Context
}

var q int

func RS(svcd *ServiceDesc, server interface{}, qq uint8) *Service {
	defer func() { q += int(qq) }()
	return nil
}
