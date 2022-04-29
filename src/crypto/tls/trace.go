// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tls

import (
	"context"
)

// unique type to prevent assignment.
type tlsEventContextKey struct{}

// ContextTlsTrace returns the TlsTrace associated with the
// provided context. If none, it returns nil.
func ContextTlsTrace(ctx context.Context) *TlsTrace {
	tlsTrace, _ := ctx.Value(tlsEventContextKey{}).(*TlsTrace)
	return tlsTrace
}

// WithTlsTrace returns a new context based on the provided parent
// ctx. TLS connections created with a context that contains a pointer
// to a TlsTrace will have the callbacks invoked at the appropriate
// times during its life. Any callback in the TlsTrace structure
// can be nil.
func WithTlsTrace(ctx context.Context, trace *TlsTrace) context.Context {
	if trace == nil {
		panic("nil trace")
	}
	ctx = context.WithValue(ctx, tlsEventContextKey{}, trace)
	return ctx
}

// TlsTrace is a set of hooks to run at various stages of an ongoing
// TLS connection. Any particular hook may be nil. Functions may be
// called concurrently from different goroutines and some may be called
// after the request has completed or failed.
//
// ClientTrace currently traces a single TLS connection.
type TlsTrace struct {
	// TLSHandshakeStart is called immediately before the TLS handshake is
	// started. When connecting to an HTTPS site via an HTTP proxy, the
	// handshake happens after the CONNECT request is processed by the proxy.
	TLSHandshakeStart func()

	// TLSHandshakeDone is called after the TLS handshake completes with
	// either information about the successful handshake's connection state,
	// or a non-nil error on handshake failure.
	TLSHandshakeDone func(ConnectionState, error)
}
