package utils

import (
	"testing"

	"github.com/jinzhu/copier"
)

func fail(t *testing.T, expected, got []string) {
	t.Errorf("Expected '%v', got '%v'", expected, got)
}

func TestCombineURLPath(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected string
	}{
		{
			name:     "Empty",
			a:        "",
			b:        "",
			expected: "",
		},
		{
			name:     "LastEmpty",
			a:        "https://example.com",
			b:        "",
			expected: "https://example.com",
		},
		{
			name:     "FirstEmpty",
			a:        "",
			b:        "https://example.com",
			expected: "https://example.com",
		},
		{
			name:     "NoSlashes",
			a:        "https://example.com",
			b:        "api",
			expected: "https://example.com/api",
		},
		{
			name:     "TrailingSlash",
			a:        "https://example.com",
			b:        "api/",
			expected: "https://example.com/api/",
		},
		{
			name:     "SlashOnA",
			a:        "https://example.com/",
			b:        "api",
			expected: "https://example.com/api",
		},
		{
			name:     "SlashOnB",
			a:        "https://example.com",
			b:        "/api",
			expected: "https://example.com/api",
		},
		{
			name:     "BothSlashes",
			a:        "https://example.com/",
			b:        "/api",
			expected: "https://example.com/api",
		},
	}
	for _, test := range tests {
		t.Run(
			test.name, func(t *testing.T) {
				url := CombineURLPath(test.a, test.b)
				if url != test.expected {
					t.Errorf("Expected '%s', got '%s'", test.expected, url)
				}
			},
		)
	}
}

func TestIntersectSlices(t *testing.T) {
	tests := []struct {
		name     string
		a        []string
		b        []string
		expected []string
	}{
		{
			name:     "AllEmpty",
			a:        []string{},
			b:        []string{},
			expected: []string{},
		},
		{
			name: "FirstEmpty",
			a:    []string{},
			b: []string{
				"not",
				"empty",
			},
			expected: []string{},
		},
		{
			name: "SecondEmpty",
			a: []string{
				"not",
				"empty",
			},
			b:        []string{},
			expected: []string{},
		},
		{
			name: "NoIntersection",
			a: []string{
				"not",
				"empty",
			},
			b: []string{
				"other",
				"values",
			},
			expected: []string{},
		},
		{
			name: "Same",
			a: []string{
				"not",
				"empty",
			},
			b: []string{
				"not",
				"empty",
			},
			expected: []string{
				"not",
				"empty",
			},
		},
		{
			name: "SomeIntersection",
			a: []string{
				"not",
				"empty",
				"same",
			},
			b: []string{
				"other",
				"same",
				"different",
			},
			expected: []string{"same"},
		},
		{
			name: "SubSet",
			a: []string{
				"not",
				"empty",
				"same",
			},
			b: []string{
				"not",
				"same",
			},
			expected: []string{
				"not",
				"same",
			},
		},
	}
	for _, test := range tests {
		t.Run(
			test.name, func(t *testing.T) {
				intersect := IntersectSlices(test.a, test.b)
				if len(intersect) != len(test.expected) {
					fail(t, test.expected, intersect)
				}
				for i, ee := range test.expected {
					if ee != intersect[i] {
						fail(t, test.expected, intersect)
					}
				}
			},
		)
	}
}

func TestIsSubSet(t *testing.T) {
	tests := []struct {
		name     string
		a        []string
		b        []string
		expected bool
	}{
		{
			name:     "AllEmpty",
			a:        []string{},
			b:        []string{},
			expected: true,
		},
		{
			name: "FirstEmpty",
			a:    []string{},
			b: []string{
				"some",
				"value",
			},
			expected: true,
		},
		{
			name: "SecondEmpty",
			a: []string{
				"some",
				"value",
			},
			b:        []string{},
			expected: false,
		},
		{
			name: "Same",
			a: []string{
				"some",
				"value",
			},
			b: []string{
				"some",
				"value",
			},
			expected: true,
		},
		{
			name: "SameDifferentOrder",
			a: []string{
				"some",
				"value",
			},
			b: []string{
				"value",
				"some",
			},
			expected: true,
		},
		{
			name: "TrueSubSet",
			a:    []string{"value"},
			b: []string{
				"some",
				"value",
			},
			expected: true,
		},
		{
			name: "SuperSet",
			a: []string{
				"some",
				"value",
			},
			b:        []string{"value"},
			expected: false,
		},
		{
			name: "Distinct",
			a: []string{
				"value",
				"different",
			},
			b: []string{
				"some",
				"other",
			},
			expected: false,
		},
	}
	for _, test := range tests {
		t.Run(
			test.name, func(t *testing.T) {
				ok := IsSubSet(test.a, test.b)
				if ok != test.expected {
					if test.expected {
						t.Errorf("Actually '%v' is a subset of '%v'", test.a, test.b)
					} else {
						t.Errorf("Actually '%v' is not a subset of '%v'", test.a, test.b)
					}
				}
			},
		)
	}
}

func TestSliceUnion(t *testing.T) {
	tests := []struct {
		name         string
		a            []string
		b            []string
		expected     []string
		secondExpect []string
	}{
		{
			name:     "AllEmpty",
			a:        []string{},
			b:        []string{},
			expected: []string{},
		},
		{
			name: "OneEmpty",
			a:    []string{},
			b: []string{
				"a",
				"b",
			},
			expected: []string{
				"a",
				"b",
			},
		},
		{
			name: "Same",
			a: []string{
				"a",
				"b",
			},
			b: []string{
				"a",
				"b",
			},
			expected: []string{
				"a",
				"b",
			},
		},
		{
			name: "Distinct",
			a: []string{
				"a",
				"b",
			},
			b: []string{
				"c",
				"d",
			},
			expected: []string{
				"a",
				"b",
				"c",
				"d",
			},
			secondExpect: []string{
				"c",
				"d",
				"a",
				"b",
			},
		},
		{
			name: "Mixed",
			a: []string{
				"a",
				"b",
				"c",
			},
			b: []string{
				"c",
				"d",
				"e",
			},
			expected: []string{
				"a",
				"b",
				"c",
				"d",
				"e",
			},
			secondExpect: []string{
				"c",
				"d",
				"e",
				"a",
				"b",
			},
		},
	}
	for _, test := range tests {
		if test.secondExpect == nil {
			test.secondExpect = test.expected
		}
		t.Run(
			test.name, func(t *testing.T) {
				u := SliceUnion(test.a, test.b)
				checkSlice(t, u, test.expected)
				u = SliceUnion(test.b, test.a)
				checkSlice(t, u, test.secondExpect)
			},
		)
	}
}

func TestStringInSlice(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		slice    []string
		expected bool
	}{
		{
			name: "First Position",
			str:  "key",
			slice: []string{
				"key",
				"second",
				"third",
			},
			expected: true,
		},
		{
			name: "Mid Position",
			str:  "key",
			slice: []string{
				"first",
				"key",
				"third",
			},
			expected: true,
		},
		{
			name: "Last Position",
			str:  "key",
			slice: []string{
				"first",
				"second",
				"key",
			},
			expected: true,
		},
		{
			name:     "Only Key",
			str:      "key",
			slice:    []string{"key"},
			expected: true,
		},
		{
			name:     "Empty Slice",
			str:      "key",
			slice:    []string{},
			expected: false,
		},
		{
			name: "Not Found",
			str:  "key",
			slice: []string{
				"first",
				"second",
				"third",
			},
			expected: false,
		},
	}
	for _, test := range tests {
		t.Run(
			test.name, func(t *testing.T) {
				found := StringInSlice(test.str, test.slice)
				if found != test.expected {
					fmt := "'%s'%s found in slice '%+q', but should%s"
					if found {
						t.Errorf(fmt, test.str, "", test.slice, " not")
					} else {
						t.Errorf(fmt, test.str, " not", test.slice, "have")
					}
				}
			},
		)
	}
}

func TestReplaceStringInSlice(t *testing.T) {
	tests := []struct {
		name        string
		in          []string
		str         string
		replace     string
		expectedCIS []string
		expectedCS  []string
	}{
		{
			name: "Normal",
			in: []string{
				"a",
				"b",
				"c",
			},
			str:     "a",
			replace: "b",
			expectedCIS: []string{
				"b",
				"b",
				"c",
			},
			expectedCS: []string{
				"b",
				"b",
				"c",
			},
		},
		{
			name: "Multiple",
			in: []string{
				"a",
				"b",
				"c",
				"a",
				"d",
			},
			str:     "a",
			replace: "b",
			expectedCIS: []string{
				"b",
				"b",
				"c",
				"b",
				"d",
			},
			expectedCS: []string{
				"b",
				"b",
				"c",
				"b",
				"d",
			},
		},
		{
			name: "Case Sensitivity",
			in: []string{
				"a",
				"b",
				"A",
				"B",
			},
			str:     "a",
			replace: "b",
			expectedCIS: []string{
				"b",
				"b",
				"b",
				"B",
			},
			expectedCS: []string{
				"b",
				"b",
				"A",
				"B",
			},
		},
		{
			name:        "Empty Slice",
			in:          []string{},
			str:         "a",
			replace:     "b",
			expectedCIS: []string{},
			expectedCS:  []string{},
		},
		{
			name: "Not Found",
			in: []string{
				"a",
				"b",
				"c",
				"a",
				"d",
			},
			str:     "g",
			replace: "h",
			expectedCIS: []string{
				"a",
				"b",
				"c",
				"a",
				"d",
			},
			expectedCS: []string{
				"a",
				"b",
				"c",
				"a",
				"d",
			},
		},
	}
	for _, test := range tests {
		t.Run(
			test.name, func(t *testing.T) {
				var in []string
				copier.Copy(&in, &test.in)
				ReplaceStringInSlice(&in, test.str, test.replace, true)
				checkSlice(t, in, test.expectedCS)
				copier.Copy(&in, &test.in)
				ReplaceStringInSlice(&in, test.str, test.replace, false)
				checkSlice(t, in, test.expectedCIS)
			},
		)
	}
}

func failSlice(t *testing.T, a, exp []string) {
	t.Errorf("Expected '%+v', but got '%+v'", exp, a)
}
func checkSlice(t *testing.T, a, exp []string) {
	if len(a) != len(exp) {
		failSlice(t, a, exp)
		return
	}
	for i, e := range exp {
		if e != a[i] {
			failSlice(t, a, exp)
			return
		}
	}
}

func TestUniqueSlice(t *testing.T) {
	tests := []struct {
		name     string
		in       []string
		expected []string
	}{
		{
			name:     "Empty",
			in:       []string{},
			expected: []string{},
		},
		{
			name: "Unique",
			in: []string{
				"a",
				"b",
				"c",
			},
			expected: []string{
				"a",
				"b",
				"c",
			},
		},
		{
			name: "Duplicates",
			in: []string{
				"a",
				"b",
				"c",
				"a",
				"d",
				"d",
				"d",
				"e",
				"f",
				"b",
			},
			expected: []string{
				"a",
				"b",
				"c",
				"d",
				"e",
				"f",
			},
		},
	}
	for _, test := range tests {
		t.Run(
			test.name, func(t *testing.T) {
				u := UniqueSlice(test.in)
				checkSlice(t, u, test.expected)
			},
		)
	}
}

func TestSplitIgnoreEmpty(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		expected []string
	}{
		{
			name:     "Empty",
			str:      "",
			expected: []string{},
		},
		{
			name: "Normal",
			str:  "a b c d",
			expected: []string{
				"a",
				"b",
				"c",
				"d",
			},
		},
		{
			name: "Multiple Empty",
			str:  "a b  c   d",
			expected: []string{
				"a",
				"b",
				"c",
				"d",
			},
		},
	}
	for _, test := range tests {
		t.Run(
			test.name, func(t *testing.T) {
				split := SplitIgnoreEmpty(test.str, " ")
				checkSlice(t, split, test.expected)
			},
		)
	}
}
