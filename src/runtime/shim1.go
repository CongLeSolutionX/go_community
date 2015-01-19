// entry point for c <- x from compiled code

package runtime

import "unsafe"
import "runtime/internal/lock"
import "runtime/internal/channels"
import _ "runtime/internal/check" // link in check.check
import _ "runtime/internal/defers"
import _ "runtime/internal/sync" // link in sync.syncsemcheck
import "runtime/internal/seq"

//go:nosplit
func chansend1(t *channels.Chantype, c *channels.Hchan, elem unsafe.Pointer) {
	channels.Chansend(t, c, elem, true, lock.Getcallerpc(unsafe.Pointer(&t)))
}

func newselect(sel *channels.Select, selsize int64, size int32) {
	channels.Newselect(sel, selsize, size)
}

func intstring(v int64) string {
	return seq.Intstring(v)
}
