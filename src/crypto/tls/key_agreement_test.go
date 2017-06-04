package tls

import (
	"crypto/rand"
	"testing"
)

func TestClientMasksHighestBit(t *testing.T) {
	// See https://github.com/golang/go/issues/20582
	config := new(Config)
	config.Rand = rand.Reader
	ka := new(ecdheKeyAgreement)
	ka.curveid = X25519

	// Checks whether the highest bit of the X25519 public key is "always" zero - if so fail
	n, tries := 1000, 1000
	for tries > 0 {
		_, ckex, err := ka.generateClientKeyExchange(config, nil, nil)
		if err != nil {
			t.Errorf("Failed to generate client key exchange: %v\n", err)
			break
		}
		if (ckex.ciphertext[1+31] & 0x80) != 0 {
			return
		}
		tries--
	}
	t.Errorf("Highest bit of X25519 public key was never set after %d tries\n", n-tries)
}

func TestServerMasksHighestBit(t *testing.T) {
	// See https://github.com/golang/go/issues/20582
	config := new(Config)
	config.Rand = rand.Reader
	config.CurvePreferences = append(config.CurvePreferences, X25519)

	ka := new(ecdheKeyAgreement)
	ka.version = VersionTLS11
	ka.curveid = X25519
	ka.sigType = signatureRSA

	clientHello := new(clientHelloMsg)
	clientHello.supportedCurves = append(clientHello.supportedCurves, X25519)

	hello := new(serverHelloMsg)
	cert := testConfig.Certificates[0]

	// Checks whether the highest bit of the X25519 public key is "always" zero - if so fail
	n, tries := 1000, 1000
	for tries > 0 {
		skex, err := ka.generateServerKeyExchange(config, &cert, clientHello, hello)
		if err != nil {
			t.Errorf("Failed to generate server key exchange: %v\n", err)
			break
		}
		if (skex.key[4+31] & 0x80) != 0 {
			return
		}
		tries--
	}
	t.Errorf("Highest bit of X25519 public key was never set after %d tries\n", n-tries)
}
