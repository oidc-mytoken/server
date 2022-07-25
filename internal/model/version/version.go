package version

import (
	"fmt"
)

// Version segments
const (
	MAJOR = 0
	MINOR = 5
	FIX   = 0
	DEV   = true
)

var version = fmt.Sprintf("%d.%d.%d", MAJOR, MINOR, FIX)
var devVersion = fmt.Sprintf("%s-dev", version)

// VERSION returns the current mytoken version
func VERSION() string {
	if DEV {
		return devVersion
	}
	return version
}
