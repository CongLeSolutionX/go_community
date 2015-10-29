// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package httptrace contains hooks for tracing the net/http Transport.
//
// This package is logically part of net/http but is separated to reduce
// clutter in the main package.
package httptrace

import "net"

// Handle is an opaque value returned by Start and passed to other hook
// phases through a request. It is where traces can put state.
type Handle interface{}

// Request is an *http.Request. It exists to break an otherwise-circular
// dependency between net/http and net/http/httptrace.
type Request interface{}

// Response is an *http.Response. It exists to break an otherwise-circular
// dependency between net/http and net/http/httptrace.
type Response interface{}

// ClientTrace is the set of hooks to be run at various phases
// of an HTTP client requests.
//
// All hooks are optional.
type ClientTrace struct {
	// Start is called at the beginning of an HTTP request, before
	// any work happens.
	Start func(Request) Handle

	// GotConn is called when a connection is acquired, just
	// before sending the request. Implementations should not
	// affect c.
	GotConn func(h Handle, c net.Conn, reused bool)

	// ExitRoundTrip is called when exiting RoundTrip.
	ExitRoundTrip func(Handle, Response, error)

	/*
		TODO / notes from meeting (all can be added later):

		* DNS start?
		* socket late binding race won/lost.

		Connecting // (or StartConnect)
		SentDNS  // implies we needed a conn
		GotDNSResponse
		Connected
		TLSConnected
		UsingCachedConn

		DialReturned func(Handle, net.Conn, error)
		DialTLSReturned func(Handle, net.Conn, error)

		WritingRequest(handle)
		WroteRequestHeaders(Handle)
		WroteRequestBody(Handle)

		GotFirstByte(Handle)
		GotResponseHeader(Handler, http.Header)

		end of response body?
	*/
}
