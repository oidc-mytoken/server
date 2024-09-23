package model

import (
	"net"
)

// IPParseResult holds the result of ip parsing
type IPParseResult struct {
	IP    net.IP
	IPNet *net.IPNet
}
