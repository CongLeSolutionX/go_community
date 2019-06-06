// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package web

import (
	"errors"
	"net/url"
	"path/filepath"
	"strings"
)

// TODO(golang.org/issue/32456): If accepted, move these functions into the
// net/url package.

var errNotAbsolute = errors.New("path is not absolute")

func urlToFilePath(u *url.URL) (string, error) {
	if u.Scheme != "file" {
		return "", errors.New("non-file URL")
	}

	checkAbs := func(path string) (string, error) {
		if !filepath.IsAbs(path) {
			return "", errNotAbsolute
		}
		return path, nil
	}

	if u.Path == "" {
		if u.Host != "" || u.Opaque == "" {
			return "", errors.New("file URL missing path")
		}
		return checkAbs(filepath.FromSlash(u.Opaque))
	}

	path, err := convertFileURLPath(u.Host, u.Path)
	if err != nil {
		return path, err
	}
	return checkAbs(path)
}

func urlFromFilePath(path string) (*url.URL, error) {
	if !filepath.IsAbs(path) {
		return nil, errNotAbsolute
	}

	u := &url.URL{Scheme: "file"}

	// If path has a Windows volume name, convert the volume to a host and prefix
	// per https://blogs.msdn.microsoft.com/ie/2006/12/06/file-uris-in-windows/.
	if vol := filepath.VolumeName(path); vol == "" {
		u.Path = filepath.ToSlash(path)
	} else if strings.HasPrefix(vol, `\\`) {
		path = filepath.ToSlash(path[2:])
		if i := strings.IndexByte(path, '/'); i < 0 {
			u.Host = path
			u.Path = "/"
		} else {
			u.Host = path[:i]
			u.Path = filepath.ToSlash(path[i:])
		}
	} else {
		u.Path = "/" + filepath.ToSlash(path)
	}

	return u, nil
}
