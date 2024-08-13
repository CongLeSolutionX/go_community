// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package auth

import (
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

var testUserAuth = `
https://example.com

Authorization: Basic YWxhZGRpbjpvcGVuc2VzYW1l
Authorization: Basic jpvcGVuc2VzYW1lYWxhZGRpb
Data: Test123

`

func TestParseUserAuth(t *testing.T) {
	// Build the expected header
	wantHeader := http.Header{}
	wantHeader.Add("Authorization", "Basic YWxhZGRpbjpvcGVuc2VzYW1l")
	wantHeader.Add("Authorization", "Basic jpvcGVuc2VzYW1lYWxhZGRpb")
	wantHeader.Add("Data", "Test123")

	// Process the simulated 'GOAUTH' output
	data := io.NopCloser(strings.NewReader(testUserAuth))
	parseAuthOutput(data)
	gotReq := &http.Request{Header: make(http.Header)}
	ok := loadCredential(gotReq, "https://example.com")
	if !ok || !reflect.DeepEqual(gotReq.Header, wantHeader) {
		t.Errorf("parseUserAuth:\nhave %q\nwant %q", gotReq.Header, wantHeader)
	}

	// Ensure that we can resolve the same credential without https://
	gotReq = &http.Request{Header: make(http.Header)}
	ok = loadCredential(gotReq, "example.com")
	if !ok || !reflect.DeepEqual(gotReq.Header, wantHeader) {
		t.Errorf("parseUserAuth:\nhave %q\nwant %q", gotReq.Header, wantHeader)
	}
}

func TestParseUserAuthClear(t *testing.T) {
	var testUserAuthClear = `
https://example.com


`
	// Process the simulated 'GOAUTH' output
	data := io.NopCloser(strings.NewReader(testUserAuth))
	parseAuthOutput(data)
	req := &http.Request{Header: make(http.Header)}
	ok := loadCredential(req, "https://example.com")
	if !ok {
		t.Errorf("parseUserAuth: failed to store credentials")
	}
	data = io.NopCloser(strings.NewReader(testUserAuthClear))
	parseAuthOutput(data)
	ok = loadCredential(req, "https://example.com")
	if ok {
		t.Errorf("parseUserAuth: empty headers should clear credentials")
	}
}
