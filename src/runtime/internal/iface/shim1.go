package iface

import (
	_base "runtime/internal/base"
	"unsafe"
)

func assertE2I(inter *Interfacetype, e interface{}, r *FInterface) {
	AssertE2I(inter, e, r)
}

func assertE2I2(inter *Interfacetype, e interface{}, r *FInterface) bool {
	return AssertE2I2(inter, e, r)
}

// implementation of new builtin
func newobject(typ *_base.Type) unsafe.Pointer {
	return Newobject(typ)
}

// NOTE: Really dst *unsafe.Pointer, src unsafe.Pointer,
// but if we do that, Go inserts a write barrier on *dst = src.
//go:nosplit
func writebarrierptr(dst *uintptr, src uintptr) {
	Writebarrierptr(dst, src)
}

// typedmemmove copies a value of type t to dst from src.
//go:nosplit
func typedmemmove(typ *_base.Type, dst, src unsafe.Pointer) {
	Typedmemmove(typ, dst, src)
}
