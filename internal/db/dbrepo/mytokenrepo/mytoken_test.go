package mytokenrepo

import (
	"testing"

	"github.com/oidc-mytoken/server/shared/mytoken/pkg/mtid"
)

func testRoot(t *testing.T, a MytokenEntry, expected bool) {
	root := a.Root()
	if root != expected {
		if expected {
			t.Errorf("Actually '%+v' is a root entry", a)
		} else {
			t.Errorf("Actually '%+v' is not a root entry", a)
		}
	}
}

func TestSuperTokenEntry_RootEmpty(t *testing.T) {
	a := MytokenEntry{}
	testRoot(t, a, true)
}
func TestSuperTokenEntry_RootHasParentAsRoot(t *testing.T) {
	id := mtid.New()
	a := MytokenEntry{
		ParentID: id,
		RootID:   id,
	}
	testRoot(t, a, false)
}
func TestSuperTokenEntry_RootHasRoot(t *testing.T) {
	pid := mtid.New()
	rid := mtid.New()
	a := MytokenEntry{
		ParentID: pid,
		RootID:   rid,
	}
	testRoot(t, a, false)
}
