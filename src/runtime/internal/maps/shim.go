package maps

import "unsafe"
import "runtime/internal/core"

func newobject(typ *core.Type) unsafe.Pointer {
	return Newobject(typ)
}

func makemap(t *Maptype, hint int64) *Hmap {
	return Makemap(t, hint)
}

func mapdelete(t *Maptype, h *Hmap, key unsafe.Pointer) {
	Mapdelete(t, h, key)
}

func mapassign1(t *Maptype, h *Hmap, key unsafe.Pointer, val unsafe.Pointer) {
	Mapassign1(t, h, key, val)
}

func mapiterinit(t *Maptype, h *Hmap, it *Hiter) {
	Mapiterinit(t, h, it)
}

func mapaccess2(t *Maptype, h *Hmap, key unsafe.Pointer) (unsafe.Pointer, bool) {
	return Mapaccess2(t, h, key)
}
