package maps

import (
	_core "runtime/internal/core"
	"unsafe"
)

func makemap(t *Maptype, hint int64) *Hmap {
	return Makemap(t, hint)
}

func mapaccess2(t *Maptype, h *Hmap, key unsafe.Pointer) (unsafe.Pointer, bool) {
	return Mapaccess2(t, h, key)
}

func mapassign1(t *Maptype, h *Hmap, key unsafe.Pointer, val unsafe.Pointer) {
	Mapassign1(t, h, key, val)
}

func mapdelete(t *Maptype, h *Hmap, key unsafe.Pointer) {
	Mapdelete(t, h, key)
}

func mapiterinit(t *Maptype, h *Hmap, it *Hiter) {
	Mapiterinit(t, h, it)
}

// implementation of new builtin
func newobject(typ *_core.Type) unsafe.Pointer {
	return Newobject(typ)
}
