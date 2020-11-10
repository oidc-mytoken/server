package utils

import "testing"

func TestSliceUnion_Empty(t *testing.T) {
	a := []string{}
	b := []string{}
	exp := []string{}
	u := SliceUnion(a, b)
	checkSlice(t, u, exp)
	u = SliceUnion(b, a)
	checkSlice(t, u, exp)
}

func TestSliceUnion_OneEmpty(t *testing.T) {
	a := []string{}
	b := []string{"a", "b"}
	exp := []string{"a", "b"}
	u := SliceUnion(a, b)
	checkSlice(t, u, exp)
	u = SliceUnion(b, a)
	checkSlice(t, u, exp)
}

func TestSliceUnion_Same(t *testing.T) {
	a := []string{"a", "b"}
	b := []string{"a", "b"}
	exp := []string{"a", "b"}
	u := SliceUnion(a, b)
	checkSlice(t, u, exp)
	u = SliceUnion(b, a)
	checkSlice(t, u, exp)
}

func TestSliceUnion_Distinct(t *testing.T) {
	a := []string{"a", "b"}
	b := []string{"c", "d"}
	exp := []string{"a", "b", "c", "d"}
	u := SliceUnion(a, b)
	checkSlice(t, u, exp)
	exp = []string{"c", "d", "a", "b"}
	u = SliceUnion(b, a)
	checkSlice(t, u, exp)
}

func TestSliceUnion_Mixed(t *testing.T) {
	a := []string{"a", "b", "c"}
	b := []string{"c", "b", "d"}
	exp := []string{"a", "b", "c", "d"}
	u := SliceUnion(a, b)
	checkSlice(t, u, exp)
	exp = []string{"c", "b", "d", "a"}
	u = SliceUnion(b, a)
	checkSlice(t, u, exp)
}
