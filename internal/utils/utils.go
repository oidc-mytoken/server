package utils

import (
	"fmt"
	"net"
	"strings"
)

// CombineSubIss combines subject and issuer
func CombineSubIss(sub, iss string) string {
	if sub == "" || iss == "" {
		return ""
	}
	return fmt.Sprintf("%s@%s", sub, iss)
}

func IntersectSlices(a, b []string) (res []string) {
	for _, bb := range b {
		if StringInSlice(bb, a) {
			res = append(res, bb)
		}
	}
	return
}

func StringInSlice(key string, slice []string) bool {
	for _, s := range slice {
		if s == key {
			return true
		}
	}
	return false
}

func IsSubSet(a, b []string) bool {
	for _, aa := range a {
		if !StringInSlice(aa, b) {
			return false
		}
	}
	return true
}

func IPsAreSubSet(ipsA, ipsB []string) bool {
	for _, ipA := range ipsA {
		if !ipIsIn(ipA, ipsB) {
			return false
		}
	}
	return true
}

func ipIsIn(ip string, ips []string) bool {
	ipA := net.ParseIP(ip)
	for _, ipp := range ips {
		if strings.Contains(ipp, "/") {
			_, ipNetB, _ := net.ParseCIDR(ipp)
			if ipNetB.Contains(ipA) {
				return true
			}
		} else {
			if ip == ipp {
				return true
			}
		}

	}
	return false
}

// CombineURLPath combines multiple parts of a url
func CombineURLPath(p string, ps ...string) (r string) {
	r = p
	for _, pp := range ps {
		if pp == "" {
			continue
		}
		if r == "" {
			r = pp
			continue
		}
		rAppend := r
		ppAppend := pp
		if strings.HasSuffix(r, "/") {
			rAppend = r[:len(r)-1]
		}
		if strings.HasPrefix(pp, "/") {
			ppAppend = pp[1:]
		}
		r = fmt.Sprintf("%s%c%s", rAppend, '/', ppAppend)
	}
	return
}
