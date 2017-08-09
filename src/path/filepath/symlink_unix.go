// +build !windows

package filepath

import (
	"os"
)

// readLink returns the destination of the symbolic link path.
func readLink(path string) (string, error) {
	return os.Readlink(path)
}

func evalSymlinks(path string) (string, error) {
	return walkSymlinks(path)
}
