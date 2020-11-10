package issuerUtils

import "testing"

func TestCompareIssuerURLBothEmpty(t *testing.T) {
	if CompareIssuerURLs("", "") != true {
		t.Errorf("Empty issuer urls should be equal")
	}
}
func TestCompareIssuerURLOneEmpty(t *testing.T) {
	a := "https://example.com"
	b := ""
	if CompareIssuerURLs(a, b) == true {
		t.Errorf("An empty issuer url should not equal a non-empty")
	}
	if CompareIssuerURLs(b, a) == true {
		t.Errorf("An empty issuer url should not equal a non-empty")
	}
}
func TestCompareIssuerURLSame(t *testing.T) {
	a := "https://example.com"
	b := a
	if CompareIssuerURLs(a, b) != true {
		t.Errorf("Equal strings should be equal")
	}
	a = "https://example.com/"
	b = a
	if CompareIssuerURLs(a, b) != true {
		t.Errorf("Equal strings should be equal")
	}
}

func TestCompareIssuerURLDifferentSlash(t *testing.T) {
	a := "https://example.com"
	b := "https://example.com/"
	if CompareIssuerURLs(a, b) != true {
		t.Errorf("Issuer urls only differing in trailing slash should be equal")
	}
	if CompareIssuerURLs(b, a) != true {
		t.Errorf("Issuer urls only differing in trailing slash should be equal")
	}
}
