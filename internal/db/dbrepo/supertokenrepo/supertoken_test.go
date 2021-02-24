package supertokenrepo

import (
	"testing"

	"github.com/oidc-mytoken/server/shared/supertoken/pkg/stid"
)

func testRoot(t *testing.T, a SuperTokenEntry, expected bool) {
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
	a := SuperTokenEntry{}
	testRoot(t, a, true)
}
func TestSuperTokenEntry_RootHasParentAsRoot(t *testing.T) {
	id := stid.New()
	a := SuperTokenEntry{ParentID: id, RootID: id}
	testRoot(t, a, false)
}
func TestSuperTokenEntry_RootHasRoot(t *testing.T) {
	pid := stid.New()
	rid := stid.New()
	a := SuperTokenEntry{ParentID: pid, RootID: rid}
	testRoot(t, a, false)
}
