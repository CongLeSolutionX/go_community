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
	const digit = '5'
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
	const letter = 'g'
	if unicode.IsLetter(letter) {
		fmt.Println(`Found letter rune`)
	}

	// Output:
	// Found letter rune
}

func ExampleIsLower() {
	const lcase = 'g'
	if unicode.IsLower(lcase) {
		fmt.Println(`Found lower case rune`)
	}

	// Output:
	// Found lower case rune
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

func ExampleIsPrint() {
	const printableRune = 'g'
	if unicode.IsPrint(printableRune) {
		fmt.Println(`'g' is printable rune`)
	}

	const notPrintableRune = '\u0008'
	if !unicode.IsPrint(notPrintableRune) {
		fmt.Println(`'\u0008' is not printable rune`)
	}

	// Output:
	// 'g' is printable rune
	// '\u0008' is not printable rune
}

func ExampleIsPunct() {
	const punct = '!'
	if unicode.IsPunct(punct) {
		fmt.Println(`Found punct rune`)
	}

	// Output:
	// Found punct rune
}

func ExampleIsSpace() {
	const space = ' '
	if unicode.IsSpace(space) {
		fmt.Println(`Found space rune`)
	}

	// Output:
	// Found space rune
}

func ExampleIsSymbol() {
	const symbol = '\u2103'
	if unicode.IsSymbol(symbol) {
		fmt.Println(`Found symbol rune`)
	}

	// Output:
	// Found symbol rune
}

func ExampleIsTitle() {
	const tc = '\u1fad'
	if unicode.IsTitle(tc) {
		fmt.Println(`Found title case rune`)
	}

	// Output:
	// Found title case rune
}

func ExampleIsUpper() {
	const ucase = 'G'
	if unicode.IsUpper(ucase) {
		fmt.Println(`Found upper case rune`)
	}

	// Output:
	// Found upper case rune
}

func ExampleSimpleFold() {
	fmt.Printf("%#U\n", unicode.SimpleFold('A'))      // 'a'
	fmt.Printf("%#U\n", unicode.SimpleFold('a'))      // 'A'
	fmt.Printf("%#U\n", unicode.SimpleFold('K'))      // 'k'
	fmt.Printf("%#U\n", unicode.SimpleFold('k'))      // '\u212A' (Kelvin symbol, K)
	fmt.Printf("%#U\n", unicode.SimpleFold('\u212A')) // 'K'
	fmt.Printf("%#U\n", unicode.SimpleFold('1'))      // '1'

	// Output:
	// U+0061 'a'
	// U+0041 'A'
	// U+006B 'k'
	// U+212A 'K'
	// U+004B 'K'
	// U+0031 '1'
}

func ExampleTo() {
	const lcG = 'g'
	fmt.Printf("%#U\n", unicode.To(unicode.UpperCase, lcG))
	fmt.Printf("%#U\n", unicode.To(unicode.LowerCase, lcG))
	fmt.Printf("%#U\n", unicode.To(unicode.TitleCase, lcG))

	const ucG = 'G'
	fmt.Printf("%#U\n", unicode.To(unicode.UpperCase, ucG))
	fmt.Printf("%#U\n", unicode.To(unicode.LowerCase, ucG))
	fmt.Printf("%#U\n", unicode.To(unicode.TitleCase, ucG))

	// Output:
	// U+0047 'G'
	// U+0067 'g'
	// U+0047 'G'
	// U+0047 'G'
	// U+0067 'g'
	// U+0047 'G'
}

func ExampleToLower() {
	const ucG = 'G'
	fmt.Printf("%#U\n", unicode.ToLower(ucG))

	// Output:
	// U+0067 'g'
}
func ExampleToTitle() {
	const ucG = 'g'
	fmt.Printf("%#U\n", unicode.ToTitle(ucG))

	// Output:
	// U+0047 'G'
}

func ExampleToUpper() {
	const ucG = 'g'
	fmt.Printf("%#U\n", unicode.ToUpper(ucG))

	// Output:
	// U+0047 'G'
}

func ExampleSpecialCase() {
	t := unicode.TurkishCase

	const lci = 'i'
	fmt.Printf("%#U\n", t.ToLower(lci))
	fmt.Printf("%#U\n", t.ToTitle(lci))
	fmt.Printf("%#U\n", t.ToUpper(lci))

	const uci = 'İ'
	fmt.Printf("%#U\n", t.ToLower(uci))
	fmt.Printf("%#U\n", t.ToTitle(uci))
	fmt.Printf("%#U\n", t.ToUpper(uci))

	// Output:
	// U+0069 'i'
	// U+0130 'İ'
	// U+0130 'İ'
	// U+0069 'i'
	// U+0130 'İ'
	// U+0130 'İ'
}
