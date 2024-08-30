package iputils

import (
	"bytes"
	"net"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/utils/cache"
)

func getHosts(ip string) (hosts []string) {
	found, err := cache.Get(cache.IPHostCache, ip, &hosts)
	if err == nil && found {
		return
	}
	hosts, err = net.LookupAddr(ip)
	if err != nil {
		return nil
	}
	if err = cache.Set(cache.IPHostCache, ip, hosts); err != nil {
		log.WithError(err).Error("error caching hosts")
	}
	return hosts
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

func parseIP(ip string) model.IPParseResult {
	result, found, err := cache.GetIPParseResult(ip)
	if err != nil {
		log.WithError(err).Error("error getting IP from cache")
	}
	if err == nil && found {
		return result
	}
	ipA, ipNet, err := net.ParseCIDR(ip)
	if err != nil {
		ipA = net.ParseIP(ip)
	}
	if ipNet != nil && !ipA.Equal(ipNet.IP) {
		ipNet = nil
	}
	result.IP = ipA
	result.IPNet = ipNet
	if err = cache.SetIPParseResult(ip, result); err != nil {
		log.WithError(err).Error("error caching IP")
	}
	return result
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
	resA := parseIP(ip)
	resB := parseIP(ipp)
	if resA.IPNet == nil && resB.IPNet == nil {
		if resA.IP.Equal(resB.IP) {
			return true
		}
	} else if resA.IPNet == nil && resB.IPNet != nil {
		if resB.IPNet.Contains(resA.IP) {
			return true
		}
	} else if resA.IPNet != nil && resB.IPNet != nil {
		if resB.IPNet.Contains(resA.IP) && bytes.Compare(resA.IPNet.Mask, resB.IPNet.Mask) >= 0 {
			return true
		}
	}
	// check for ipNetA != nil && ipNetB == nil not needed -> won't work
	return false
}

func compareIPToHost(ip, host string) bool {
	ipHosts := getHosts(ip)
	if len(ipHosts) == 0 {
		return false
	}
	for _, ipHost := range ipHosts {
		if ipHost[len(ipHost)-1] == '.' && host[len(host)-1] != '.' {
			host += "."
		}
		if len(host) > 1 && host[0] == '*' {
			if strings.HasSuffix(ipHost, host[1:]) {
				return true
			}
		} else if strings.Compare(ipHost, host) == 0 {
			return true
		}
	}
	return false
}

func ipValid(ip net.IP) bool {
	return ip != nil && !ip.IsUnspecified()
}

func compareIPToIPOrHost(ip, iphost string) bool {
	res := parseIP(ip)
	resHost := parseIP(iphost)
	if ipValid(res.IP) && !ipValid(resHost.IP) {
		return compareIPToHost(ip, iphost)
	}
	if ipValid(res.IP) && ipValid(resHost.IP) {
		return compareIPToIP(ip, iphost)
	}
	if !ipValid(res.IP) && !ipValid(resHost.IP) {
		return compareHostToHost(ip, iphost)
	}
	if !ipValid(res.IP) && ipValid(resHost.IP) {
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
	ipHosts := getHosts(ip)
	for _, ipHost := range ipHosts {
		if len(ipHost) > 0 && ipHost[len(ipHost)-1] != '.' {
			ipHost += "."
		}
		if host == ipHost {
			return true
		}
	}
	return false
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
