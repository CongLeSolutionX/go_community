// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package web

import (
	"net/url"
	"testing"
)

func TestURLToFilePath(t *testing.T) {
	for _, tc := range urlTests {
		if tc.url == "" {
			continue
		}
		tc := tc

		t.Run(tc.url, func(t *testing.T) {
			u, err := url.Parse(tc.url)
			if err != nil {
				t.Fatalf("url.Parse(%q): %v", tc.url, err)
			}

			path, err := urlToFilePath(u)
			if err != nil {
				if err.Error() == tc.wantErr {
					return
				}
				if tc.wantErr == "" {
					t.Fatalf("urlToFilePath(%v): %v; want <nil>", u, err)
				} else {
					t.Fatalf("urlToFilePath(%v): %v; want %s", u, err, tc.wantErr)
				}
			}

			if tc.wantErr != "" {
				t.Fatalf("urlToFilePath(%v) = <nil>; want error: %s", u, tc.wantErr)
			}

			if path != tc.filePath {
				t.Errorf("urlToFilePath(%v) = %q; want %q", u, path, tc.filePath)
			}
		})
	}
}

func TestURLFromFilePath(t *testing.T) {
	for _, tc := range urlTests {
		if tc.filePath == "" {
			continue
		}
		tc := tc

		t.Run(tc.filePath, func(t *testing.T) {
			u, err := urlFromFilePath(tc.filePath)
			if err != nil {
				if err.Error() == tc.wantErr {
					return
				}
				if tc.wantErr == "" {
					t.Fatalf("urlFromFilePath(%v): %v; want <nil>", tc.filePath, err)
				} else {
					t.Fatalf("urlFromFilePath(%v): %v; want %s", tc.filePath, err, tc.wantErr)
				}
			}

			if tc.wantErr != "" {
				t.Fatalf("urlFromFilePath(%v) = <nil>; want error: %s", tc.filePath, tc.wantErr)
			}

			wantURL := tc.url
			if tc.canonicalURL != "" {
				wantURL = tc.canonicalURL
			}
			if u.String() != wantURL {
				t.Errorf("urlFromFilePath(%v) = %v; want %s", tc.filePath, u, wantURL)
			}
		})
	}
}
