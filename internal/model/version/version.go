package version

import (
	_ "embed" // for go:embed
	"strconv"
	"strings"
)

// SOFTWAREID is a unique string identifying this software
const SOFTWAREID = "ZWR1LmtpdC5teXRva2VuLjI5MTI=sm9K4rrvJppPgM"

// VERSION holds the server's version
//
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
		pre := strings.TrimPrefix(ps[1], "pr")
		PRE, _ = strconv.Atoi(pre)
	}
}
