// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tls

import (
	"context"
	"errors"
	"fmt"
)

// EncryptionLevel represents a QUIC encryption level used to transmit
// handshake messages.
type EncryptionLevel int

const (
	EncryptionLevelInitial = EncryptionLevel(iota)
	EncryptionLevelHandshake
	EncryptionLevelApplication
)

func (l EncryptionLevel) String() string {
	switch l {
	case EncryptionLevelInitial:
		return "Initial"
	case EncryptionLevelHandshake:
		return "Handshake"
	case EncryptionLevelApplication:
		return "Application"
	default:
		return fmt.Sprintf("EncryptionLevel(%v)", int(l))
	}
}

// A QUICConn represents a connection which uses a QUIC implementation as the underlying
// transport as described in RFC 9001.
//
// Methods of QUICConn are not safe for concurrent use.
//
// A QUICConn is created with a QUICTransport containing a set of functions used to
// inform the underlying transport of key changes, new data to send in CRYPTO frames,
// and so forth. These functions are called synchronously by Start and HandleCryptoData,
// never asynchronously in the background.
type QUICConn struct {
	conn *Conn
}

type quicState struct {
	transport *QUICTransport
	readc     chan struct{}   // handshake data is available to be read
	readingc  chan struct{}   // handshake is waiting for data, or closed when done
	cancelc   <-chan struct{} // handshake has been canceled
	cancel    context.CancelFunc

	// readbuf is shared between HandleCryptoData and the handshake goroutine.
	// HandshakeCryptoData passes ownership to the handshake goroutine by
	// reading from readc, and reclaims ownership by reading from readingc.
	readbuf []byte
}

// QUICClient returns a new TLS client side connection using QUICTransport as the
// underlying transport. The config cannot be nil.
//
// The config's MinVersion must be at least TLS 1.3.
func QUICClient(t *QUICTransport, config *Config) *QUICConn {
	q := &QUICConn{
		conn: Client(nil, config),
	}
	q.conn.quic = &quicState{
		transport: t,
	}
	return q
}

// QUICServer returns a new TLS server side connection using QUICTransport as the
// underlying transport. The config cannot be nil.
//
// The config's MinVersion must be at least TLS 1.3.
func QUICServer(t *QUICTransport, config *Config) *QUICConn {
	q := &QUICConn{
		conn: Server(nil, config),
	}
	q.conn.quic = &quicState{
		transport: t,
	}
	return q
}

// Start starts the client or server handshake protocol.
//
// Start provides the initial flight of CRYPTO data (if any) to the QUICTransport
// and returns immediately. It does not block until the handshake is complete.
//
// The Context parameter becomes the parent of the handshake context returned
// by ClientHelloInfo.Context. Canceling the context will abort the handshake.
//
// Start must be called at most once.
func (q *QUICConn) Start(ctx context.Context) error {
	if q.conn.config.MinVersion < VersionTLS13 {
		return quicError(errors.New("tls: Config MinVersion must be at least TLS 1.13"))
	}
	q.conn.quic.readc = make(chan struct{})
	q.conn.quic.readingc = make(chan struct{})
	go q.conn.HandshakeContext(ctx)
	if _, ok := <-q.conn.quic.readingc; !ok {
		return q.conn.handshakeErr
	}
	return nil
}

// Close closes the connection and stops any in-progress handshake.
func (q *QUICConn) Close() error {
	if q.conn.quic.cancel == nil {
		return nil // never started
	}
	q.conn.quic.cancel()
	for range q.conn.quic.readingc {
		// Wait for the handshake goroutine to return.
	}
	return q.conn.handshakeErr
	return nil
}

// HandleCryptoData handles handshake bytes received from the peer.
func (q *QUICConn) HandleCryptoData(level EncryptionLevel, data []byte) error {
	c := q.conn
	if c.in.level != level {
		return quicError(c.in.setErrorLocked(errors.New("tls: handshake data received at wrong level")))
	}
	c.quic.readbuf = data
	<-c.quic.readc
	_, ok := <-c.quic.readingc
	if ok {
		// The handshake goroutine is waiting for more data.
		return nil
	}
	c.hand.Write(c.quic.readbuf)
	c.quic.readbuf = nil
	for q.conn.hand.Len() >= 4 && q.conn.handshakeErr == nil {
		b := q.conn.hand.Bytes()
		n := int(b[1])<<16 | int(b[2])<<8 | int(b[3])
		if 4+n < len(b) {
			return nil
		}
		if err := q.conn.handlePostHandshakeMessage(); err != nil {
			return quicError(err)
		}
	}
	c.out.Lock()
	defer c.out.Unlock()
	return quicError(c.out.err)
}

// ConnectionState returns basic TLS details about the connection.
func (q *QUICConn) ConnectionState() ConnectionState {
	return q.conn.ConnectionState()
}

// QUICTransport describes hooks used by a QUIC implementation.
// Functions may be nil, in which case they will not be called.
//
// If any QUICTransport function returns an error, the QUIC handshake will
// be terminated.
type QUICTransport struct {
	// SetReadSecret configures the read secret and cipher suite for the given
	// encryption level. It will be called at most once per encryption level.
	//
	// QUIC ACKs packets at the same level they were received at, except that
	// early data (0-RTT) packets trigger application (1-RTT) acks. ACK-writing
	// keys will always be installed with SetWriteSecret before the
	// packet-reading keys with SetReadSecret, ensuring that QUIC can always
	// ACK any packet that it decrypts.
	//
	// Keys are not provided for the Initial encryption level.
	SetReadSecret func(level EncryptionLevel, suite uint16, secret []byte)

	// SetWriteSecret configures the write secret and cipher suite for the
	// given encryption level. It will be called at most once per encryption
	// level.
	//
	// See SetReadSecret for additional invariants between packets and their
	// ACKs.
	SetWriteSecret func(level EncryptionLevel, suite uint16, secret []byte)

	// WriteCryptoData adds handshake data to the current flight at the
	// given encryption level.
	WriteCryptoData func(level EncryptionLevel, data []byte) error

	// SetTransportParameters provides the extension_data field of the
	// quic_transport_parameters extension sent by the peer.
	//
	// For client connections, SetTransportParameters will be called before
	// EncryptionLevelApplication keys are installed with SetWriteSecret.
	// For server connections, SetTransportParameters will be called before
	// EncryptionLevelHandshake keys are installed with SetWriteSecret.
	SetTransportParameters func([]byte) error

	// GetTransportParameters returns the extension_data field of the
	// quic_transport_parameters extension to send to the peer.
	GetTransportParameters func() []byte

	// HandshakeComplete is called when the handshake concludes successfully.
	//
	// Unsuccessful completion is indicated by an error returned by
	// Start or HandleCryptoData.
	HandshakeComplete func()
}

// withErr wraps multiple errors, using the error text of the first.
type withErr struct {
	errs []error
}

func (e withErr) Error() string   { return e.errs[0].Error() }
func (e withErr) Unwrap() []error { return e.errs }

// wrapErr wraps an error, ensuring that inspection of the error must
// use errors.Is/errors.As.
type wrapErr struct {
	error
}

func (e wrapErr) Unwrap() error { return e.error }

// quicError ensures err is an AlertError.
// If err is not already, quicError wraps it with alertInternalError.
func quicError(err error) error {
	if err == nil {
		return nil
	}
	if _, ok := err.(AlertError); ok {
		return wrapErr{err}
	}
	var a AlertError
	if err == nil || errors.As(err, &a) {
		return err
	}
	return withErr{[]error{err, alertInternalError}}
}

func (c *Conn) quicSetReadSecret(level EncryptionLevel, suite uint16, secret []byte) {
	if c.quic == nil || c.quic.transport.SetReadSecret == nil {
		return
	}
	c.handshakeMutex.Unlock()
	defer c.handshakeMutex.Lock()
	c.quic.transport.SetReadSecret(level, suite, secret)
}

func (c *Conn) quicSetWriteSecret(level EncryptionLevel, suite uint16, secret []byte) {
	if c.quic == nil || c.quic.transport.SetWriteSecret == nil {
		return
	}
	c.handshakeMutex.Unlock()
	defer c.handshakeMutex.Lock()
	c.quic.transport.SetWriteSecret(level, suite, secret)
}

func (c *Conn) quicWriteCryptoData(level EncryptionLevel, data []byte) error {
	if c.quic == nil || c.quic.transport.WriteCryptoData == nil {
		return nil
	}
	c.handshakeMutex.Unlock()
	defer c.handshakeMutex.Lock()
	return c.quic.transport.WriteCryptoData(level, data)
}

func (c *Conn) quicSetTransportParameters(params []byte) error {
	if c.quic == nil || c.quic.transport.SetTransportParameters == nil {
		return nil
	}
	c.handshakeMutex.Unlock()
	defer c.handshakeMutex.Lock()
	return c.quic.transport.SetTransportParameters(params)
}

func (c *Conn) quicGetTransportParameters() []byte {
	if c.quic == nil || c.quic.transport.GetTransportParameters == nil {
		return nil
	}
	c.handshakeMutex.Unlock()
	defer c.handshakeMutex.Lock()
	return c.quic.transport.GetTransportParameters()
}

func (c *Conn) quicHandshakeComplete() {
	if c.quic == nil || c.quic.transport.HandshakeComplete == nil {
		return
	}
	c.handshakeMutex.Unlock()
	defer c.handshakeMutex.Lock()
	c.quic.transport.HandshakeComplete()
}
