package issuerutils

import (
	"fmt"
	"strings"
)

// GetIssuerWithAndWithoutSlash takes an issuer url that might or might not end with a slash and returns the variant
// with and without trailing slash
func GetIssuerWithAndWithoutSlash(iss string) (string, string) {
	iss0 := iss
	var iss1 string
	if strings.HasSuffix(iss0, "/") {
		iss1 = strings.TrimSuffix(iss0, "/")
		return iss1, iss0
	}
	iss1 = fmt.Sprintf("%s%c", iss0, '/')
	return iss0, iss1
}

// CompareIssuerURLs compares two issuer urls. Issuer urls are also accepted as
// equal if they only differ in a trailing slash.
func CompareIssuerURLs(a, b string) bool {
	if a == b {
		return true
	}
	aLen := len(a)
	bLen := len(b)
	if bLen == aLen-1 {
		a, b = b, a
		aLen, bLen = bLen, aLen
	}
	if aLen == bLen-1 && b[bLen-1] == '/' {
		if a == b[:bLen-1] {
			return true
		}
	}
	return false
}

// CombineSubIss combines subject and issuer
func CombineSubIss(sub, iss string) string {
	if sub == "" || iss == "" {
		return ""
	}
	return fmt.Sprintf("%s@%s", sub, iss)
}
