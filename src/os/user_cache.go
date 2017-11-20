// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os

import "runtime"

// UserCacheDir returns the default directory to use for user-specific
// cached data.
//
// On Unix systems, it returns $XDG_CACHE_HOME as specified by
// https://standards.freedesktop.org/basedir-spec/basedir-spec-latest.html if
// non-empty, else $HOME/.cache.
// On Darwin, it returns $HOME/Library/Caches.
// On Windows, it returns %LocalAppData%.
// On Plan 9, it returns $home/lib/cache.
//
// If the location cannot be determined (for example, $HOME is not defined),
// then it will return an empty string.
func UserCacheDir() string {
	var dir string

	switch runtime.GOOS {
	case "windows":
		dir = Getenv("LocalAppData")

	case "darwin":
		dir = Getenv("HOME")
		if dir == "" {
			return ""
		}
		dir += "/Library/Caches"

	case "plan9":
		dir = Getenv("home")
		if dir == "" {
			return ""
		}
		dir += "/lib/cache"

	default: // Unix
		dir = Getenv("XDG_CACHE_HOME")
		if dir == "" {
			dir = Getenv("HOME")
			if dir == "" {
				return ""
			}
			dir += "/.cache"
		}
	}

	return dir
}
