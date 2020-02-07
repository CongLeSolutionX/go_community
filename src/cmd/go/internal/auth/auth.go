// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package auth provides access to user-provided authentication credentials.
package auth

import (
	"net"
	"net/http"
)

// AddCredentials fills in the user's credentials for req, if any.
// The return value reports whether any matching credentials were found.
func AddCredentials(req *http.Request) (added bool) {
	// Host from req.URL.Host can be of the form host:post, but netrc spec
	// defines machine to be "machine name", so check for :port and remove
	// before matching to netrc machine names.
	// https://www.gnu.org/software/inetutils/manual/html_node/The-_002enetrc-file.html
	host, _, err := net.SplitHostPort(req.URL.Host)
	if err != nil {
		// Missing port (or invalid), leave as-is.
		host = req.URL.Host
	}

	// TODO(golang.org/issue/26232): Support arbitrary user-provided credentials.
	netrcOnce.Do(readNetrc)
	for _, l := range netrc {
		if l.machine == host {
			req.SetBasicAuth(l.login, l.password)
			return true
		}
	}

	return false
}
