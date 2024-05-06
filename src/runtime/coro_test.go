// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime_test

import (
	"runtime"
	"strings"
	"testing"
)

func TestCoroLockOSThread(t *testing.T) {
	for _, test := range []struct {
		name, failure string
	}{
		{
			name: "CoroLockOSThreadIterLock",
		},
		{
			name:    "CoroLockOSThreadIterLockYield",
			failure: "thread locked state does not match",
		},
		{
			name: "CoroLockOSThreadLock",
		},
		{
			name: "CoroLockOSThreadLockIterNested",
		},
		{
			name: "CoroLockOSThreadLockIterLock",
		},
		{
			name:    "CoroLockOSThreadLockIterLockYield",
			failure: "thread locked state does not match",
		},
		{
			name:    "CoroLockOSThreadLockIterYieldNewG",
			failure: "thread locked state does not match",
		},
		{
			name:    "CoroLockOSThreadLockAfterPull",
			failure: "thread locked state does not match",
		},
		{
			name: "CoroLockOSThreadStopLocked",
		},
		{
			name: "CoroLockOSThreadStopLockedIterNested",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			got := runTestProg(t, "testprog", test.name)
			if test.failure == "" {
				want := "OK\n"
				if got != want {
					t.Fatalf("expected %q, but got:\n%s", want, got)
				}
			} else {
				if !strings.Contains(got, test.failure) {
					t.Fatalf("expected %q in output, but got:\n%s", test.failure, got)
				}
			}
		})
	}
}

func TestCoroCgoCallback(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("coro cgo callback tests not supported on Windows")
	}
	for _, test := range []struct {
		name, failure string
	}{
		{
			name: "CoroCgoIterCallback",
		},
		{
			name:    "CoroCgoIterCallbackYield",
			failure: "thread locked state does not match",
		},
		{
			name: "CoroCgoCallback",
		},
		{
			name: "CoroCgoCallbackIterNested",
		},
		{
			name: "CoroCgoCallbackIterCallback",
		},
		{
			name:    "CoroCgoCallbackIterCallbackYield",
			failure: "thread locked state does not match",
		},
		{
			name:    "CoroCgoCallbackAfterPull",
			failure: "thread locked state does not match",
		},
		{
			name: "CoroCgoStopCallback",
		},
		{
			name: "CoroCgoStopCallbackIterNested",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			got := runTestProg(t, "testprogcgo", test.name)
			if test.failure == "" {
				want := "OK\n"
				if got != want {
					t.Fatalf("expected %q, but got:\n%s", want, got)
				}
			} else {
				if !strings.Contains(got, test.failure) {
					t.Fatalf("expected %q in output, but got:\n%s", test.failure, got)
				}
			}
		})
	}
}
