package secret

import "internal/cpu"

var ditSupported = cpu.ARM64.HasDIT

func enableDIT() bool
func ditEnabled() bool
func disableDIT()
