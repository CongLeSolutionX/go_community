//go:build !arm64

package secret

const ditSupported = false

func enableDIT() bool  { return false }
func ditEnabled() bool { return false }
func disableDIT()      {}
