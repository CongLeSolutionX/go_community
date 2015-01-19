package runtime

import (
	"runtime/internal/channels"
	_ "runtime/internal/check" // rt0
	"runtime/internal/core"
	_ "runtime/internal/defers"
	"runtime/internal/finalize"
	"runtime/internal/gc"
	"runtime/internal/ifacestuff"
	"runtime/internal/lock"
	"runtime/internal/printf"
	"runtime/internal/seq"
	_ "runtime/internal/sync"
	_ "runtime/internal/vdso"
	"unsafe"
)

// HACK(matloob) propagating GOOS/GOARCH
const GOOS string = lock.GOOS
const GOARCH string = lock.GOARCH

func SetFinalizer(obj interface{}, finalizer interface{}) {
	finalize.SetFinalizer(obj, finalizer)
}

func GC() {
	gc.GC()
}

func printstring(s string) {
	printf.Printstring(s)
}

func convT2I(t *core.Type, inter *core.Interfacetype, cache **core.Itab, elem unsafe.Pointer) (i ifacestuff.FInterface) {
	return ifacestuff.ConvT2I(t, inter, cache, elem)
}

// entry point for c <- x from compiled code
//go:nosplit
func chansend1(t *channels.Chantype, c *channels.Hchan, elem unsafe.Pointer) {
	channels.Chansend(t, c, elem, true, lock.Getcallerpc(unsafe.Pointer(&t)))
}

func intstring(v int64) string {
	return seq.Intstring(v)
}

func newselect(sel *channels.Select, selsize int64, size int32) {
	channels.Newselect(sel, selsize, size)
}
