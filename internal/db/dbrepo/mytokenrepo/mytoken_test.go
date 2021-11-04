package mytokenrepo

import (
	"testing"

	"github.com/oidc-mytoken/server/shared/mytoken/pkg/mtid"
)

func TestMytokenEntry_Root(t *testing.T) {
	parentRoot := mtid.New()
	tests := []struct {
		name     string
		mt       MytokenEntry
		expected bool
	}{
		{
			name:     "Empty",
			mt:       MytokenEntry{},
			expected: true,
		},
		{
			name: "HasParentAsRoot",
			mt: MytokenEntry{
				ParentID: parentRoot,
				RootID:   parentRoot,
			},
			expected: false,
		},
		{
			name: "HasParentAndRoot",
			mt: MytokenEntry{
				ParentID: mtid.New(),
				RootID:   mtid.New(),
			},
			expected: false,
		},
	}
	for _, test := range tests {
		t.Run(
			test.name, func(t *testing.T) {
				root := test.mt.Root()
				if root != test.expected {
					if test.expected {
						t.Errorf("Actually '%+v' is a root entry", test.mt)
					} else {
						t.Errorf("Actually '%+v' is not a root entry", test.mt)
					}
				}
			},
		)
	}
}
