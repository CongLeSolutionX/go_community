// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package json_test

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

type Animal int

const (
	Unknown Animal = iota
	Gopher
	Zebra
)

func (a *Animal) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	switch strings.ToLower(s) {
	case "gopher":
		*a = Gopher
	case "zebra":
		*a = Zebra
	default:
		*a = Unknown
	}

	return nil
}

func (a Animal) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

func (a Animal) String() string {
	switch a {
	default:
		return "unknown"
	case Gopher:
		return "gopher"
	case Zebra:
		return "zebra"
	}
}

func Example_marshalJSON() {
	zoo := []Animal{Gopher, Zebra, Unknown, Gopher, Gopher, Zebra}
	marshaled, err := json.Marshal(&zoo)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s", marshaled)

	// Output:
	// ["gopher","zebra","unknown","gopher","gopher","zebra"]
}

func Example_unmarshalJSON() {
	rawZooManifest := `["zebra", "gopher", "zebra", "unknown", "gopher", "gopher", "undiscovered"]`
	var zooManifest []*Animal
	if err := json.Unmarshal([]byte(rawZooManifest), &zooManifest); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%v", zooManifest)

	// Output:
	// [zebra gopher zebra unknown gopher gopher unknown]
}
