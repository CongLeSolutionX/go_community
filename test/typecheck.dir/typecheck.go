// Does not compile.

package main

import "./foo"

func mine(int b) int {	// ERROR "undefined.*b"
	return b + 2	// ERROR "undefined.*b"
}

func main() {
	// Verify that the Go compiler will not
	// die after running into an undefined
	// type in the argument list for a
	// function.
	mine()		// GCCGO_ERROR "not enough arguments"
	c = mine()	// ERROR "undefined.*c|not enough arguments"

	_ = foo.Foo.Baz // ERROR "foo.Foo\.Baz undefined \(type foo.Foo has no method Baz\)"
	_ = foo.Foo.Bar // ERROR "foo.Foo\.Bar undefined \(type foo.Foo used as instance of foo.Foo\)"
	_ = foo.Foo2.Bar // ERROR "foo.Foo2\.Bar undefined \(type foo.Foo2 used as instance of foo.Foo2\)"
	_ = foo.Foo.baz // ERROR "foo.Foo\.baz undefined \(type foo.Foo used as instance of foo.Foo\)"
}
