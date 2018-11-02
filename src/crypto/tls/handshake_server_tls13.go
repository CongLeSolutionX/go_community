// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tls

import (
	"bytes"
	"crypto"
	"crypto/hmac"
	"crypto/rsa"
	"errors"
	"fmt"
	"hash"
	"io"
	"sync/atomic"
)

type serverHandshakeStateTLS13 struct {
	c               *Conn
	clientHello     *clientHelloMsg
	hello           *serverHelloMsg
	suite           *cipherSuiteTLS13
	cert            *Certificate
	sigAlg          SignatureScheme
	handshakeSecret []byte
	trafficSecret   []byte // client_application_traffic_secret_0
	transcript      hash.Hash
}

func (hs *serverHandshakeStateTLS13) handshake() error {
	c := hs.c

	// For an overview of the TLS 1.3 handshake, see RFC 8446, Section 2.
	if err := hs.processClientHello(); err != nil {
		return err
	}
	c.buffering = true
	if err := hs.sendServerParameters(); err != nil {
		return err
	}
	if err := hs.sendServerCertificate(); err != nil {
		return err
	}
	if err := hs.sendServerFinished(); err != nil {
		return err
	}
	if _, err := c.flush(); err != nil {
		return err
	}
	if err := hs.readClientFinished(); err != nil {
		return err
	}

	atomic.StoreUint32(&c.handshakeStatus, 1)

	return nil
}

func (hs *serverHandshakeStateTLS13) processClientHello() error {
	c := hs.c

	hs.hello = new(serverHelloMsg)

	// TLS 1.3 froze the ServerHello.legacy_version field, and uses
	// supported_versions instead. See RFC 8446, Section 4.1.3 and 4.2.1.
	hs.hello.vers = VersionTLS12
	hs.hello.supportedVersion = c.vers

	if len(hs.clientHello.supportedVersions) == 0 {
		c.sendAlert(alertIllegalParameter)
		return errors.New("tls: client used the legacy version field to negotiate TLS 1.3")
	}

	if len(hs.clientHello.compressionMethods) != 1 ||
		hs.clientHello.compressionMethods[0] != compressionNone {
		c.sendAlert(alertIllegalParameter)
		return errors.New("tls: TLS 1.3 client supports illegal compression methods")
	}

	hs.hello.random = make([]byte, 32)
	if _, err := io.ReadFull(c.config.rand(), hs.hello.random); err != nil {
		c.sendAlert(alertInternalError)
		return err
	}

	if len(hs.clientHello.secureRenegotiation) != 0 {
		c.sendAlert(alertHandshakeFailure)
		return errors.New("tls: initial handshake had non-empty renegotiation extension")
	}

	if hs.clientHello.earlyData {
		return errors.New("tls: early data skipping not implemented") // TODO(filippo)
	}

	hs.hello.sessionId = hs.clientHello.sessionId
	hs.hello.compressionMethod = compressionNone

	var preferenceList, supportedList []uint16
	if c.config.PreferServerCipherSuites {
		preferenceList = defaultCipherSuitesTLS13()
		supportedList = hs.clientHello.cipherSuites
	} else {
		preferenceList = hs.clientHello.cipherSuites
		supportedList = defaultCipherSuitesTLS13()
	}
	for _, suiteID := range preferenceList {
		hs.suite = mutualCipherSuiteTLS13(supportedList, suiteID)
		if hs.suite != nil {
			break
		}
	}
	if hs.suite == nil {
		c.sendAlert(alertHandshakeFailure)
		return errors.New("tls: no cipher suite supported by both client and server")
	}
	c.cipherSuite = hs.suite.id
	hs.hello.cipherSuite = hs.suite.id
	hs.transcript = hs.suite.hash.New()

	for _, id := range hs.clientHello.cipherSuites {
		if id == TLS_FALLBACK_SCSV { // See RFC 7507.
			// The client is doing a fallback connection.
			if hs.clientHello.vers < c.config.supportedVersions(false)[0] {
				c.sendAlert(alertInappropriateFallback)
				return errors.New("tls: client using inappropriate protocol fallback")
			}
			break
		}
	}

	// Pick the ECDHE group in server preference order, but give priority to
	// groups with a key share, to avoid a HelloRetryRequest round-trip.
	var selectedGroup CurveID
	var clientKeyShare *keyShare
GroupSelection:
	for _, preferredGroup := range c.config.curvePreferences() {
		for _, ks := range hs.clientHello.keyShares {
			if ks.group == preferredGroup {
				selectedGroup = ks.group
				clientKeyShare = &ks
				break GroupSelection
			}
		}
		if selectedGroup != 0 {
			continue
		}
		for _, group := range hs.clientHello.supportedCurves {
			if group == preferredGroup {
				selectedGroup = group
				break
			}
		}
	}
	if selectedGroup == 0 {
		c.sendAlert(alertHandshakeFailure)
		return errors.New("tls: no ECDHE curve supported by both client and server")
	}
	if clientKeyShare == nil {
		if err := hs.doHelloRetryRequest(selectedGroup); err != nil {
			return err
		}
		clientKeyShare = &hs.clientHello.keyShares[0]
	}

	if _, ok := curveForCurveID(selectedGroup); selectedGroup != X25519 && !ok {
		c.sendAlert(alertInternalError)
		return errors.New("tls: CurvePreferences includes unsupported curve")
	}
	params, err := generateECDHEParameters(c.config.rand(), selectedGroup)
	if err != nil {
		c.sendAlert(alertInternalError)
		return err
	}
	hs.hello.serverShare = keyShare{group: selectedGroup, data: params.PublicKey()}
	sharedKey := params.SharedKey(clientKeyShare.data)
	if sharedKey == nil {
		c.sendAlert(alertIllegalParameter)
		return errors.New("tls: invalid client key share")
	}
	earlySecret := hs.suite.extract(nil, nil)
	hs.handshakeSecret = hs.suite.extract(sharedKey,
		hs.suite.deriveSecret(earlySecret, "derived", nil))

	// This implements a very simplistic certificate selection strategy for now:
	// getCertificate delegates to the application Config.GetCertificate, or
	// selects based on the server_name only. If the selected certificate's
	// public key does not match the client signature_algorithms, the handshake
	// is aborted. No attention is given to signature_algorithms_cert, and it is
	// not passed to the application Config.GetCertificate. This will need to
	// improve according to RFC 8446, Section 4.4.2.2 and 4.2.3.
	certificate, err := c.config.getCertificate(clientHelloInfo(c, hs.clientHello))
	if err != nil {
		c.sendAlert(alertInternalError)
		return err
	}
	supportedAlgs := signatureSchemesForCertificate(certificate)
	if supportedAlgs == nil {
		c.sendAlert(alertInternalError)
		return fmt.Errorf("tls: unsupported certificate key (%T)", certificate.PrivateKey)
	}
	// Pick signature scheme in client preference order, as the server
	// preference order is not configurable.
	for _, preferredAlg := range hs.clientHello.supportedSignatureAlgorithms {
		if isSupportedSignatureAlgorithm(preferredAlg, supportedAlgs) {
			hs.sigAlg = preferredAlg
			break
		}
	}
	if hs.sigAlg == 0 {
		c.sendAlert(alertHandshakeFailure)
		return errors.New("tls: client doesn't support selected certificate")
	}
	hs.cert = certificate

	return nil
}

func (hs *serverHandshakeStateTLS13) doHelloRetryRequest(selectedGroup CurveID) error {
	c := hs.c

	// The first ClientHello gets double-hashed into the transcript upon a
	// HelloRetryRequest. See RFC 8446, Section 4.4.1.
	hs.transcript.Write(hs.clientHello.marshal())
	chHash := hs.transcript.Sum(nil)
	hs.transcript.Reset()
	hs.transcript.Write([]byte{typeMessageHash, 0, 0, uint8(len(chHash))})
	hs.transcript.Write(chHash)

	helloRetryRequest := &serverHelloMsg{
		vers:              hs.hello.vers,
		random:            helloRetryRequestRandom,
		sessionId:         hs.hello.sessionId,
		cipherSuite:       hs.hello.cipherSuite,
		compressionMethod: hs.hello.compressionMethod,
		supportedVersion:  hs.hello.supportedVersion,
		selectedGroup:     selectedGroup,
	}

	hs.transcript.Write(helloRetryRequest.marshal())
	if _, err := c.writeRecord(recordTypeHandshake, helloRetryRequest.marshal()); err != nil {
		return err
	}

	msg, err := c.readHandshake()
	if err != nil {
		return err
	}

	clientHello, ok := msg.(*clientHelloMsg)
	if !ok {
		c.sendAlert(alertUnexpectedMessage)
		return unexpectedMessageError(clientHello, msg)
	}

	if len(clientHello.keyShares) != 1 || clientHello.keyShares[0].group != selectedGroup {
		c.sendAlert(alertIllegalParameter)
		return errors.New("tls: client sent invalid key share in second ClientHello")
	}

	if clientHello.earlyData {
		c.sendAlert(alertIllegalParameter)
		return errors.New("tls: client indicated early data in second ClientHello")
	}

	illegalClientHelloChange := false
	if clientHello.vers != hs.clientHello.vers ||
		!bytes.Equal(clientHello.random, hs.clientHello.random) ||
		!bytes.Equal(clientHello.sessionId, hs.clientHello.sessionId) ||
		len(clientHello.cipherSuites) != len(hs.clientHello.cipherSuites) ||
		!bytes.Equal(clientHello.compressionMethods, hs.clientHello.compressionMethods) ||
		clientHello.nextProtoNeg != hs.clientHello.nextProtoNeg ||
		clientHello.serverName != hs.clientHello.serverName ||
		clientHello.ocspStapling != hs.clientHello.ocspStapling ||
		len(clientHello.supportedCurves) != len(hs.clientHello.supportedCurves) ||
		!bytes.Equal(clientHello.supportedPoints, hs.clientHello.supportedPoints) ||
		clientHello.ticketSupported != hs.clientHello.ticketSupported ||
		!bytes.Equal(clientHello.sessionTicket, hs.clientHello.sessionTicket) ||
		len(clientHello.supportedSignatureAlgorithms) != len(hs.clientHello.supportedSignatureAlgorithms) ||
		len(clientHello.supportedSignatureAlgorithmsCert) != len(hs.clientHello.supportedSignatureAlgorithmsCert) ||
		clientHello.secureRenegotiationSupported != hs.clientHello.secureRenegotiationSupported ||
		!bytes.Equal(clientHello.secureRenegotiation, hs.clientHello.secureRenegotiation) ||
		len(clientHello.alpnProtocols) != len(hs.clientHello.alpnProtocols) ||
		clientHello.scts != hs.clientHello.scts ||
		len(clientHello.supportedVersions) != len(hs.clientHello.supportedVersions) ||
		!bytes.Equal(clientHello.cookie, hs.clientHello.cookie) ||
		!bytes.Equal(clientHello.pskModes, hs.clientHello.pskModes) {
		illegalClientHelloChange = true
	}
	for i := range hs.clientHello.cipherSuites {
		if clientHello.cipherSuites[i] != hs.clientHello.cipherSuites[i] {
			illegalClientHelloChange = true
		}
	}
	for i := range hs.clientHello.supportedCurves {
		if clientHello.supportedCurves[i] != hs.clientHello.supportedCurves[i] {
			illegalClientHelloChange = true
		}
	}
	for i := range hs.clientHello.supportedSignatureAlgorithms {
		if clientHello.supportedSignatureAlgorithms[i] != hs.clientHello.supportedSignatureAlgorithms[i] {
			illegalClientHelloChange = true
		}
	}
	for i := range hs.clientHello.supportedSignatureAlgorithmsCert {
		if clientHello.supportedSignatureAlgorithmsCert[i] != hs.clientHello.supportedSignatureAlgorithmsCert[i] {
			illegalClientHelloChange = true
		}
	}
	for i := range hs.clientHello.alpnProtocols {
		if clientHello.alpnProtocols[i] != hs.clientHello.alpnProtocols[i] {
			illegalClientHelloChange = true
		}
	}
	for i := range hs.clientHello.supportedVersions {
		if clientHello.supportedVersions[i] != hs.clientHello.supportedVersions[i] {
			illegalClientHelloChange = true
		}
	}
	if illegalClientHelloChange {
		c.sendAlert(alertIllegalParameter)
		return errors.New("tls: client illegally modified second ClientHello")
	}

	hs.clientHello = clientHello
	return nil
}

func (hs *serverHandshakeStateTLS13) sendServerParameters() error {
	c := hs.c

	hs.transcript.Write(hs.clientHello.marshal())
	hs.transcript.Write(hs.hello.marshal())
	if _, err := c.writeRecord(recordTypeHandshake, hs.hello.marshal()); err != nil {
		return err
	}

	clientSecret := hs.suite.deriveSecret(hs.handshakeSecret,
		clientHandshakeTrafficLabel, hs.transcript)
	c.in.setTrafficSecret(hs.suite, clientSecret)
	serverSecret := hs.suite.deriveSecret(hs.handshakeSecret,
		serverHandshakeTrafficLabel, hs.transcript)
	c.out.setTrafficSecret(hs.suite, serverSecret)

	encryptedExtensions := new(encryptedExtensionsMsg)

	if len(hs.clientHello.alpnProtocols) > 0 {
		if selectedProto, fallback := mutualProtocol(hs.clientHello.alpnProtocols, c.config.NextProtos); !fallback {
			encryptedExtensions.alpnProtocol = selectedProto
			c.clientProtocol = selectedProto
		}
	}

	hs.transcript.Write(encryptedExtensions.marshal())
	if _, err := c.writeRecord(recordTypeHandshake, encryptedExtensions.marshal()); err != nil {
		return err
	}

	return nil
}

func (hs *serverHandshakeStateTLS13) sendServerCertificate() error {
	c := hs.c

	certMsg := new(certificateMsgTLS13)

	certMsg.certificate = *hs.cert
	certMsg.scts = hs.clientHello.scts
	certMsg.ocspStapling = hs.clientHello.ocspStapling

	hs.transcript.Write(certMsg.marshal())
	if _, err := c.writeRecord(recordTypeHandshake, certMsg.marshal()); err != nil {
		return err
	}

	certVerifyMsg := new(certificateVerifyMsg)
	certVerifyMsg.hasSignatureAlgorithm = true
	certVerifyMsg.signatureAlgorithm = hs.sigAlg

	sigType := signatureFromSignatureScheme(hs.sigAlg)
	sigHash, err := hashFromSignatureScheme(hs.sigAlg)
	if sigType == 0 || err != nil {
		c.sendAlert(alertInternalError)
		return err
	}
	h := sigHash.New()
	writeSignedMessage(h, serverSignatureContext, hs.transcript)

	signOpts := crypto.SignerOpts(sigHash)
	if sigType == signatureRSAPSS {
		signOpts = &rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthEqualsHash, Hash: sigHash}
	}
	sig, err := hs.cert.PrivateKey.(crypto.Signer).Sign(c.config.rand(), h.Sum(nil), signOpts)
	if err != nil {
		c.sendAlert(alertInternalError)
		return errors.New("tls: failed to sign handshake: " + err.Error())
	}
	certVerifyMsg.signature = sig

	hs.transcript.Write(certVerifyMsg.marshal())
	if _, err := c.writeRecord(recordTypeHandshake, certVerifyMsg.marshal()); err != nil {
		return err
	}

	return nil
}

func (hs *serverHandshakeStateTLS13) sendServerFinished() error {
	c := hs.c

	// See RFC 8446, Section 4.4.4 and 4.4.
	finishedKey := hs.suite.expandLabel(c.out.trafficSecret, "finished", nil, hs.suite.hash.Size())
	verifyData := hmac.New(hs.suite.hash.New, finishedKey)
	verifyData.Write(hs.transcript.Sum(nil))
	finished := &finishedMsg{
		verifyData: verifyData.Sum(nil),
	}

	hs.transcript.Write(finished.marshal())
	if _, err := c.writeRecord(recordTypeHandshake, finished.marshal()); err != nil {
		return err
	}

	// Derive secrets that take context through the server Finished.

	masterSecret := hs.suite.extract(nil,
		hs.suite.deriveSecret(hs.handshakeSecret, "derived", nil))

	hs.trafficSecret = hs.suite.deriveSecret(masterSecret,
		clientApplicationTrafficLabel, hs.transcript)
	serverSecret := hs.suite.deriveSecret(masterSecret,
		serverApplicationTrafficLabel, hs.transcript)
	c.out.setTrafficSecret(hs.suite, serverSecret)

	c.ekm = hs.suite.exportKeyingMaterial(masterSecret, hs.transcript)

	return nil
}

func (hs *serverHandshakeStateTLS13) readClientFinished() error {
	c := hs.c

	msg, err := c.readHandshake()
	if err != nil {
		return err
	}

	finished, ok := msg.(*finishedMsg)
	if !ok {
		c.sendAlert(alertUnexpectedMessage)
		return unexpectedMessageError(finished, msg)
	}

	finishedKey := hs.suite.expandLabel(c.in.trafficSecret, "finished", nil, hs.suite.hash.Size())
	expectedMAC := hmac.New(hs.suite.hash.New, finishedKey)
	expectedMAC.Write(hs.transcript.Sum(nil))
	if !hmac.Equal(expectedMAC.Sum(nil), finished.verifyData) {
		c.sendAlert(alertDecryptError)
		return errors.New("tls: invalid client finished hash")
	}

	hs.transcript.Write(finished.marshal())

	c.in.setTrafficSecret(hs.suite, hs.trafficSecret)

	return nil
}
