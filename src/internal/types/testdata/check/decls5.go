// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

// declarations of main
const _, main /* ERR cannot declare main */ , _ = 0, 1, 2
type main /* ERR cannot declare main */ struct{}
var _, main /* ERR cannot declare main */ int
