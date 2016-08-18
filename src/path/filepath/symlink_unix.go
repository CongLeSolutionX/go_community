// +build !windows

package filepath

import (
	"os"
	"syscall"
)

// return true if the error indicates that a file is exist and it's not a symlink.
func isNotSymlink(err error) bool {
	if err, ok := err.(*os.PathError); ok && err.Err == syscall.EINVAL {
		return true
	}
	return false
}

func evalSymlinks(path string) (string, error) {
	return walkSymlinks(path)
}
