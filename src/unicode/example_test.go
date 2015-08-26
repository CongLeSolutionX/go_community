// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package unicode_test

import (
	"fmt"
	"unicode"
)

// func ExampleIn(){}
// func ExampleIs(){}

func ExampleIsControl() {
	const backspace = '\u0008'
	if unicode.IsControl(backspace) {
		fmt.Println(`Found control rune`)
	}

	// Output:
	// Found control rune
}

func ExampleIsgraphic() {
	const digit = '1'
	if unicode.IsDigit(digit) {
		fmt.Println(`Found digit rune`)
	}

	// Output:
	// Found digit rune
}

func ExampleIsGraphic() {
	const graphic = '\u1F4A'
	if unicode.IsGraphic(graphic) {
		fmt.Println(`Found graphic rune`)
	}

	// Output:
	// Found graphic rune
}

func ExampleIsLetter() {
	const letter = 'l'
	if unicode.IsLetter(letter) {
		fmt.Println(`Found letter rune`)
	}

	// Output:
	// Found letter rune
}

func ExampleIsLower() {
	const lcase = 'l'
	if unicode.IsLower(lcase) {
		fmt.Println(`Found lower case letter rune`)
	}

	// Output:
	// Found lower case letter rune
}

func ExampleIsMark() {
	const mark = '\u0300'
	if unicode.IsMark(mark) {
		fmt.Println(`Found mark rune`)
	}

	// Output:
	// Found mark rune
}

func ExampleIsNumber() {
	const number = '\u0039'
	if unicode.IsNumber(number) {
		fmt.Println(`Found number rune`)
	}

	// Output:
	// Found number rune
}
