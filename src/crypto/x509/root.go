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
	systemRootsMu  sync.RWMutex
	systemRoots    *CertPool
	systemRootsErr error
)

func systemRootsPool() *CertPool {
	once.Do(initSystemRoots)
	systemRootsMu.RLock()
	defer systemRootsMu.RUnlock()
	return systemRoots
}

func initSystemRoots() {
	systemRootsMu.Lock()
	defer systemRootsMu.Unlock()
	systemRoots, systemRootsErr = loadSystemRoots()
	if systemRootsErr != nil {
		systemRoots = nil
	}
}

var (
	fallbackMu   sync.Mutex
	fallbacksSet bool

	forceFallback = godebug.New("x509usefallbackroots")
)

// SetFallbackRoots sets the roots to use during certificate verification, if no
// custom roots are specified and a platform verifier or a system certificate
// pool is not available (for instance in a container which does not have a root
// certificate bundle).
//
// SetFallbackRoots may only be called once, if called multiple times it will
// panic.
//
// The fallback behavior can be forced on all platforms, even when there is a
// system certificate pool, by setting GODEBUG=x509usefallbackroots=1 (note that
// on Windows and macOS this will disable usage of the platform verification
// APIs and cause the pure Go verifier to be used.) Setting
// x509usefallbackroots=1 without calling SetFallbackRoots has no effect.
func SetFallbackRoots(roots *CertPool) {
	fallbackMu.Lock()
	defer fallbackMu.Unlock()

	if fallbacksSet {
		panic("SetFallbackRoots has already been called")
	}
	fallbacksSet = true

	// populate the systemRoots if systemRootsPool has not been called already
	_ = systemRootsPool()

	systemRootsMu.Lock()
	defer systemRootsMu.Unlock()

	if systemRoots != nil && ((systemRoots.len() > 0 || systemRoots.systemPool) && forceFallback.Value() != "1") {
		return
	}
	systemRoots, systemRootsErr = roots, nil
}
