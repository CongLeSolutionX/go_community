package finalize

import _ "unsafe"

// TODO(matloob): This should be trivial to generate in the
// shim generation stage.

//go:linkname SetFinalizer runtime.SetFinalizer
