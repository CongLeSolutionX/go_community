// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !amd64,!amd64p32,!s390x

package crc32

var updateCastagnoli = updateCastagnoliGeneric
var updateCastagnoliString = updateCastagnoliGenericString
var updateIEEE = updateIEEEGeneric
var updateIEEEString = updateIEEEGenericString
