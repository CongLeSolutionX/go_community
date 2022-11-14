// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package httpguts

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/textproto"
	"net/url"
	"time"
)

type Header = textproto.MIMEHeader

type ClientRequest struct {
	Context         context.Context
	Method          string
	URL             *url.URL
	Header          Header
	Trailer         Header
	ResponseTrailer *Header
	Body            io.ReadCloser // if NoBody, Body must be nil
	GetBody         func() (io.ReadCloser, error)
	ContentLength   int64
	Close           bool
	Cancel          <-chan struct{}
	Host            string

	DisableCompression    bool
	ExpectContinueTimeout time.Duration
	ResponseHeaderTimeout time.Duration
	Dial                  bool
}

type ClientResponse struct {
	StatusCode    int
	Proto         string
	ProtoMajor    int
	ProtoMinor    int
	Header        Header
	Body          io.ReadCloser
	ContentLength int64
	Uncompressed  bool
	TLS           *tls.ConnectionState
}

type RoundTripper interface {
	RoundTrip(*ClientRequest) (*ClientResponse, error)
	UpgradeConn(string, *tls.Conn, UpgradeConnOpts) error
	CloseIdleConnections()
}

type UpgradeConnOpts struct {
	IdleConnTimeout time.Duration
}

var ErrNoCachedConn = errors.New("no cached connection was available")

type ServerConfig struct {
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	IdleTimeout    time.Duration
	MaxHeaderBytes int
	ConnState      func(net.Conn, ConnState)
}

type ServerRequest struct {
	Context       context.Context
	Proto         string
	ProtoMajor    int
	ProtoMinor    int
	Method        string
	Host          string
	URL           *url.URL
	Header        Header
	Trailer       Header
	Body          io.ReadCloser
	ContentLength int64
	RemoteAddr    string
	RequestURI    string
	TLS           *tls.ConnectionState
}

type Handler interface {
	ServeHTTP(ResponseWriter, *ServerRequest)
}

type ResponseWriter interface {
	Header() Header
	Write([]byte) (int, error)
	WriteHeader(int)
	Flush()
	FlushError() error
	SetReadDeadline(time.Time) error
	SetWriteDeadline(time.Time) error
	CloseNotify() <-chan bool
}

var (
	// ServerContextKey is a context key. It can be used in HTTP
	// handlers with Context.Value to access the server that
	// started the handler. The associated value will be of
	// type *Server.
	ServerContextKey = &contextKey{"http-server"}

	// LocalAddrContextKey is a context key. It can be used in
	// HTTP handlers with Context.Value to access the local
	// address the connection arrived on.
	// The associated value will be of type net.Addr.
	LocalAddrContextKey = &contextKey{"local-addr"}
)

// contextKey is a value for use with context.WithValue. It's used as
// a pointer so it fits in an interface{} without allocation.
type contextKey struct {
	name string
}

func (k *contextKey) String() string { return "net/http context value " + k.name }

const DefaultMaxHeaderBytes = 1 << 20 // 1 MB

// A ConnState represents the state of a client connection to a server.
// It's used by the optional Server.ConnState hook.
type ConnState int

const (
	// StateNew represents a new connection that is expected to
	// send a request immediately. Connections begin at this
	// state and then transition to either StateActive or
	// StateClosed.
	StateNew ConnState = iota

	// StateActive represents a connection that has read 1 or more
	// bytes of a request. The Server.ConnState hook for
	// StateActive fires before the request has entered a handler
	// and doesn't fire again until the request has been
	// handled. After the request is handled, the state
	// transitions to StateClosed, StateHijacked, or StateIdle.
	// For HTTP/2, StateActive fires on the transition from zero
	// to one active request, and only transitions away once all
	// active requests are complete. That means that ConnState
	// cannot be used to do per-request work; ConnState only notes
	// the overall state of the connection.
	StateActive

	// StateIdle represents a connection that has finished
	// handling a request and is in the keep-alive state, waiting
	// for a new request. Connections transition from StateIdle
	// to either StateActive or StateClosed.
	StateIdle

	// StateHijacked represents a hijacked connection.
	// This is a terminal state. It does not transition to StateClosed.
	StateHijacked

	// StateClosed represents a closed connection.
	// This is a terminal state. Hijacked connections do not
	// transition to StateClosed.
	StateClosed
)

var stateName = map[ConnState]string{
	StateNew:      "new",
	StateActive:   "active",
	StateIdle:     "idle",
	StateHijacked: "hijacked",
	StateClosed:   "closed",
}

func (c ConnState) String() string {
	return stateName[c]
}

var (
	ErrAbortHandler    = errors.New("net/http: abort Handler")
	ErrBodyNotAllowed  = errors.New("http: request method or response status code does not allow body")
	ErrHijacked        = errors.New("http: connection has been hijacked")
	ErrContentLength   = errors.New("http: wrote more than the declared Content-Length")
	ErrWriteAfterFlush = errors.New("unused")
	ErrNotSupported    = errors.New("feature not supported")
	ErrRequestCanceled = errors.New("net/http: request canceled")
)

// TODO: The Header type here won't match.
type PushOptions struct {
	Method string
	Header Header
}
