// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Implementation of Client

package httptest

import (
	"net/http"
)

// A Client is an HTTP client for testing that allows the called to close
// its own idle connections.
type Client struct {
	*http.Client
}

// NewClient returns an HTTP client. The caller should call
// CloseIdleConnections when finished for the transport to close any idle
// connections.
func NewClient() *Client {
	return &Client{
		Client: &http.Client{Transport: &http.Transport{}},
	}
}

// CloseIdleTransport closes idle connections in the transport.
func (c *Client) CloseIdleTransport() {
	if t, ok := c.Transport.(closeIdleTransport); ok {
		t.CloseIdleConnections()
	}
}
