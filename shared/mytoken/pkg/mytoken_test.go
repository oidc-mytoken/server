package mytoken

import (
	"testing"

	"github.com/oidc-mytoken/server/shared/utils/unixtime"
)

func TestMytoken_ExpiresIn(t *testing.T) {
	tests := []struct {
		name     string
		mt       Mytoken
		expected uint64
	}{
		{
			name:     "Empty",
			mt:       Mytoken{},
			expected: 0,
		},
		{
			name:     "Valid",
			mt:       Mytoken{ExpiresAt: unixtime.InSeconds(100)},
			expected: 100,
		},
		{
			name:     "Past",
			mt:       Mytoken{ExpiresAt: 100},
			expected: 0,
		},
	}
	for _, test := range tests {
		t.Run(
			test.name, func(t *testing.T) {
				expires := test.mt.ExpiresIn()
				if expires != test.expected {
					t.Errorf(
						"Expected expires in for '%+v' to be '%d', but instead is '%d'", test.mt, test.expected, expires,
					)
				}
			},
		)
	}
}
