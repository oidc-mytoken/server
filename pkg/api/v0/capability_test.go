package api

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
	b := NewCapabilities([]string{"not", "empty"})
	expected := Capabilities{}
	testTighten(t, a, b, expected)
	testTighten(t, b, a, expected)
}
func TestTightenNoIntersection(t *testing.T) {
	a := NewCapabilities([]string{"some", "values"})
	b := NewCapabilities([]string{"completly", "different"})
	expected := Capabilities{}
	testTighten(t, a, b, expected)
	testTighten(t, b, a, expected)
}
func TestTightenSame(t *testing.T) {
	a := NewCapabilities([]string{"some", "values"})
	testTighten(t, a, a, a)
}
func TestTightenSomeIntersection(t *testing.T) {
	a := NewCapabilities([]string{"some", "values"})
	b := NewCapabilities([]string{"some", "different"})
	expected := NewCapabilities([]string{"some"})
	testTighten(t, a, b, expected)
	testTighten(t, b, a, expected)
}
func TestTightenSubSet(t *testing.T) {
	a := NewCapabilities([]string{"some", "values"})
	b := NewCapabilities([]string{"some", "more", "values"})
	expected := NewCapabilities([]string{"some", "values"})
	testTighten(t, a, b, expected)
	testTighten(t, b, a, expected)
}
