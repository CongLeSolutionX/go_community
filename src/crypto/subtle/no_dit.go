// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !arm64 || purego

package subtle

var ditSupported = false

func enableDIT() bool  { return false }
func ditEnabled() bool { return false }
func disableDIT()      {}
