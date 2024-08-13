// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package auth provides access to user-provided authentication credentials.
package auth

import (
	"cmd/go/internal/base"
	"cmd/go/internal/cfg"
	"log"
	"net/http"
	"strings"
	"sync"
)

var (
	credentialCache sync.Map // map[string]http.Header of prefix, header
	authOnce        sync.Once
)

// AddCredentials fills in the user's credentials for req, if any.
// The return value reports whether any matching credentials were found.
func AddCredentials(req *http.Request, prefix string) (ok bool) {
	log.SetPrefix("# GOAUTH: ")
	defer log.SetPrefix("")
	if req.URL.Scheme != "https" {
		return false
	}
	if cfg.GOAUTH == "off" { // Clear GOAUTH if requested.
		credentialCache.Clear()
		return false
	}
	authOnce.Do(func() {
		runGoAuth("")
	})
	if prefix != "" { // First fetch failed; re-invoke GOAUTH with prefix.
		runGoAuth(prefix)
	}
	if prefix == "" {
		prefix = req.Host
		if prefix == "" {
			prefix = req.URL.Hostname()
		}
	}
	return loadCredential(req, prefix)
}

// runGoAuth executes authentication commands specified by the GOAUTH
// environment variable handling 'netrc' and 'git' commands specially,
// and storing retrieved credentials for future module access.
// These credentials will be used for go-import resolution and the
// HTTPS module proxy protocol.
func runGoAuth(prefix string) {
	goAuthCmds := strings.Split(cfg.GOAUTH, ";")
	for _, cmdStr := range goAuthCmds {
		cmdStr = strings.TrimSpace(cmdStr)
		switch {
		case cmdStr == "netrc":
			netrcOnce.Do(readNetrc)
		case strings.HasPrefix(cmdStr, "git"):
			cmdParts := strings.Fields(cmdStr)
			if len(cmdParts) != 2 { // git $HOME
				base.Fatalf("provide the absolute path to the 'git' command's working directory as the first argument.")
			}
			runGitAuth(cmdParts[1], prefix)
		default:
			continue // TODO
		}
	}
}

// loadCredential retrieves cached credentials for the given url prefix and adds
// them to the request headers. A prefix could look like example.com or https://example.com/repo1
func loadCredential(req *http.Request, prefix string) (ok bool) {
	prefix, _ = strings.CutPrefix(prefix, "https://")
	headers, ok := credentialCache.Load(prefix)
	if !ok {
		return false
	}
	for key, values := range headers.(http.Header) {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	return true
}

// storeCredential caches or removes credentials (represented by HTTP headers)
// associated with given URL prefixes with the leading 'https://' removed.
func storeCredential(prefixes []string, header http.Header) {
	for _, prefix := range prefixes {
		prefix, _ = strings.CutPrefix(prefix, "https://")
		if len(header) == 0 {
			credentialCache.Delete(prefix)
		} else {
			credentialCache.Store(prefix, header)
		}
	}
}
