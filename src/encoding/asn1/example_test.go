package asn1_test

import (
	"encoding/asn1"
	"fmt"
)

func ExampleEncodingDecoding() {
	type Road struct {
		Number int
		Name   string
	}

	roads := Road{29, "Diamond Fork"}

	buffer, err := asn1.Marshal(roads)
	if err != nil {
		panic(err.Error())
	}

	var road Road
	_, err1 := asn1.Unmarshal(buffer, &road)
	if err1 != nil {
		panic(err1.Error())
	}

	fmt.Println(road.Number)
	fmt.Println(road.Name)
	//Output:
	//29
	//Diamond Fork
}
