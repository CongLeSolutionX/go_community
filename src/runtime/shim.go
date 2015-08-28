package runtime

import _ "unsafe"
import "runtime/internal/base"

import "runtime/internal/iface"

//go:linkname runtime/internal/base.MemProfileRate MemProfileRate
var MemProfileRate int

// RuntimeError runtime/internal/base RuntimeError method
//func (runtime.errorString).RuntimeError()

const GOARCH = base.GOARCH

const GOOS = base.GOOS

// Error runtime/internal/base Error method
//func (runtime.errorString).Error() string

//go:linkname runtime/internal/gc.Gosched Gosched
func Gosched()

// Error runtime/internal/iface Error method
//func (*runtime.TypeAssertionError).Error() string

// RuntimeError runtime/internal/iface RuntimeError method
//func (*runtime.TypeAssertionError).RuntimeError()

func GC() { iface.GC() }

// Add a dummy struct decl
type TypeAssertionError struct {
	InterfaceString string
	ConcreteString  string
	AssertedString  string
	MissingMethod   string
}
