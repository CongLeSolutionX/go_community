// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fipstest

import (
	"crypto/internal/boring"
	"crypto/internal/fips"
	"crypto/internal/fips/pbkdf2"
	"crypto/sha256"
	"testing"
)

func TestPBKDF2ServiceIndicator(t *testing.T) {
	if boring.Enabled {
		t.Skip("in BoringCrypto mode PBKDF2 is not from the Go FIPS module")
	}

	goodSalt := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10}

	fips.ResetServiceIndicator()
	_, err := pbkdf2.Key(sha256.New, "password", goodSalt, 1, 32)
	if err != nil {
		t.Error(err)
	}
	if !fips.ServiceIndicator() {
		t.Error("FIPS service indicator should be set")
	}

	// Salt too short
	fips.ResetServiceIndicator()
	_, err = pbkdf2.Key(sha256.New, "password", goodSalt[:8], 1, 32)
	if err != nil {
		t.Error(err)
	}
	if fips.ServiceIndicator() {
		t.Error("FIPS service indicator should not be set")
	}

	// Key length too short
	fips.ResetServiceIndicator()
	_, err = pbkdf2.Key(sha256.New, "password", goodSalt, 1, 10)
	if err != nil {
		t.Error(err)
	}
	if fips.ServiceIndicator() {
		t.Error("FIPS service indicator should not be set")
	}
}
