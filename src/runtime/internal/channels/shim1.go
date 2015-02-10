package channels

import (
	_core "runtime/internal/core"
	"unsafe"
)

// NOTE: Really dst *unsafe.Pointer, src unsafe.Pointer,
// but if we do that, Go inserts a write barrier on *dst = src.
//go:nosplit
func writebarrierptr(dst *uintptr, src uintptr) {
	Writebarrierptr(dst, src)
}

// typedmemmove copies a value of type t to dst from src.
//go:nosplit
func typedmemmove(typ *_core.Type, dst, src unsafe.Pointer) {
	Typedmemmove(typ, dst, src)
}

func newselect(sel *Select, selsize int64, size int32) {
	Newselect(sel, selsize, size)
}
