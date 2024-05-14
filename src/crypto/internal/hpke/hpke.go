// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hpke

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"encoding/binary"
	"errors"
	"io"
	"math/bits"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/hkdf"
)

type hkdfKDF struct {
	hash crypto.Hash
}

func (kdf *hkdfKDF) LabeledExtract(suiteID, salt, label, inputKey []byte) []byte {
	labeledIKM := append([]byte("HPKE-v1"), suiteID...)
	labeledIKM = append(labeledIKM, label...)
	labeledIKM = append(labeledIKM, inputKey...)
	return hkdf.Extract(kdf.hash.New, labeledIKM, salt)
}

func (kdf *hkdfKDF) LabeledExpand(suiteID, randomKey, label, info []byte, length uint16) []byte {
	labeledInfo := make([]byte, 0, 2+7+len(suiteID)+len(label)+len(info))
	labeledInfo = binary.BigEndian.AppendUint16(labeledInfo, length)
	labeledInfo = append(labeledInfo, []byte("HPKE-v1")...)
	labeledInfo = append(labeledInfo, suiteID...)
	labeledInfo = append(labeledInfo, label...)
	labeledInfo = append(labeledInfo, info...)
	out := make([]byte, length)
	n, err := hkdf.Expand(kdf.hash.New, randomKey, labeledInfo).Read(out)
	if err != nil || n != int(length) {
		panic("hpke: LabeledExpand failed unexpectedly")
	}
	return out
}

// dhKEM implements the KEM specified in section 4.1. of RFC 9180
type dhKEM struct {
	rand io.Reader

	dh  ecdh.Curve
	kdf hkdfKDF

	suiteID []byte
	nSecret uint16
}

var SupportedKEMs = map[uint16]struct {
	curve   ecdh.Curve
	hash    crypto.Hash
	nSecret uint16
}{
	// RFC 9180 Section 7.1
	0x0010: {ecdh.P256(), crypto.SHA256, 32},
	0x0011: {ecdh.P384(), crypto.SHA384, 48},
	0x0012: {ecdh.P521(), crypto.SHA512, 64},
	0x0020: {ecdh.X25519(), crypto.SHA256, 32},
	// We don't support 0x0021 (X448-HKDF-SHA512)
}

func newDHKem(rand io.Reader, kemID uint16) (*dhKEM, error) {
	suite, ok := SupportedKEMs[kemID]
	if !ok {
		return nil, errors.New("unknown suite ID")
	}
	suiteID := make([]byte, 0, 3+2)
	suiteID = append(suiteID, []byte("KEM")...)
	suiteID = binary.BigEndian.AppendUint16(suiteID, kemID)
	return &dhKEM{
		rand:    rand,
		dh:      suite.curve,
		kdf:     hkdfKDF{suite.hash},
		suiteID: suiteID,
		nSecret: suite.nSecret,
	}, nil
}

func (dh *dhKEM) ExtractAndExpand(dhKey, kemContext []byte) []byte {
	eaePRK := dh.kdf.LabeledExtract(dh.suiteID[:], []byte(""), []byte("eae_prk"), dhKey)
	return dh.kdf.LabeledExpand(dh.suiteID[:], eaePRK, []byte("shared_secret"), kemContext, dh.nSecret)
}

func (dh *dhKEM) Encap(pubRecipient *ecdh.PublicKey) ([]byte, []byte, error) {
	privEph, err := dh.dh.GenerateKey(dh.rand)
	if err != nil {
		return nil, nil, err
	}
	dhVal, err := privEph.ECDH(pubRecipient)
	if err != nil {
		return nil, nil, err
	}
	encPubEph := privEph.PublicKey().Bytes()

	encPubRecip := pubRecipient.Bytes()
	kemContext := append(encPubEph, encPubRecip...)

	sharedSecret := dh.ExtractAndExpand(dhVal, kemContext)
	return sharedSecret, encPubEph, nil
}

func (dh *dhKEM) Decap(enc []byte, privRecipient *ecdh.PrivateKey) ([]byte, error) {
	pubEph, err := dh.dh.NewPublicKey(enc)
	if err != nil {
		return nil, err
	}
	dhVal, err := privRecipient.ECDH(pubEph)
	if err != nil {
		return nil, err
	}

	encPubRecipient := privRecipient.PublicKey().Bytes()
	kemContext := append(enc, encPubRecipient...)

	return dh.ExtractAndExpand(dhVal, kemContext), nil
}

func (dh *dhKEM) AuthEncap(pubRecipient *ecdh.PublicKey, privSender *ecdh.PrivateKey) ([]byte, []byte, error) {
	privEph, err := dh.dh.GenerateKey(dh.rand)
	if err != nil {
		return nil, nil, err
	}
	pubEph := privEph.PublicKey()
	dhA, err := privEph.ECDH(pubRecipient)
	if err != nil {
		return nil, nil, err
	}
	dhB, err := privSender.ECDH(pubEph)
	if err != nil {
		return nil, nil, err
	}
	dhVal := append(dhA, dhB...)
	encPubEph := pubEph.Bytes()

	encPubRecip := pubRecipient.Bytes()
	encPubSender := privSender.PublicKey().Bytes()
	kemContext := append(encPubEph, encPubRecip...)
	kemContext = append(kemContext, encPubSender...)

	sharedSecret := dh.ExtractAndExpand(dhVal, kemContext)
	return sharedSecret, encPubEph, nil
}

func (dh *dhKEM) AuthDecap(enc []byte, privRecipient *ecdh.PrivateKey, pubSender *ecdh.PublicKey) ([]byte, error) {
	pubEph, err := dh.dh.NewPublicKey(enc)
	if err != nil {
		return nil, err
	}
	dhA, err := privRecipient.ECDH(pubEph)
	if err != nil {
		return nil, err
	}
	dhB, err := privRecipient.ECDH(pubSender)
	if err != nil {
		return nil, err
	}
	dhVal := append(dhA, dhB...)

	encPubRecip := privRecipient.PublicKey().Bytes()
	encPubSender := pubSender.Bytes()
	kemContext := append(enc, encPubRecip...)
	kemContext = append(kemContext, encPubSender...)

	return dh.ExtractAndExpand(dhVal, kemContext), nil
}

type Sender struct {
	aead cipher.AEAD
	kem  *dhKEM

	sharedSecret []byte

	suiteID []byte

	key            []byte
	baseNonce      []byte
	exporterSecret []byte

	seqNum uint128
}

var SupportedAEADs = map[uint16]struct {
	keySize   int
	nonceSize int
	aead      func([]byte) (cipher.AEAD, error)
}{
	// RFC 9180 Section 7.3
	0x0001: {keySize: 16, nonceSize: 12, aead: func(key []byte) (cipher.AEAD, error) {
		block, err := aes.NewCipher(key)
		if err != nil {
			return nil, err
		}
		return cipher.NewGCM(block)
	}},
	0x0002: {keySize: 32, nonceSize: 12, aead: func(key []byte) (cipher.AEAD, error) {
		block, err := aes.NewCipher(key)
		if err != nil {
			return nil, err
		}
		return cipher.NewGCM(block)
	}},
	0x0003: {keySize: 32, nonceSize: 12, aead: chacha20poly1305.New},
}

var SupportedKDFs = map[uint16]func() *hkdfKDF{
	// RFC 9180 Section 7.2
	0x0001: func() *hkdfKDF { return &hkdfKDF{crypto.SHA256} },
	0x0002: func() *hkdfKDF { return &hkdfKDF{crypto.SHA384} },
	0x0003: func() *hkdfKDF { return &hkdfKDF{crypto.SHA512} },
}

func SetupSender(rand io.Reader, kemID, kdfID, aeadID uint16, pub crypto.PublicKey, info []byte) ([]byte, *Sender, error) {
	suiteID := SuiteID(kemID, kdfID, aeadID)

	kem, err := newDHKem(rand, kemID)
	if err != nil {
		return nil, nil, err
	}
	sharedSecret, encapsulatedKey, err := kem.Encap(pub.(*ecdh.PublicKey))
	if err != nil {
		return nil, nil, err
	}

	kdfInit, ok := SupportedKDFs[kdfID]
	if !ok {
		return nil, nil, errors.New("unknown KDF id")
	}
	kdf := kdfInit()

	aeadInfo, ok := SupportedAEADs[aeadID]
	if !ok {
		return nil, nil, errors.New("unknown AEAD id")
	}

	pskIDHash := kdf.LabeledExtract(suiteID, []byte(""), []byte("psk_id_hash"), []byte(""))
	infoHash := kdf.LabeledExtract(suiteID, []byte(""), []byte("info_hash"), info)
	ksContext := append([]byte{0}, pskIDHash...)
	ksContext = append(ksContext, infoHash...)

	secret := kdf.LabeledExtract(suiteID, sharedSecret, []byte("secret"), []byte(""))

	key := kdf.LabeledExpand(suiteID, secret, []byte("key"), ksContext, uint16(aeadInfo.keySize) /* Nk - key size for AEAD */)
	baseNonce := kdf.LabeledExpand(suiteID, secret, []byte("base_nonce"), ksContext, uint16(aeadInfo.nonceSize) /* Nn - nonce size for AEAD */)
	exporterSecret := kdf.LabeledExpand(suiteID, secret, []byte("exp"), ksContext, uint16(kdf.hash.Size()) /* Nh - hash output size of the kdf*/)

	aead, err := aeadInfo.aead(key)
	if err != nil {
		return nil, nil, err
	}

	return encapsulatedKey, &Sender{
		kem:            kem,
		aead:           aead,
		sharedSecret:   sharedSecret,
		suiteID:        suiteID,
		key:            key,
		baseNonce:      baseNonce,
		exporterSecret: exporterSecret,
	}, nil
}

func (s *Sender) computeNonce() []byte {
	nonce := s.seqNum.bytes()[16-s.aead.NonceSize():]
	for i := range s.baseNonce {
		nonce[i] ^= s.baseNonce[i]
	}
	return nonce
}

func (s *Sender) incrementSeqNum() error {
	if s.seqNum.bitLen() >= (s.aead.NonceSize()*8)-1 {
		return errors.New("message limit reached")
	}
	s.seqNum = s.seqNum.addOne()
	return nil
}

func (s *Sender) Seal(aad, plaintext []byte) ([]byte, error) {
	// fmt.Print("seq num is", s.seqNum)
	ciphertext := s.aead.Seal(nil, s.computeNonce(), plaintext, aad)
	if err := s.incrementSeqNum(); err != nil {
		return nil, err
	}
	return ciphertext, nil
}

func (s *Sender) Open(aad, ciphertext []byte) ([]byte, error) {
	plaintext, err := s.aead.Open(nil, s.computeNonce(), ciphertext, aad)
	if err != nil {
		return nil, err
	}
	if err := s.incrementSeqNum(); err != nil {
		return nil, err
	}
	return plaintext, nil
}

func SuiteID(kemID, kdfID, aeadID uint16) []byte {
	suiteID := make([]byte, 0, 4+2+2+2)
	suiteID = append(suiteID, []byte("HPKE")...)
	suiteID = binary.BigEndian.AppendUint16(suiteID, kemID)
	suiteID = binary.BigEndian.AppendUint16(suiteID, kdfID)
	suiteID = binary.BigEndian.AppendUint16(suiteID, aeadID)
	return suiteID
}

func ParseHPKEPublicKey(kemID uint16, bytes []byte) (*ecdh.PublicKey, error) {
	kemInfo, ok := SupportedKEMs[kemID]
	if !ok {
		return nil, errors.New("unknown KEM id")
	}
	return kemInfo.curve.NewPublicKey(bytes)
}

type uint128 struct {
	hi, lo uint64
}

func (u uint128) addOne() uint128 {
	lo, carry := bits.Add64(u.lo, 1, 0)
	return uint128{u.hi + carry, lo}
}

func (u uint128) bitLen() int {
	return bits.Len64(u.hi) + bits.Len64(u.lo)
}

func (u uint128) bytes() []byte {
	b := make([]byte, 16)
	binary.BigEndian.PutUint64(b[0:], u.hi)
	binary.BigEndian.PutUint64(b[8:], u.lo)
	return b
}
