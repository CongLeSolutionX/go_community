// +build !windows

package filepath

import "os"

// readLink returns the destination of the symbolic link path.
func readLink(path string) (string, bool, error) {
	p, err := os.Readlink(path)
	if err != nil {
		return "", false, err
	}
	return p, true, nil
}

func evalSymlinks(path string) (string, error) {
	return walkSymlinks(path)
}
