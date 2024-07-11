package secret

import "runtime"

func WithDIT(f func() (any, error)) (any, error) {
	if !ditSupported {
		return f()
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	alreadyEnabled := enableDIT()
	res, err := f()
	if !alreadyEnabled {
		disableDIT()
	}

	return res, err
}
