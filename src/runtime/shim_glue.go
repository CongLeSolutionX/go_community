// import all internal packages into runtime so they are linked in.
package runtime

import (
	_ "runtime/internal/cgo"
	_ "runtime/internal/channels"
	_ "runtime/internal/check"
	_ "runtime/internal/core"
	_ "runtime/internal/defers"
	_ "runtime/internal/finalize"
	_ "runtime/internal/fp"
	_ "runtime/internal/gc"
	_ "runtime/internal/hash"
	_ "runtime/internal/heapdump"
	_ "runtime/internal/ifacestuff"
	_ "runtime/internal/lock"
	_ "runtime/internal/maps"
	_ "runtime/internal/netpoll"
	_ "runtime/internal/printf"
	_ "runtime/internal/prof"
	_ "runtime/internal/sched"
	_ "runtime/internal/schedinit"
	_ "runtime/internal/sem"
	_ "runtime/internal/seq"
	_ "runtime/internal/stackwb"
	_ "runtime/internal/strings"
	_ "runtime/internal/sync"
	_ "runtime/internal/vdso"
)
