package hidden

import "strings"

// IsHidden checks if a path is hidden or not
// Eg: .DS_Store in mac
// Eg: . in linux
func IsHidden(path string) (flag bool) {

	if strings.HasPrefix(path, ".") {
		flag = true
	}

	return
}
