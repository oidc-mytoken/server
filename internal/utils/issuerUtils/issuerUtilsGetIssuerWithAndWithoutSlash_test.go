package issuerUtils

import "testing"

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
