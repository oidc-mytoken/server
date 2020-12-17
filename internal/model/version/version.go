package version

import (
	"fmt"
)

// Version segments
const (
	MAJOR = 0
	MINOR = 1
	FIX   = 0
)

// VERSION is the current mytoken version
var VERSION = fmt.Sprintf("%d.%d.%d", MAJOR, MINOR, FIX)
