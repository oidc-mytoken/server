package supertokenrepo

import "testing"

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
	a := SuperTokenEntry{ParentID: "id", RootID: "id"}
	testRoot(t, a, false)
}
func TestSuperTokenEntry_RootHasRoot(t *testing.T) {
	a := SuperTokenEntry{ParentID: "parentid", RootID: "rootid"}
	testRoot(t, a, false)
}
