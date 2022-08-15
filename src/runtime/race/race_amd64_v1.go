//go:build amd64 && !amd64.v3
// +build amd64,!amd64.v3

package race

import _ "runtime/race/internal/amd64v1"
