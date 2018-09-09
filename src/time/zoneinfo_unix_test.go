// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time_test

import (
	"testing"
	"os"
	"time"
)

func TestEnvTZUsage(t *testing.T) {
	const env = "TZ"
	tz, ok := os.LookupEnv(env)
	if !ok {
		defer os.Unsetenv(env)
	} else {
		defer os.Setenv(env, tz)
	}
	defer time.ForceUSPacificForTesting()

	cases := []struct {
		nilFlag bool
		tz string
		local string
	}{
		// no $TZ means use the system default /etc/localtime.
		{true, "", "Local"},
		// $TZ="" means use UTC.
		{false, "", "UTC"},
		{false, "Asia/Shanghai", "Asia/Shanghai"},
		{false, ":Asia/Shanghai", "Asia/Shanghai"},
	}

	for _, c := range cases {
		time.ResetLocalOnceForTest()
		if c.nilFlag {
			os.Unsetenv(env)
		} else {
			os.Setenv(env, c.tz)
		}
		if time.Local.String() != c.local {
			t.Errorf("invalid Local location name for %q: got %q want %q", c.tz, time.Local, c.local)
		}
	}

	time.ResetLocalOnceForTest()
	// The file may not exists on Solaris 2 and IRIX 6.
	path := "/usr/share/zoneinfo/Asia/Shanghai"
	os.Setenv(env, path)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if time.Local.String() != "UTC" {
			t.Errorf(`invalid path should fallback to UTC: got %q want "UTC"`, time.Local)
		}
		return
	}
	if time.Local.String() != "Local" {
		t.Errorf(`custom path should lead to Local: got %q want "Local"`, time.Local)
	}

	timeInUTC := time.Date(2009, 1, 1, 12, 0, 0, 0, time.UTC)
	sameTimeInShanghai := time.Date(2009, 1, 1, 20, 0, 0, 0, time.Local)
	if !timeInUTC.Equal(sameTimeInShanghai) {
		t.Errorf("invalid timezone: got %q want %q", timeInUTC, sameTimeInShanghai)
	}

	time.ResetLocalOnceForTest()
	os.Setenv(env, path[:len(path) - 1])
	if time.Local.String() != "UTC" {
		t.Errorf(`invalid path should fallback to UTC: got %q want "UTC"`, time.Local)
	}
}
