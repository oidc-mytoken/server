package utils

import "testing"

func testIsSubSet(t *testing.T, a, b []string, expected bool) {
	ok := IsSubSet(a, b)
	if ok != expected {
		if expected {
			t.Errorf("Actually '%v' is a subset of '%v'", a, b)
		} else {
			t.Errorf("Actually '%v' is not a subset of '%v'", a, b)
		}
	}
}

func TestIsSubSetAllEmpty(t *testing.T) {
	a := []string{}
	b := []string{}
	testIsSubSet(t, a, b, true)
}
func TestIsSubSetOneEmpty(t *testing.T) {
	a := []string{}
	b := []string{"some", "values"}
	testIsSubSet(t, a, b, true)
	testIsSubSet(t, b, a, false)
}
func TestIsSubSetSubset(t *testing.T) {
	a := []string{"some"}
	b := []string{"some", "values"}
	testIsSubSet(t, a, b, true)
	testIsSubSet(t, b, a, false)
}
func TestIsSubSetDistinct(t *testing.T) {
	a := []string{"other"}
	b := []string{"some", "values"}
	testIsSubSet(t, a, b, false)
	testIsSubSet(t, b, a, false)
}
