package hash_test

import (
	"crypto/sha256"
	"encoding"
	"fmt"
	"log"
)

func Example_binaryMarshaler() {
	const (
		input1 = "The tunneling gopher digs downwards, "
		input2 = "unaware of what he will find."
	)

	first := sha256.New()
	first.Write([]byte(input1))

	marshaler, ok := first.(encoding.BinaryMarshaler)
	if !ok {
		log.Fatal("first does not implement encoding.BinaryMarshaler")
	}
	state, err := marshaler.MarshalBinary()
	if err != nil {
		log.Fatal("unable to marshal hash:", err)
	}

	second := sha256.New()

	unmarshaler, ok := second.(encoding.BinaryUnmarshaler)
	if !ok {
		log.Fatal("second does not implement encoding.BinaryUnmarshaler")
	}
	if err := unmarshaler.UnmarshalBinary(state); err != nil {
		log.Fatal("unable to unmarshal hash:", err)
	}

	first.Write([]byte(input2))
	second.Write([]byte(input2))

	fmt.Printf("First sum:  %x\n", first.Sum(nil))
	fmt.Printf("Second sum: %x\n", second.Sum(nil))
	// Output:
	// First sum:  57d51a066f3a39942649cd9a76c77e97ceab246756ff3888659e6aa5a07f4a52
	// Second sum: 57d51a066f3a39942649cd9a76c77e97ceab246756ff3888659e6aa5a07f4a52
}
