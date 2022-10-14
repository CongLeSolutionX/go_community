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

// An AlertError is a TLS alert.
//
// When using a QUIC transport, Conn.Handshake will return an error wrapping AlertError
// rather than sending a TLS alert.
type AlertError = alert

// A QUICConn represents a connection which uses a QUIC implementation as the underlying
// transport as described in RFC 9001.
type QUICConn struct {
	conn *Conn
}

// QUICClient returns a new TLS client side connection using QUICTransport as the
// underlying transport. The config cannot be nil.
//
// The config's MinVersion must be at least TLS 1.3.
func QUICClient(t *QUICTransport, config *Config) *QUICConn {
	q := &QUICConn{
		conn: Client(nil, config),
	}
	q.conn.quic = t
	return q
}

// QUICClient returns a new TLS server side connection using QUICTransport as the
// underlying transport. The config cannot be nil.
//
// The config's MinVersion must be at least TLS 1.3.
func QUICServer(t *QUICTransport, config *Config) *QUICConn {
	q := &QUICConn{
		conn: Server(nil, config),
	}
	q.conn.quic = t
	return q
}

// Handshake runs the client or server handshake protocol.
func (q *QUICConn) Handshake() error {
	if q.conn.config.MinVersion < VersionTLS13 {
		return errors.New("tls: Config MinVersion must be at least TLS 1.13")
	}
	return q.conn.Handshake()
}

func (q *QUICConn) HandlePostHandshakeData(data []byte) (n int, err error) {
	for len(data) > 0 {
		nn, err := q.handleOnePostHandshakeMessage(data)
		if nn == 0 {
			return 0, err
		}
		n += nn
		data = data[nn:]
	}
	return n, nil
}

func (q *QUICConn) handleOnePostHandshakeMessage(data []byte) (n int, err error) {
	n = int(data[1])<<16 | int(data[2])<<8 | int(data[3])
	if n > maxHandshake {
		q.conn.sendAlertLocked(alertInternalError)
		return 0, q.conn.in.setErrorLocked(fmt.Errorf("tls: handshake message of length %d bytes exceeds maximum of %d bytes", n, maxHandshake))
	}
	n += 4
	if n > len(data) {
		return 0, nil
	}
	data = data[:n]
	msg, err := q.conn.unmarshalHandshakeMessage(data)
	if err != nil {
		return 0, err
	}
	if err := q.conn.handlePostHandshakeMessage(msg); err != nil {
		return 0, err
	}
	return n, nil
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
	SetReadSecret func(level EncryptionLevel, suite uint16, secret []byte) error

	// SetWriteSecret configures the write secret and cipher suite for the
	// given encryption level. It will be called at most once per encryption
	// level.
	//
	// See SetReadSecret for additional invariants between packets and their
	// ACKs.
	SetWriteSecret func(level EncryptionLevel, suite uint16, secret []byte) error

	// WriteHandshakeData adds handshake data to the current flight at the
	// given encryption level.
	//
	// A single handshake flight may include data from multiple encryption
	// levels. QUIC implementations should defer writing data to the network
	// until FlushHandshakeData to better pack QUIC packets into transport
	// datagrams.
	WriteHandshakeData func(level EncryptionLevel, data []byte) error

	// FlushHandshakeData is called when the current flight is complete and
	// should be written to the transport. Note that a flight may contain
	// data at several encryption levels.
	FlushHandshakeData func() error

	// ReadHandshakeData is called to request handshake data. It follows the
	// same contract as io.Reader's Read method, but returns the encryption
	// level of the data as well as the number of bytes read and error.
	//
	// ReadHandshakeData must not combine data from multiple encryption levels.
	//
	// ReadHandshakeData must block until at least one byte of data is
	// available, and must return as soon as least one byte of data is
	// available.
	ReadHandshakeData func(p []byte) (level EncryptionLevel, n int, err error)

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
	GetTransportParameters func() ([]byte, error)
}
