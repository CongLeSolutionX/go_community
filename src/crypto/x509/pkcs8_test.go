// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package x509

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/hex"
	"reflect"
	"testing"
)

// Generated using:
//   openssl genrsa 1024 | openssl pkcs8 -topk8 -nocrypt
var pkcs8RSAPrivateKeyHex = `30820278020100300d06092a864886f70d0101010500048202623082025e02010002818100cfb1b5bf9685ffa97b4f99df4ff122b70e59ac9b992f3bc2b3dde17d53c1a34928719b02e8fd17839499bfbd515bd6ef99c7a1c47a239718fe36bfd824c0d96060084b5f67f0273443007a24dfaf5634f7772c9346e10eb294c2306671a5a5e719ae24b4de467291bc571014b0e02dec04534d66a9bb171d644b66b091780e8d020301000102818100b595778383c4afdbab95d2bfed12b3f93bb0a73a7ad952f44d7185fd9ec6c34de8f03a48770f2009c8580bcd275e9632714e9a5e3f32f29dc55474b2329ff0ebc08b3ffcb35bc96e6516b483df80a4a59cceb71918cbabf91564e64a39d7e35dce21cb3031824fdbc845dba6458852ec16af5dddf51a8397a8797ae0337b1439024100ea0eb1b914158c70db39031dd8904d6f18f408c85fbbc592d7d20dee7986969efbda081fdf8bc40e1b1336d6b638110c836bfdc3f314560d2e49cd4fbde1e20b024100e32a4e793b574c9c4a94c8803db5152141e72d03de64e54ef2c8ed104988ca780cd11397bc359630d01b97ebd87067c5451ba777cf045ca23f5912f1031308c702406dfcdbbd5a57c9f85abc4edf9e9e29153507b07ce0a7ef6f52e60dcfebe1b8341babd8b789a837485da6c8d55b29bbb142ace3c24a1f5b54b454d01b51e2ad03024100bd6a2b60dee01e1b3bfcef6a2f09ed027c273cdbbaf6ba55a80f6dcc64e4509ee560f84b4f3e076bd03b11e42fe71a3fdd2dffe7e0902c8584f8cad877cdc945024100aa512fa4ada69881f1d8bb8ad6614f192b83200aef5edf4811313d5ef30a86cbd0a90f7b025c71ea06ec6b34db6306c86b1040670fd8654ad7291d066d06d031`

// Generated using:
//   openssl ecparam -genkey -name secp224r1 | openssl pkcs8 -topk8 -nocrypt
var pkcs8EC224PrivateKeyHex = `3078020100301006072a8648ce3d020106052b810400210461305f020101041cca3d72b3e88fed2684576dad9b80a9180363a5424986900e3abcab3fa13c033a0004f8f2a6372872a4e61263ed893afb919576a4cacfecd6c081a2cbc76873cf4ba8530703c6042b3a00e2205087e87d2435d2e339e25702fae1`

// Generated using:
//   openssl ecparam -genkey -name secp256r1 | openssl pkcs8 -topk8 -nocrypt
var pkcs8EC256PrivateKeyHex = `308187020100301306072a8648ce3d020106082a8648ce3d030107046d306b0201010420dad6b2f49ca774c36d8ae9517e935226f667c929498f0343d2424d0b9b591b43a14403420004b9c9b90095476afe7b860d8bd43568cab7bcb2eed7b8bf2fa0ce1762dd20b04193f859d2d782b1e4cbfd48492f1f533113a6804903f292258513837f07fda735`

// Generated using:
//   openssl ecparam -genkey -name secp384r1 | openssl pkcs8 -topk8 -nocrypt
var pkcs8EC384PrivateKeyHex = `3081b6020100301006072a8648ce3d020106052b8104002204819e30819b02010104309bf832f6aaaeacb78ce47ffb15e6fd0fd48683ae79df6eca39bfb8e33829ac94aa29d08911568684c2264a08a4ceb679a164036200049070ad4ed993c7770d700e9f6dc2baa83f63dd165b5507f98e8ff29b5d2e78ccbe05c8ddc955dbf0f7497e8222cfa49314fe4e269459f8e880147f70d785e530f2939e4bf9f838325bb1a80ad4cf59272ae0e5efe9a9dc33d874492596304bd3`

// Generated using:
//   openssl ecparam -genkey -name secp521r1 | openssl pkcs8 -topk8 -nocrypt
var pkcs8EC521PrivateKeyHex = `3081ed020100301006072a8648ce3d020106052b810400230481d53081d20201010441850d81618c5da1aec74c2eed608ba816038506975e6427237c2def150c96a3b13efbfa1f89f1be15cdf4d0ac26422e680e65a0ddd4ad3541ad76165fbf54d6e34ba18189038186000400da97bcedba1eb6d30aeb93c9f9a1454598fa47278df27d6f60ea73eb672d8dc528a9b67885b5b5dcef93c9824f7449ab512ee6a27e76142f56b94b474cfd697e810046c8ca70419365245c1d7d44d0db82c334073835d002232714548abbae6e5700f5ef315ee08b929d8581383dcf2d1c98c2f8a9fccbf79c9579f7b2fd8a90115ac2`

func TestPKCS8(t *testing.T) {
	tests := []struct {
		name    string
		keyHex  string
		keyType reflect.Type
		curve   elliptic.Curve
	}{
		{
			name:    "RSA private key",
			keyHex:  pkcs8RSAPrivateKeyHex,
			keyType: reflect.TypeOf(&rsa.PrivateKey{}),
		},
		{
			name:    "EC 224 private key",
			keyHex:  pkcs8EC224PrivateKeyHex,
			keyType: reflect.TypeOf(&ecdsa.PrivateKey{}),
			curve:   elliptic.P224(),
		},
		{
			name:    "EC 256 private key",
			keyHex:  pkcs8EC256PrivateKeyHex,
			keyType: reflect.TypeOf(&ecdsa.PrivateKey{}),
			curve:   elliptic.P256(),
		},
		{
			name:    "EC 384 private key",
			keyHex:  pkcs8EC384PrivateKeyHex,
			keyType: reflect.TypeOf(&ecdsa.PrivateKey{}),
			curve:   elliptic.P384(),
		},
		{
			name:    "EC 521 private key",
			keyHex:  pkcs8EC521PrivateKeyHex,
			keyType: reflect.TypeOf(&ecdsa.PrivateKey{}),
			curve:   elliptic.P521(),
		},
	}

	for _, test := range tests {
		derBytes, err := hex.DecodeString(test.keyHex)
		if err != nil {
			t.Errorf("failed to decode hex with %s: %s", test.name, err)
			continue
		}
		pk, err := ParsePKCS8PrivateKey(derBytes)
		if err != nil {
			t.Errorf("failed to decode PKCS8 with %s: %s", test.name, err)
			continue
		}
		if reflect.TypeOf(pk) != test.keyType {
			t.Errorf("decoded PKCS8 returned unexpected key type with %s: %T", test.name, pk)
			continue
		}
		ecKey, isEC := pk.(*ecdsa.PrivateKey)
		if isEC && ecKey.Curve != test.curve {
			t.Errorf("decoded PKCS8 returned unexpected curve with %s", test.name)
			continue
		}
		newBytes, err := MarshalPKCS8PrivateKey(pk)
		if err != nil {
			t.Errorf("failed to marshal into PKCS8 with %s: %s", test.name, err)
			continue
		}
		// TODO: When asn1 marshalling the go-marshalled EC keys are different than
		// the openssl-marshalled ones, so test the keys instead
		if isEC {
			pk, _ := ParsePKCS8PrivateKey(newBytes)
			if newECKey, ok := pk.(*ecdsa.PrivateKey); !ok {
				t.Errorf("failed to parse PKCS8 with %s: %T", test.name, pk)
			} else if !reflect.DeepEqual(ecKey, newECKey) {
				t.Errorf("parsed PKCS8 didn't match source key with %s", test.name)
			}
		} else if !bytes.Equal(derBytes, newBytes) {
			t.Errorf("marshalled PKCS8 didn't match source with %s", test.name)
		}
	}
}
