// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ppc64

package aes

// For compatibility with cipher_asm.go
type aesCipherGCM struct {
	aesCipherAsm
}
