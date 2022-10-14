// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tls

import (
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

// An AlertError is a TLS alert.
//
// When using a QUIC transport, Conn.Handshake will return an error wrapping AlertError
// rather than sending a TLS alert.
type AlertError = alert

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
	*QUICTransport

	readc    chan struct{}
	readingc chan struct{}
	closec   chan struct{}
	readbuf  []byte
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
		QUICTransport: t,
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
		QUICTransport: t,
	}
	return q
}

// Start starts the client or server handshake protocol.
// It must be called at most once.
func (q *QUICConn) Start() error {
	if q.conn.config.MinVersion < VersionTLS13 {
		return errors.New("tls: Config MinVersion must be at least TLS 1.13")
	}
	q.conn.quic.readc = make(chan struct{})
	q.conn.quic.readingc = make(chan struct{})
	q.conn.quic.closec = make(chan struct{})
	go q.conn.Handshake()
	if _, ok := <-q.conn.quic.readingc; !ok {
		return q.conn.handshakeErr
	}
	return nil
}

// Close closes the connection and stops any in-progress handshake.
func (q *QUICConn) Close() error {
	if q.conn.quic.closec == nil {
		return nil // never started
	}
	select {
	case <-q.conn.quic.closec:
		return nil // already closed
	default:
	}
	// Synchronous Close calls might race and double-close this channel,
	// but we don't allow concurrent use of QUICConn.
	close(q.conn.quic.closec)
	return nil
}

// HandleCryptoData handles handshake bytes received from the peer.
func (q *QUICConn) HandleCryptoData(level EncryptionLevel, data []byte) error {
	c := q.conn
	if c.in.level != level {
		return c.in.setErrorLocked(errors.New("tls: handshake data received at wrong level"))
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
			return err
		}
	}
	return q.conn.handshakeErr
}

// ConnectionState returns basic TLS details about the connection.
func (q *QUICConn) ConnectionState() ConnectionState {
	return q.conn.ConnectionState()
}

// QUICTransport describes hooks used by a QUIC implementation.
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

	// HandshakeDone is called when the handshake completes successfully.
	//
	// Unsuccessful completion is indicated by an error returned by
	// Start or HandleCryptoData.
	HandshakeDone func()
}
