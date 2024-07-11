package secret

// TODO: this should be initialized by looking at FEAT_DIT to check,
// since some arm64 platforms may not actually support the bit.
const ditSupported = true

func enableDIT() bool
func ditEnabled() bool
func disableDIT()
