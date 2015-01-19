package ifacestuff

import "runtime/internal/core"

func assertE2I(inter *core.Interfacetype, e interface{}) (r FInterface) {
	return AssertE2I(inter, e)
}

func assertE2I2(inter *core.Interfacetype, e interface{}) (r FInterface, ok bool) {
	return AssertE2I2(inter, e)
}
