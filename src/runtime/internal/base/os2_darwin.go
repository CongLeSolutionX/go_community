// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package base

const (
	NSIG        = 32
	SI_USER     = 0 /* empirically true, but not what headers say */
	SIG_BLOCK   = 1
	SIG_UNBLOCK = 2
	SIG_SETMASK = 3
	SS_DISABLE  = 4
)
