package mtid

import (
	"testing"
)

func TestMTID_HashValid(t *testing.T) {
	tests := []struct {
		name     string
		id       MTID
		expected bool
	}{
		{
			name:     "Empty",
			id:       MTID{},
			expected: false,
		},
		{
			name:     "Valid",
			id:       New(),
			expected: true,
		},
		{
			name: "OnlyHash",
			id: MTID{
				hash: "hash",
			},
			expected: true,
		},
	}
	for _, test := range tests {
		t.Run(
			test.name, func(t *testing.T) {
				if test.id.HashValid() != test.expected {
					var not string
					if !test.expected {
						not = "not "
					}
					t.Errorf("Expected MTID '%+q' to %shave valid hash, but HashValid() does not say so", test.id, not)
				}
			},
		)
	}
}
func TestMTID_Valid(t *testing.T) {
	tests := []struct {
		name     string
		id       MTID
		expected bool
	}{
		{
			name:     "Empty",
			id:       MTID{},
			expected: false,
		},
		{
			name:     "Valid",
			id:       New(),
			expected: true,
		},
		{
			name: "OnlyHash",
			id: MTID{
				hash: "hash",
			},
			expected: false,
		},
	}
	for _, test := range tests {
		t.Run(
			test.name, func(t *testing.T) {
				if test.id.Valid() != test.expected {
					var not string
					if !test.expected {
						not = "not "
					}
					t.Errorf("Expected MTID '%+q' to %sbe valid, but Valid() does not say so", test.id, not)
				}
			},
		)
	}
}
