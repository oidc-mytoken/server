package utils

import "testing"

func fail(t *testing.T, expected, got []string) {
	t.Errorf("Expected '%v', got '%v'", expected, got)
}

func testIntersectList(t *testing.T, a, b, expected []string) {
	intersect := IntersectSlices(a, b)
	if len(intersect) != len(expected) {
		fail(t, expected, intersect)
	}
	for i, ee := range expected {
		if ee != intersect[i] {
			fail(t, expected, intersect)
		}
	}
}

func TestIntersectSlicesAllEmpty(t *testing.T) {
	a := []string{}
	b := []string{}
	expected := []string{}
	testIntersectList(t, a, b, expected)
}
func TestIntersectSlicesOneEmpty(t *testing.T) {
	a := []string{}
	b := []string{"not", "empty"}
	expected := []string{}
	testIntersectList(t, a, b, expected)
	testIntersectList(t, b, a, expected)
}
func TestIntersectSlicesNoIntersection(t *testing.T) {
	a := []string{"some", "values"}
	b := []string{"completly", "different"}
	expected := []string{}
	testIntersectList(t, a, b, expected)
	testIntersectList(t, b, a, expected)
}
func TestIntersectSlicesSame(t *testing.T) {
	a := []string{"some", "values"}
	testIntersectList(t, a, a, a)
}
func TestIntersectSlicesSomeIntersection(t *testing.T) {
	a := []string{"some", "values"}
	b := []string{"some", "different"}
	expected := []string{"some"}
	testIntersectList(t, a, b, expected)
	testIntersectList(t, b, a, expected)
}
func TestIntersectSlicesSubSet(t *testing.T) {
	a := []string{"some", "values"}
	b := []string{"some", "more", "values"}
	expected := []string{"some", "values"}
	testIntersectList(t, a, b, expected)
	testIntersectList(t, b, a, expected)
}
