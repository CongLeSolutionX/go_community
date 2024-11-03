// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build (!amd64 && !s390x && !ppc64 && !ppc64le && !arm64) || purego

package gcm

func initGCM(g *GCM) {
	initGCMGeneric(g)
}

func seal(g *GCM, dst, nonce, plaintext, data []byte) []byte {
	return sealGeneric(g, dst, nonce, plaintext, data)
}

func open(g *GCM, dst, nonce, ciphertext, data []byte) ([]byte, error) {
	return openGeneric(g, dst, nonce, ciphertext, data)
}
