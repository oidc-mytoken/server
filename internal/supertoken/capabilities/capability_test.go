package capabilities

import "testing"

func fail(t *testing.T, expected, got Capabilities) {
	t.Errorf("Expected '%v', got '%v'", expected, got)
}

func testTighten(t *testing.T, a, b, expected Capabilities) {
	intersect := Tighten(a, b)
	if len(intersect) != len(expected) {
		fail(t, expected, intersect)
	}
	for i, ee := range expected {
		if ee != intersect[i] {
			fail(t, expected, intersect)
		}
	}
}

func TestTightenAllEmpty(t *testing.T) {
	a := Capabilities{}
	b := Capabilities{}
	expected := Capabilities{}
	testTighten(t, a, b, expected)
}
func TestTightenOneEmpty(t *testing.T) {
	a := Capabilities{}
	b := Capabilities{"not", "empty"}
	expected := Capabilities{}
	testTighten(t, a, b, expected)
	testTighten(t, b, a, expected)
}
func TestTightenNoIntersection(t *testing.T) {
	a := Capabilities{"some", "values"}
	b := Capabilities{"completly", "different"}
	expected := Capabilities{}
	testTighten(t, a, b, expected)
	testTighten(t, b, a, expected)
}
func TestTightenSame(t *testing.T) {
	a := Capabilities{"some", "values"}
	testTighten(t, a, a, a)
}
func TestTightenSomeIntersection(t *testing.T) {
	a := Capabilities{"some", "values"}
	b := Capabilities{"some", "different"}
	expected := Capabilities{"some"}
	testTighten(t, a, b, expected)
	testTighten(t, b, a, expected)
}
func TestTightenSubSet(t *testing.T) {
	a := Capabilities{"some", "values"}
	b := Capabilities{"some", "more", "values"}
	expected := Capabilities{"some", "values"}
	testTighten(t, a, b, expected)
	testTighten(t, b, a, expected)
}
