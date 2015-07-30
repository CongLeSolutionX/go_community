// import all internal packages into runtime so they are linked in.
package runtime

import (
	_ "runtime/internal/base"
	_ "runtime/internal/gc"
	_ "runtime/internal/iface"
	_ "runtime/internal/lock"
	_ "runtime/internal/print"
	_ "runtime/internal/race"
	_ "runtime/internal/writebarrier"
)
