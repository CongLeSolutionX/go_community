// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime_test

import (
	"runtime"
	"syscall"
	"testing"
)

func TestFixedGOROOT(t *testing.T) {
	envs := runtime.Envs()
	oldenvs := append([]string{}, envs...)
	defer runtime.SetEnvs(oldenvs)

	// attempt to reuse existing envs backing array.
	runtime.SetEnvs(append(envs[:0], "GOROOT=/"))

	if got := runtime.GOROOT(); got != "/" {
		t.Errorf(`runtime.GOROOT()=%q, want "/"`, got)
	}
	if err := syscall.Setenv("GOROOT", "/os"); err != nil {
		t.Fatal(err)
	}
	if got := runtime.GOROOT(); got != "/" {
		t.Errorf(`runtime.GOROOT()=%q, want "/"`, got)
	}
	if err := syscall.Unsetenv("GOROOT"); err != nil {
		t.Fatal(err)
	}

}
