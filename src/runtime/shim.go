package runtime

import _ "unsafe"
import "runtime/internal/base"

import "runtime/internal/iface"

//go:linkname runtime/internal/base.MemProfileRate MemProfileRate
var MemProfileRate int

const GOOS = base.GOOS

const GOARCH = base.GOARCH

// RuntimeError runtime/internal/base RuntimeError method
//func (runtime.errorString).RuntimeError()

// Error runtime/internal/base Error method
//func (runtime.errorString).Error() string

//go:linkname runtime/internal/gc.Gosched Gosched
func Gosched()

// Add a dummy struct decl
type TypeAssertionError struct {
	InterfaceString string
	ConcreteString  string
	AssertedString  string
	MissingMethod   string
}

func GC() { iface.GC() }

// Error runtime/internal/iface Error method
//func (*runtime.TypeAssertionError).Error() string

// RuntimeError runtime/internal/iface RuntimeError method
//func (*runtime.TypeAssertionError).RuntimeError()
