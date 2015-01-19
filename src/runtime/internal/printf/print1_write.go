// +build !android

package printf

import (
	_core "runtime/internal/core"
	"unsafe"
)

func writeErr(b []byte) {
	_core.Write(2, unsafe.Pointer(&b[0]), int32(len(b)))
}
