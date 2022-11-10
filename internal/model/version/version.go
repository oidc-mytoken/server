package version

import (
	_ "embed"
	"strconv"
	"strings"
)

//go:embed VERSION
var VERSION string

// Version segments
var (
	MAJOR int
	MINOR int
	FIX   int
	PRE   int
)

func init() {
	if VERSION[len(VERSION)-1] == '\n' {
		VERSION = VERSION[:len(VERSION)-1]
	}
	v := strings.Split(VERSION, ".")
	MAJOR, _ = strconv.Atoi(v[0])
	MINOR, _ = strconv.Atoi(v[1])
	ps := strings.Split(v[2], "-")
	FIX, _ = strconv.Atoi(ps[0])
	if len(ps) > 1 {
		pre := ps[1]
		if strings.HasPrefix(pre, "pr") {
			pre = pre[2:]
		}
		PRE, _ = strconv.Atoi(pre)
	}
}
