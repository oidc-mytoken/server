package supertoken

import "testing"

//TODO
// func TestValid(t *testing.T) {
// a := newSuperToken{}
// 	testValid(t, a,  true)
// }

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

func TestRootEmpty(t *testing.T) {
	a := SuperTokenEntry{}
	testRoot(t, a, true)
}
func TestRootHasParentAsRoot(t *testing.T) {
	a := SuperTokenEntry{ParentID: "id", RootID: "id"}
	testRoot(t, a, false)
}
func TestRootHasRoot(t *testing.T) {
	a := SuperTokenEntry{ParentID: "parentid", RootID: "rootid"}
	testRoot(t, a, false)
}
