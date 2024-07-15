package runtime

func WithDataIndependentTiming(f func()) {
	if !ditSupported {
		f()
		return
	}

	LockOSThread()
	defer UnlockOSThread()

	alreadyEnabled := enableDIT()
	f()
	if !alreadyEnabled {
		disableDIT()
	}
}

func DataIndependentTimingEnabled() bool {
	return ditEnabled()
}