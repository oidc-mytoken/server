package issuerutils

import "testing"

func TestCombineSubIss(t *testing.T) {
	tests := []struct {
		name     string
		sub      string
		iss      string
		expected string
	}{
		{
			name:     "Valid",
			sub:      "sub",
			iss:      "iss",
			expected: "sub@iss",
		},
		{
			name:     "EmptyIss",
			sub:      "sub",
			iss:      "",
			expected: "",
		},
		{
			name:     "EmptySub",
			sub:      "",
			iss:      "iss",
			expected: "",
		},
	}
	for _, test := range tests {
		t.Run(
			test.name, func(t *testing.T) {
				str := CombineSubIss(test.sub, test.iss)
				if str != test.expected {
					t.Errorf("Expected '%s', got '%s'", test.expected, str)
				}
			},
		)
	}
}

func TestCompareIssuerURLs(t *testing.T) {
	tests := []struct {
		name     string
		url1     string
		url2     string
		expected bool
	}{
		{
			name:     "BothEmpty",
			url1:     "",
			url2:     "",
			expected: true,
		},
		{
			name:     "OneEmpty",
			url1:     "",
			url2:     "https://example.com",
			expected: false,
		},
		{
			name:     "SameTrailingSlash",
			url1:     "https://example.com/",
			url2:     "https://example.com/",
			expected: true,
		},
		{
			name:     "SameNoTrailingSlash",
			url1:     "https://example.com",
			url2:     "https://example.com",
			expected: true,
		},
		{
			name:     "SameDifferentTrailingSlash",
			url1:     "https://example.com/",
			url2:     "https://example.com",
			expected: true,
		},
	}
	for _, test := range tests {
		t.Run(
			test.name, func(t *testing.T) {
				pairs := [2][2]string{
					{
						test.url1,
						test.url2,
					},
					{
						test.url2,
						test.url1,
					},
				}
				for _, p := range pairs {
					if CompareIssuerURLs(p[0], p[1]) != test.expected {
						fmt := "URLs '%s' and '%s' are not correctly recognized as "
						if test.expected {
							fmt += "equal"
						} else {
							fmt += "different"
						}
						t.Errorf(fmt, p[0], p[1])
					}
				}
			},
		)
	}
}

func TestGetIssuerWithAndWithoutSlash(t *testing.T) {
	tests := []struct {
		name               string
		in                 string
		expectedOutWithout string
		expectedOutWith    string
	}{
		{
			name:               "Empty",
			in:                 "",
			expectedOutWithout: "",
			expectedOutWith:    "/",
		},
		{
			name:               "TrailingSlash",
			in:                 "https://example.com/",
			expectedOutWithout: "https://example.com",
			expectedOutWith:    "https://example.com/",
		},
		{
			name:               "NoTrailingSlash",
			in:                 "https://example.com",
			expectedOutWithout: "https://example.com",
			expectedOutWith:    "https://example.com/",
		},
	}
	for _, test := range tests {
		t.Run(
			test.name, func(t *testing.T) {
				issWO, issW := GetIssuerWithAndWithoutSlash(test.in)
				if issWO != test.expectedOutWithout {
					t.Errorf(
						"Expected '%s' as issuer without trailing slash, but got '%s'", test.expectedOutWithout, issWO,
					)
				}
				if issW != test.expectedOutWith {
					t.Errorf("Expected '%s' as issuer with trailing slash, but got '%s'", test.expectedOutWith, issW)
				}
			},
		)
	}
}
