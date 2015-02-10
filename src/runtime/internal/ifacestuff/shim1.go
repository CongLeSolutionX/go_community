package ifacestuff

import _core "runtime/internal/core"

func assertE2I(inter *_core.Interfacetype, e interface{}, r *FInterface) {
	AssertE2I(inter, e, r)
}

func assertE2I2(inter *_core.Interfacetype, e interface{}, r *FInterface) bool {
	return AssertE2I2(inter, e, r)
}
