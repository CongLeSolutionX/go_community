// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tls

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
)

type testQUICConn struct {
	t           *testing.T
	conn        *QUICConn
	writeLevel  EncryptionLevel
	writeBuf    map[EncryptionLevel][]byte
	readSecret  map[EncryptionLevel]suiteSecret
	writeSecret map[EncryptionLevel]suiteSecret
	gotParams   []byte
	sendParams  []byte
	complete    bool
	transcript  strings.Builder

	err      error
	errAfter int
	ops      int
	errCause string
}

func newTestQUICClient(t *testing.T, config *Config) *testQUICConn {
	q := &testQUICConn{t: t}
	q.conn = QUICClient(q.transport(), config)
	t.Cleanup(func() {
		q.conn.Close()
	})
	return q
}

func newTestQUICServer(t *testing.T, config *Config) *testQUICConn {
	q := &testQUICConn{t: t}
	q.conn = QUICServer(q.transport(), config)
	t.Cleanup(func() {
		q.conn.Close()
	})
	return q
}

func (q *testQUICConn) transport() *QUICTransport {
	return &QUICTransport{
		SetReadSecret:          q.SetReadSecret,
		SetWriteSecret:         q.SetWriteSecret,
		WriteCryptoData:        q.WriteCryptoData,
		SetTransportParameters: q.SetTransportParameters,
		GetTransportParameters: q.GetTransportParameters,
		HandshakeComplete:      q.HandshakeComplete,
	}
}

type suiteSecret struct {
	suite  uint16
	secret []byte
}

func (q *testQUICConn) SetReadSecret(level EncryptionLevel, suite uint16, secret []byte) {
	if _, ok := q.writeSecret[level]; !ok {
		q.t.Errorf("SetReadSecret for level %v called before SetWriteSecret", level)
	}
	if level == EncryptionLevelApplication && !q.complete {
		q.t.Errorf("SetReadSecret for level %v called before HandshakeComplete", level)
	}
	if _, ok := q.readSecret[level]; ok {
		q.t.Errorf("SetReadSecret for level %v called twice", level)
	}
	if q.readSecret == nil {
		q.readSecret = map[EncryptionLevel]suiteSecret{}
	}
	switch level {
	case EncryptionLevelHandshake, EncryptionLevelApplication:
		q.readSecret[level] = suiteSecret{suite, secret}
	default:
		q.t.Errorf("SetReadSecret for unexpected level %v", level)
	}
}

func (q *testQUICConn) SetWriteSecret(level EncryptionLevel, suite uint16, secret []byte) {
	if _, ok := q.writeSecret[level]; ok {
		q.t.Errorf("SetWriteSecret for level %v called twice", level)
	}
	if q.writeSecret == nil {
		q.writeSecret = map[EncryptionLevel]suiteSecret{}
	}
	switch level {
	case EncryptionLevelHandshake, EncryptionLevelApplication:
		q.writeSecret[level] = suiteSecret{suite, secret}
	default:
		q.t.Errorf("SetWriteSecret for unexpected level %v", level)
	}
	q.writeLevel = level
}

func (q *testQUICConn) WriteCryptoData(level EncryptionLevel, data []byte) error {
	if q.writeLevel != level {
		q.t.Errorf("WriteCryptoData at level %v, but last write secret was %v", level, q.writeLevel)
	}
	if q.writeBuf == nil {
		q.writeBuf = make(map[EncryptionLevel][]byte)
	}
	q.writeBuf[level] = append(q.writeBuf[level], data...)
	return q.maybeErr("WriteCryptoData")
}

func (q *testQUICConn) SetTransportParameters(p []byte) error {
	q.gotParams = p
	return q.maybeErr("SetTransportParameters")
}

func (q *testQUICConn) GetTransportParameters() []byte {
	return q.sendParams
}

func (q *testQUICConn) HandshakeComplete() {
	q.complete = true
}

func (q *testQUICConn) maybeErr(cause string) error {
	q.ops++
	if q.ops == q.errAfter {
		q.errCause = cause
		return q.err
	}
	if q.errCause != "" {
		q.t.Errorf("%v called after earlier error from %v", cause, q.errCause)
	}
	return nil
}

func runTestQUICConnection(ctx context.Context, a, b *testQUICConn, onHandleCryptoData func()) error {
	if err := a.conn.Start(ctx); err != nil {
		return err
	}
	if err := b.conn.Start(ctx); err != nil {
		return err
	}
	idleCount := 0
	for {
		idleCount++
		for _, level := range []EncryptionLevel{
			EncryptionLevelInitial,
			EncryptionLevelHandshake,
			EncryptionLevelApplication,
		} {
			if len(a.writeBuf[level]) == 0 {
				continue
			}
			idleCount = 0
			if err := b.conn.HandleCryptoData(level, a.writeBuf[level]); err != nil {
				return err
			}
			a.writeBuf[level] = nil
			if onHandleCryptoData != nil {
				onHandleCryptoData()
			}
		}
		if idleCount == 2 {
			break
		}
		a, b = b, a
	}
	return nil
}

func TestQUICConnection(t *testing.T) {
	config := testConfig.Clone()
	config.MinVersion = VersionTLS13

	cli := newTestQUICClient(t, config)
	srv := newTestQUICServer(t, config)

	if err := runTestQUICConnection(context.Background(), cli, srv, nil); err != nil {
		t.Fatalf("error during connection handshake: %v", err)
	}

	if _, ok := cli.readSecret[EncryptionLevelHandshake]; !ok {
		t.Errorf("client has no Handshake secret")
	}
	if _, ok := cli.readSecret[EncryptionLevelApplication]; !ok {
		t.Errorf("client has no Application secret")
	}
	if _, ok := srv.readSecret[EncryptionLevelHandshake]; !ok {
		t.Errorf("server has no Handshake secret")
	}
	if _, ok := srv.readSecret[EncryptionLevelApplication]; !ok {
		t.Errorf("server has no Application secret")
	}
	for _, level := range []EncryptionLevel{EncryptionLevelHandshake, EncryptionLevelApplication} {
		if _, ok := cli.readSecret[level]; !ok {
			t.Errorf("client has no %v read secret", level)
		}
		if _, ok := srv.readSecret[level]; !ok {
			t.Errorf("server has no %v read secret", level)
		}
		if !reflect.DeepEqual(cli.readSecret[level], srv.writeSecret[level]) {
			t.Errorf("client read secret does not match server write secret for level %v", level)
		}
		if !reflect.DeepEqual(cli.writeSecret[level], srv.readSecret[level]) {
			t.Errorf("client write secret does not match server read secret for level %v", level)
		}
	}
}

func TestQUICSessionResumption(t *testing.T) {
	clientConfig := testConfig.Clone()
	clientConfig.MinVersion = VersionTLS13
	clientConfig.ClientSessionCache = NewLRUClientSessionCache(1)
	clientConfig.ServerName = "example.go.dev"

	serverConfig := testConfig.Clone()
	serverConfig.MinVersion = VersionTLS13

	cli := newTestQUICClient(t, clientConfig)
	srv := newTestQUICServer(t, serverConfig)
	if err := runTestQUICConnection(context.Background(), cli, srv, nil); err != nil {
		t.Fatalf("error during first connection handshake: %v", err)
	}
	if cli.conn.ConnectionState().DidResume {
		t.Errorf("first connection unexpectedly used session resumption")
	}

	cli2 := newTestQUICClient(t, clientConfig)
	srv2 := newTestQUICServer(t, serverConfig)
	if err := runTestQUICConnection(context.Background(), cli2, srv2, nil); err != nil {
		t.Fatalf("error during second connection handshake: %v", err)
	}
	if !cli2.conn.ConnectionState().DidResume {
		t.Errorf("second connection did not use session resumption")
	}
}

func TestQUICConnectionErrors(t *testing.T) {
	config := testConfig.Clone()
	config.MinVersion = VersionTLS13
	wantErr := errors.New("error")

	for _, clientErr := range []bool{false, true} {
		name := "server"
		if clientErr {
			name = "client"
		}
		t.Run(name, func(t *testing.T) {
			errAfter := 0
			for {
				errAfter++
				cli := newTestQUICClient(t, config)
				srv := newTestQUICServer(t, config)

				failing := srv
				if clientErr {
					failing = cli
				}
				failing.err = wantErr
				failing.errAfter = errAfter

				gotErr := runTestQUICConnection(context.Background(), cli, srv, nil)
				if errors.Is(gotErr, wantErr) {
					continue
				}
				if gotErr == nil && failing.ops < errAfter {
					return
				}
				if gotErr != nil {
					t.Fatalf("unexpected handshake result (failure on op %v, client=%v): %v", errAfter, clientErr, gotErr)
				}
			}
		})
	}
}

func TestQUICPostHandshakeClientAuthentication(t *testing.T) {
	// RFC 9001, Section 4.4.
	config := testConfig.Clone()
	config.MinVersion = VersionTLS13
	cli := newTestQUICClient(t, config)
	srv := newTestQUICServer(t, config)
	if err := runTestQUICConnection(context.Background(), cli, srv, nil); err != nil {
		t.Fatalf("error during connection handshake: %v", err)
	}

	certReq := new(certificateRequestMsgTLS13)
	certReq.ocspStapling = true
	certReq.scts = true
	certReq.supportedSignatureAlgorithms = supportedSignatureAlgorithms()
	certReqBytes, err := certReq.marshal()
	if err != nil {
		t.Fatal(err)
	}
	if err := cli.conn.HandleCryptoData(EncryptionLevelApplication, append([]byte{
		byte(typeCertificateRequest),
		byte(0), byte(0), byte(len(certReqBytes)),
	}, certReqBytes...)); err == nil {
		t.Fatalf("post-handshake authentication request: got no error, want one")
	}
}

func TestQUICPostHandshakeKeyUpdate(t *testing.T) {
	// RFC 9001, Section 6.
	config := testConfig.Clone()
	config.MinVersion = VersionTLS13
	cli := newTestQUICClient(t, config)
	srv := newTestQUICServer(t, config)
	if err := runTestQUICConnection(context.Background(), cli, srv, nil); err != nil {
		t.Fatalf("error during connection handshake: %v", err)
	}

	keyUpdate := new(keyUpdateMsg)
	keyUpdateBytes, err := keyUpdate.marshal()
	if err != nil {
		t.Fatal(err)
	}
	if err := cli.conn.HandleCryptoData(EncryptionLevelApplication, append([]byte{
		byte(typeKeyUpdate),
		byte(0), byte(0), byte(len(keyUpdateBytes)),
	}, keyUpdateBytes...)); !errors.Is(err, alertUnexpectedMessage) {
		t.Fatalf("key update request: got error %v, want alertUnexpectedMessage", err)
	}
}

func TestQUICHandshakeError(t *testing.T) {
	clientConfig := testConfig.Clone()
	clientConfig.MinVersion = VersionTLS13
	clientConfig.InsecureSkipVerify = false
	clientConfig.ServerName = "name"

	serverConfig := testConfig.Clone()
	serverConfig.MinVersion = VersionTLS13

	cli := newTestQUICClient(t, clientConfig)
	srv := newTestQUICServer(t, serverConfig)
	err := runTestQUICConnection(context.Background(), cli, srv, nil)
	if !errors.Is(err, alertBadCertificate) {
		t.Errorf("connection handshake terminated with error %q, want alertBadCertificate", err)
	}
	var e *CertificateVerificationError
	if !errors.As(err, &e) {
		t.Errorf("connection handshake terminated with error %q, want CertificateVerificationError", err)
	}
}

// Test that QUICConn.ConnectionState can be used during the handshake,
// and that it reports the application protocol as soon as it has been
// negotiated.
func TestQUICConnectionState(t *testing.T) {
	config := testConfig.Clone()
	config.MinVersion = VersionTLS13
	config.NextProtos = []string{"h3"}
	cli := newTestQUICClient(t, config)
	srv := newTestQUICServer(t, config)
	onHandleCryptoData := func() {
		cliCS := cli.conn.ConnectionState()
		cliWantALPN := ""
		if _, ok := cli.readSecret[EncryptionLevelApplication]; ok {
			cliWantALPN = "h3"
		}
		if want, got := cliCS.NegotiatedProtocol, cliWantALPN; want != got {
			t.Errorf("cli.ConnectionState().NegotiatedProtocol = %q, want %q", want, got)
		}

		srvCS := srv.conn.ConnectionState()
		srvWantALPN := ""
		if _, ok := srv.readSecret[EncryptionLevelHandshake]; ok {
			srvWantALPN = "h3"
		}
		if want, got := srvCS.NegotiatedProtocol, srvWantALPN; want != got {
			t.Errorf("srv.ConnectionState().NegotiatedProtocol = %q, want %q", want, got)
		}
	}
	if err := runTestQUICConnection(context.Background(), cli, srv, onHandleCryptoData); err != nil {
		t.Fatalf("error during connection handshake: %v", err)
	}
}

func TestQUICStartContextPropagation(t *testing.T) {
	const key = "key"
	const value = "value"
	ctx := context.WithValue(context.Background(), key, value)
	config := testConfig.Clone()
	config.MinVersion = VersionTLS13
	calls := 0
	config.GetConfigForClient = func(info *ClientHelloInfo) (*Config, error) {
		calls++
		got, _ := info.Context().Value(key).(string)
		if got != value {
			t.Errorf("GetConfigForClient context key %q has value %q, want %q", key, got, value)
		}
		return nil, nil
	}
	cli := newTestQUICClient(t, config)
	srv := newTestQUICServer(t, config)
	if err := runTestQUICConnection(ctx, cli, srv, nil); err != nil {
		t.Fatalf("error during connection handshake: %v", err)
	}
	if calls != 1 {
		t.Errorf("GetConfigForClient called %v times, want 1", calls)
	}
}
