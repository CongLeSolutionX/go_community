package main

import (
	"linkname"
)

func main() {
	s, i := linkname.Test("exe4")
	if s != "runtime error: exe4" || i <= 0 {
		panic("unexpected failure after linking successfully")
	}
}
