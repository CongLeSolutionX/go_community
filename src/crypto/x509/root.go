// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package x509

// To update the embedded iOS root store, update the -version
// argument to the latest security_certificates version from
// https://opensource.apple.com/source/security_certificates/
// and run "go generate". See https://golang.org/issue/38843.
//
//go:generate go run root_ios_gen.go -version 55188.120.1.0.1

import (
	"internal/godebug"
	"sync"
)

var (
	once           sync.Once
	systemRoots    *CertPool
	systemRootsErr error
)

func systemRootsPool() *CertPool {
	once.Do(initSystemRoots)
	return systemRoots
}

func initSystemRoots() {
	systemRoots, systemRootsErr = loadSystemRoots()
	if systemRootsErr != nil {
		systemRoots = nil
	}
}

var (
	fallbackMu   sync.Mutex
	fallbacksSet bool
)

// SetFallbackRoots sets the roots to use during certificate verification, if no
// custom roots are specified and a platform verifier or a system certificate
// pool is not available (for instance in a container which does not have a root
// certificate bundle).
//
// SetFallbackRoots may only be called once, if called multiple times, it will
// panic.
//
// The fallback behavior can be forced, even when there is a system certificate
// pool, by setting GODEBUG=forcerootfallback=1. Note this behavior cannot be
// forced on platforms where the platform certificate verifier is used (Windows
// and macOS) as we have no insight into the contents of the platforms root pool.
func SetFallbackRoots(roots *CertPool) {
	fallbackMu.Lock()
	defer fallbackMu.Unlock()

	if fallbacksSet {
		panic("SetFallbackRoots has already been called")
	}
	fallbacksSet = true

	_ = systemRootsPool()

	forceFallback := godebug.Get("forcerootfallback") == "1"

	if systemRoots != nil && ((systemRoots.len() > 0 && !forceFallback) || systemRoots.systemPool) {
		return
	}

	systemRoots, systemRootsErr = roots, nil
}
