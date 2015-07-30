// +build ignore
package PACKAGE

func newobject(typ *byte) *any
func convI2E(elem any) (ret any)
func convI2I(typ *byte, elem any) (ret any)
func convT2E(typ *byte, elem, buf *any) (ret any)
func convT2I(typ *byte, typ2 *byte, cache **byte, elem, buf *any) (ret any)
func assertE2E(typ *byte, iface any, ret *any)
func assertE2E2(typ *byte, iface any, ret *any) bool
func assertE2I(typ *byte, iface any, ret *any)
func assertE2I2(typ *byte, iface any, ret *any) bool
func assertE2T(typ *byte, iface any, ret *any)
func assertE2T2(typ *byte, iface any, ret *any) bool
func assertI2E(typ *byte, iface any, ret *any)
func assertI2E2(typ *byte, iface any, ret *any) bool
func assertI2I(typ *byte, iface any, ret *any)
func assertI2I2(typ *byte, iface any, ret *any) bool
func assertI2T(typ *byte, iface any, ret *any)
func assertI2T2(typ *byte, iface any, ret *any) bool
func writebarrierptr(dst *any, src any)
func typedmemmove(typ *byte, dst *any, src *any)
