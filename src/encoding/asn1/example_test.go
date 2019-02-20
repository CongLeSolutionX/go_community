// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package asn1_test

import (
	"crypto/rsa"
	"encoding/asn1"
	"encoding/base64"
	"fmt"
)

func ExampleMarshal() {
	type Road struct {
		Number int
		Name   string
	}

	roads := Road{29, "Diamond Fork"}

	buffer, err := asn1.Marshal(roads)
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("%s", buffer)
	// Output:
	// 0Diamond Fork
}

func ExampleUnmarshal() {
	type Road struct {
		Number int
		Name   string
	}
	// encoded value
	buffer := []byte{48, 17, 2, 1, 29, 19, 12, 68, 105, 97, 109, 111, 110, 100, 32, 70, 111, 114, 107}
	var road Road
	_, err1 := asn1.Unmarshal(buffer, &road)
	if err1 != nil {
		panic(err1.Error())
	}
	fmt.Println(road.Number)
	fmt.Println(road.Name)
	// Output:
	// 29
	// Diamond Fork
}

func ExampleUnmarshal_pem() {

	publicKeyBase64 := "MIIBCgKCAQEA+xGZ/wcz9ugFpP07Nspo6U17l0YhFiFpxxU4pTk3Lifz9R3zsIsuERwta7+fWIfxOo208ett/jhskiVodSEt3QBGh4XBipyWopKwZ93HHaDVZAALi/2A+xTBtWdEo7XGUujKDvC2/aZKukfjpOiUI8AhLAfjmlcD/UZ1QPh0mHsglRNCmpCwmwSXA9VNmhz+PiB+Dml4WWnKW/VHo2ujTXxq7+efMU4H2fny3Se3KYOsFPFGZ1TNQSYlFuShWrHPtiLmUdPoP6CV2mML1tk+l7DIIqXrQhLUKDACeM5roMx0kLhUWB8P+0uj1CNlNN4JRZlC7xFfqiMbFRU9Z4N6YwIDAQAB"
	// Base64 decode.
	publicKeyBinary, err := base64.StdEncoding.DecodeString(publicKeyBase64)
	if err != nil {
		panic(err)
	}

	var pubKey rsa.PublicKey
	if _, err := asn1.Unmarshal(publicKeyBinary, &pubKey); err != nil {
		panic(err.Error())
	}

	fmt.Printf("%+v", pubKey)
	// Output:
	// {N:+31694494193131919450762706770809468840354434666672424008781573515431625746096585280368250621252631794631453815565120424334098050529875071792744935607775591594371021649824660139042437373082830123871741534477779584668021334953680599012465401947458226088370461865552891478585096543406113181950641833873129358146677978145875731988172846697966127755347490321808734930449792906958988143870053813662496372475791076907799415699868494450132314401903370388663057758988383314607252089667633950836452611098799199281145096780828073017774267281929381261415199175599661354830101985696624851282605981615581853627961488141725060004451 E:65537}
}
