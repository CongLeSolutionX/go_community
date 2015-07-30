package runtime

import _ "unsafe"
import "runtime/internal/base"

//go:linkname MemProfileRate runtime/internal/base.MemProfileRate
var MemProfileRate int

const GOOS = base.GOOS

const GOARCH = base.GOARCH

//go:linkname runtime/internal/gc.Gosched Gosched
func Gosched()

//go:linkname runtime/internal/iface.Error Error
func (x *TypeAssertionError) Error() string

// Add a dummy struct decl
type TypeAssertionError struct {
	InterfaceString string
	ConcreteString  string
	AssertedString  string
	MissingMethod   string
}

//go:linkname runtime/internal/iface.RuntimeError RuntimeError
func (x *TypeAssertionError) RuntimeError()
