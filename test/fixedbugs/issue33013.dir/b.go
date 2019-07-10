// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package b

import (
	"a"
	"context"
)

type BI interface {
	Something(s int64) int64
	Another(pxp context.Context) int32
}

func BRS(sd *a.ServiceDesc, server BI, xyz int) *a.Service {
	return a.RS(sd, server, 7)
}
