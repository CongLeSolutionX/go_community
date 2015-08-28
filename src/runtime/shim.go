package runtime

import _ "unsafe"
import "runtime/internal/base"

import "runtime/internal/iface"

const GOARCH = base.GOARCH

//go:linkname runtime/internal/base.MemProfileRate MemProfileRate
var MemProfileRate int

const GOOS = base.GOOS

// Error runtime/internal/base Error method
//func (runtime.errorString).Error() string

// RuntimeError runtime/internal/base RuntimeError method
//func (runtime.errorString).RuntimeError()

//go:linkname runtime/internal/gc.Gosched Gosched
func Gosched()

func GC() { iface.GC() }

// RuntimeError runtime/internal/iface RuntimeError method
//func (*runtime.TypeAssertionError).RuntimeError()

// Error runtime/internal/iface Error method
//func (*runtime.TypeAssertionError).Error() string

// Add a dummy struct decl
type TypeAssertionError struct {
	InterfaceString string
	ConcreteString  string
	AssertedString  string
	MissingMethod   string
}
