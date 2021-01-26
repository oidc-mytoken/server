package issuerUtils

import "testing"

func TestCombineSubIssValid(t *testing.T) {
	str := CombineSubIss("sub", "iss")
	expected := "sub@iss"
	if str != expected {
		t.Errorf("Expected '%s', got '%s'", expected, str)
	}
}
func TestCombineSubIssEmptyIss(t *testing.T) {
	str := CombineSubIss("sub", "")
	expected := ""
	if str != expected {
		t.Errorf("Expected '%s', got '%s'", expected, str)
	}
}
func TestCombineSubIssEmptySub(t *testing.T) {
	str := CombineSubIss("", "iss")
	expected := ""
	if str != expected {
		t.Errorf("Expected '%s', got '%s'", expected, str)
	}
}

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

func TestGetIssuerWithAndWithoutSlashEmpty(t *testing.T) {
	iss0, iss1 := GetIssuerWithAndWithoutSlash("")
	iss0Expected := ""
	iss1Expected := "/"
	if iss0 != iss0Expected {
		t.Errorf("Iss0 Expected '%s', got '%s'", iss0Expected, iss0)
	}
	if iss1 != iss1Expected {
		t.Errorf("Iss1 Expected '%s', got '%s'", iss1Expected, iss1)
	}
}
func TestGetIssuerWithAndWithoutSlashTrailingSlash(t *testing.T) {
	iss0, iss1 := GetIssuerWithAndWithoutSlash("https://example.com/")
	iss0Expected := "https://example.com"
	iss1Expected := "https://example.com/"
	if iss0 != iss0Expected {
		t.Errorf("Iss0 Expected '%s', got '%s'", iss0Expected, iss0)
	}
	if iss1 != iss1Expected {
		t.Errorf("Iss1 Expected '%s', got '%s'", iss1Expected, iss1)
	}
}
func TestGetIssuerWithAndWithoutSlashNoTrailingSlash(t *testing.T) {
	iss0, iss1 := GetIssuerWithAndWithoutSlash("https://example.com")
	iss0Expected := "https://example.com"
	iss1Expected := "https://example.com/"
	if iss0 != iss0Expected {
		t.Errorf("Iss0 Expected '%s', got '%s'", iss0Expected, iss0)
	}
	if iss1 != iss1Expected {
		t.Errorf("Iss1 Expected '%s', got '%s'", iss1Expected, iss1)
	}
}
