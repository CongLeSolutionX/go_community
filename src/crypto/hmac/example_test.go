package hmac_test

import (
	"crypto/hmac"
	"crypto/sha1"
	"fmt"
)

func ExampleSum() {
	message := []byte("This page intentionally left blank.")
	key := []byte("secret")

	h := hmac.New(sha1.New, key)
	h.Write(message)

	fmt.Printf("%x", h.Sum(nil))
	// Output: 2916e4de8048ace53e53596204f9eb364297904d
}

func ExampleEqual() {
	message := []byte("This page intentionally left blank.")
	key := []byte("secret")
	result := []byte{0x29, 0x16, 0xe4, 0xde, 0x80, 0x48, 0xac, 0xe5, 0x3e, 0x53, 0x59, 0x62, 0x04, 0xf9, 0xeb, 0x36, 0x42, 0x97, 0x90, 0x4d}

	h := hmac.New(sha1.New, key)
	h.Write(message)

	fmt.Printf("%v", hmac.Equal(result, h.Sum(nil)))
	// Output: true
}
