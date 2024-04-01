// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"context"
	"net/http/httptrace"
	"net/url"
)

// A ClientConnPool manages a pool of HTTP client connections.
//
// The pool is responsible for determining which connection to use for new requests
// and managing the lifetime of idle connections.
//
// If a ClientConnPool implementation includes a CloseIdleConnections method,
// it will be called by [Client.CloseIdleConnections].
type ClientConnPool interface {
	// Get requests a connection from the pool.
	//
	// When Get returns successfully, one request is consumed on the returned connection.
	Get(context.Context, ClientConnRequest) (_ *ClientConn, _ error)
}

// ClientConnRequest is a request to a ClientConnPool for a request.
type ClientConnRequest struct {
	Transport *Transport
	Request   *Request
	Trace     *httptrace.ClientTrace
	Options   TransportDialOptions
}

// TransportDialOptions describe the requirements for a connection.
type TransportDialOptions struct {
	ProxyURL *url.URL // nil for no proxy, else full proxy URL
	Scheme   string   // "http" or "https"
	Address  string

	// TODO: Make this more extensible (HTTP/3, unencrypted HTTP/2, etc.)
	OnlyH1 bool // whether to disable HTTP/2 and force HTTP/1
}

// Key returns a comparable represenation of TransportDialOptions,
// suitable for use as a map key.
func (opts *TransportDialOptions) Key() TransportDialKey {
	return TransportDialKey{}
}

// TransportDialKey is a comparable representation of a TransportDialOptions,
// suitable for use as a map key.
type TransportDialKey struct{}

// A ClientConn is an HTTP client connection.
type ClientConn persistConn

// RoundTrip sends a request to the connection.
func (cc *ClientConn) RoundTrip(req *Request) (*Response, error) {
	return nil, nil
}

// SetEventHandler sets the event handler for the connection.
// The connection will call f to report changes in the state of the connection.
// See [ClientConnEventType] for event types.
//
// At most one call to f will be made at a time.
// If the event handler blocks, it may block requests on the connection.
//
// If the request limit for the connection has changed since it was created,
// a ClientConnRequestLimitIncreased event will be immediately generated.
func (cc *ClientConn) SetEventHandler(f func(ClientConnEvent)) {}

type ClientConnEventType int

const (
	// ClientConnRequestLimitIncreased reports that a ClientConn
	// can handle additional requests.
	// The [ClientConnEvent.Value] field contains the number of additional requests.
	// Note that this is an increase in the number of requests, not an absolute value:
	// A Value of 1 indicates that the ClientConn can handle one more request than before.
	ClientConnRequestLimitIncreased = ClientConnEventType(iota + 1)

	// ClientConnClosed reports that a ClientConn cannot handle further requests.
	// Requests in flight on the connection may or may not complete successfully.
	ClientConnClosed
)

type ClientConnEvent struct {
	Type     ClientConnEventType
	Conn     *ClientConn
	IntValue int   // set for ClientConnRequestLimitIncreased
	Err      error // set for ClientConnClosed
}

// NewClientConn creates a new connection with the given options.
//
// The initialRequestLimit is the number of simultaneous requests that may be made
// on the new connection.
func (t *Transport) NewClientConn(ctx context.Context, opts TransportDialOptions) (_ *ClientConn, initialRequestLimit int, _ error) {
	return nil, 0, nil
}
