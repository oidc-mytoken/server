package utils

import (
	"testing"
)

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
