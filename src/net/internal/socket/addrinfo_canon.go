// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build cgo,!netgo
// +build android netbsd openbsd

package socket

/*
#include <netdb.h>
*/
import "C"

const addrinfoFlags = C.AI_CANONNAME
