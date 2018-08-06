// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package x509

import (
	"os"
	"os/exec"
	"runtime"
	"testing"
	"time"
)

func TestSystemRoots(t *testing.T) {
	switch runtime.GOARCH {
	case "arm", "arm64":
		t.Skipf("skipping on %s/%s, no system root", runtime.GOOS, runtime.GOARCH)
	}

	t0 := time.Now()
	sysRoots := systemRootsPool() // actual system roots
	sysRootsDuration := time.Since(t0)

	t1 := time.Now()
	execRoots, err := execSecurityRoots() // non-cgo roots
	execSysRootsDuration := time.Since(t1)

	if err != nil {
		t.Fatalf("failed to read system roots: %v", err)
	}

	t.Logf("    cgo sys roots: %v", sysRootsDuration)
	t.Logf("non-cgo sys roots: %v", execSysRootsDuration)

	// On Mavericks, there are 212 bundled certs, at least there was at
	// one point in time on one machine. (Maybe it was a corp laptop
	// with extra certs?) Other OS X users report 135, 142, 145...
	// Let's try requiring at least 100, since this is just a sanity
	// check.
	if want, have := 100, len(sysRoots.certs); have < want {
		t.Errorf("want at least %d system roots, have %d", want, have)
	}

	// Check that the two cert pools are the same.
	sysPool := make(map[string]*Certificate, len(sysRoots.certs))
	for _, c := range sysRoots.certs {
		sysPool[string(c.Raw)] = c
	}
	for _, c := range execRoots.certs {
		if _, ok := sysPool[string(c.Raw)]; ok {
			delete(sysPool, string(c.Raw))
		} else {
			// verify-cert lets in certificates that are not trusted roots, but are
			// signed by trusted roots. This should not be a problem, so confirm that's
			// the case and skip them.
			if _, err := c.Verify(VerifyOptions{
				Roots:         sysRoots,
				Intermediates: execRoots, // the intermediates for EAP certs are stored in the keychain
				KeyUsages:     []ExtKeyUsage{ExtKeyUsageAny},
			}); err != nil {
				t.Errorf("certificate only present in non-cgo pool: %v (verify error: %v)", c.Subject, err)
			} else {
				t.Logf("signed certificate only present in non-cgo pool (acceptable): %v", c.Subject)
			}
		}
	}
	for _, c := range sysPool {
		t.Errorf("certificate only present in cgo pool: %v", c.Subject)
	}

	if t.Failed() && debugDarwinRoots {
		cmd := exec.Command("security", "dump-trust-settings")
		cmd.Stdout = os.Stdout
		cmd.Run()
		cmd = exec.Command("security", "dump-trust-settings", "-d")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}
