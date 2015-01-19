package runtime

import "runtime/internal/schedinit"

func morestack() {
	schedinit.Morestack()
}
