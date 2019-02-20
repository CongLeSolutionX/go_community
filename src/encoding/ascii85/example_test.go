// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package ascii85_test

import (
	"bytes"
	"encoding/ascii85"
	"fmt"
)

func ExampleEncode() {
	// plain message
	src := []byte("Man is distinguished, not only by his reason, but by this singular passion from " +
		"other animals, which is a lust of the mind, that by a perseverance of delight in " +
		"the continued and indefatigable generation of knowledge, exceeds the short " +
		"vehemence of any carnal pleasure.")
	buffer := make([]byte, ascii85.MaxEncodedLen(len(src)))
	ascii85.Encode(buffer, src)

	fmt.Println(string(buffer))
	// Output:
	// 9jqo^BlbD-BleB1DJ+*+F(f,q/0JhKF<GL>Cj@.4Gp$d7F!,L7@<6@)/0JDEF<G%<+EV:2F!,O<DJ+*.@<*K0@<6L(Df-\0Ec5e;DffZ(EZee.Bl.9pF"AGXBPCsi+DGm>@3BB/F*&OCAfu2/AKYi(DIb:@FD,*)+C]U=@3BN#EcYf8ATD3s@q?d$AftVqCh[NqF<G:8+EV:.+Cf>-FD5W8ARlolDIal(DId<j@<?3r@:F%a+D58'ATD4$Bl@l3De:,-DJs`8ARoFb/0JMK@qB4^F!,R<AKZ&-DfTqBG%G>uD.RTpAKYo'+CT/5+Cei#DII?(E,9)oF*2M7/cYkO

}

func ExampleDecode() {
	// encoded message
	src := "9jqo^BlbD-BleB1DJ+*+F(f,q/0JhKF<GL>Cj@.4Gp$d7F!,L7@<6@)/0JDEF<G%<+EV:2F!,O<DJ+*.@<*K0@<6L(Df-\\0Ec5e;DffZ(EZee.Bl.9pF\"AGXBPCsi+DGm>@3BB/F*&OCAfu2/AKYi(DIb:@FD,*)+C]U=@3BN#EcYf8ATD3s@q?d$AftVqCh[NqF<G:8+EV:.+Cf>-FD5W8ARlolDIal(DId<j@<?3r@:F%a+D58'ATD4$Bl@l3De:,-DJs`8ARoFb/0JMK@qB4^F!,R<AKZ&-DfTqBG%G>uD.RTpAKYo'+CT/5+Cei#DII?(E,9)oF*2M7/cYkO"

	buffer := make([]byte, len(src))
	_, _, err := ascii85.Decode(buffer, []byte(src), true)

	if err != nil {
		fmt.Println(err)
	}
	// resize bytes buffer
	b := bytes.Trim(buffer, "\x00")
	fmt.Println(string(b))

	// Output:
	// Man is distinguished, not only by his reason, but by this singular passion from other animals, which is a lust of the mind, that by a perseverance of delight in the continued and indefatigable generation of knowledge, exceeds the short vehemence of any carnal pleasure.
}
