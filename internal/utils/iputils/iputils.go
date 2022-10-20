package iputils

import (
	"bytes"
	"context"
	"net"
	"strings"
	"time"

	"github.com/oidc-mytoken/server/internal/utils/cache"
)

const hostLookupTimeout = 50 * time.Millisecond

func getHost(ip string) string {
	cacheHost, found := cache.Get(cache.IPHostCache, ip)
	if found {
		return cacheHost.(string)
	}
	ctx, cancel := context.WithTimeout(context.TODO(), hostLookupTimeout)
	defer cancel()
	r := net.Resolver{
		// PreferGo: true,
	}
	hosts, err := r.LookupAddr(ctx, ip)
	if err != nil && len(hosts) < 1 {
		return ""
	}
	host := hosts[0]
	cache.Set(cache.IPHostCache, ip, host)
	return host
}

// IPsAreSubSet checks if all ips of ipsA are contained in ipsB, it will also check ip subnets
func IPsAreSubSet(ipsA, ipsB []string) bool {
	for _, ipA := range ipsA {
		if !IPIsIn(ipA, ipsB) {
			return false
		}
	}
	return true
}

func parseIP(ip string) (net.IP, *net.IPNet) {
	ipA, ipNet, found := cache.GetIPParseResult(ip)
	if found {
		return ipA, ipNet
	}
	var err error
	ipA, ipNet, err = net.ParseCIDR(ip)
	if err != nil {
		ipA = net.ParseIP(ip)
	}
	if ipNet != nil && !ipA.Equal(ipNet.IP) {
		ipNet = nil
	}
	cache.SetIPParseResult(ip, ipA, ipNet)
	return ipA, ipNet
}

// IPIsIn checks if an ip is in a slice of ip/hosts, it will also check ip subnets
func IPIsIn(ip string, ipOrHosts []string) bool {
	for _, ipOrHost := range ipOrHosts {
		if compareIPToIPOrHost(ip, ipOrHost) {
			return true
		}
	}
	return false
}

func compareIPToIP(ip, ipp string) bool {
	ipA, ipNetA := parseIP(ip)
	ipB, ipNetB := parseIP(ipp)
	if ipNetA == nil && ipNetB == nil {
		if ipA.Equal(ipB) {
			return true
		}
	} else if ipNetA == nil && ipNetB != nil {
		if ipNetB.Contains(ipA) {
			return true
		}
	} else if ipNetA != nil && ipNetB != nil {
		if ipNetB.Contains(ipA) && bytes.Compare(ipNetA.Mask, ipNetB.Mask) >= 0 {
			return true
		}
	}
	// check for ipNetA != nil && ipNetB == nil not needed -> won't work
	return false
}

func compareIPToHost(ip, host string) bool {
	ipHost := getHost(ip)
	if ipHost == "" {
		return false
	}
	if ipHost[len(ipHost)-1] == '.' && host[len(host)-1] != '.' {
		host += "."
	}
	if len(host) > 1 && host[0] == '*' {
		return strings.HasSuffix(ipHost, host[1:])
	}
	return strings.Compare(ipHost, host) == 0
}

func ipValid(ip net.IP) bool {
	return ip != nil && !ip.IsUnspecified()
}

func compareIPToIPOrHost(ip, iphost string) bool {
	ipA, _ := parseIP(ip)
	ipHostA, _ := parseIP(iphost)
	if ipValid(ipA) && !ipValid(ipHostA) {
		return compareIPToHost(ip, iphost)
	}
	if ipValid(ipA) && ipValid(ipHostA) {
		return compareIPToIP(ip, iphost)
	}
	if !ipValid(ipA) && !ipValid(ipHostA) {
		return compareHostToHost(ip, iphost)
	}
	if !ipValid(ipA) && ipValid(ipHostA) {
		return compareHostToIP(ip, iphost)
	}
	return false
}

func compareHostToIP(host, ip string) bool {
	if len(host) > 1 && host[0] == '*' {
		return false
	}
	if len(host) > 0 && host[len(host)-1] != '.' {
		host += "."
	}
	ipHost := getHost(ip)
	if len(ipHost) > 0 && ipHost[len(ipHost)-1] != '.' {
		ipHost += "."
	}
	return host == ipHost
}

func compareHostToHost(a, b string) bool {
	if len(a) > 0 && a[len(a)-1] != '.' {
		a += "."
	}
	if len(b) > 0 && b[len(b)-1] != '.' {
		b += "."
	}
	if a == b {
		return true
	}
	if len(a) > 1 && a[0] == '*' {
		if len(b) > 1 && b[0] != '*' {
			return false
		}
		return strings.HasSuffix(a[1:], b[1:])
	}
	if len(b) > 0 && b[0] != '*' {
		// a!=b
		// a[0]!='*'
		// b[0]!='*'
		// -> a really different from b
		return false
	}
	return len(b) > 1 && strings.HasSuffix(a, b[1:])
}
