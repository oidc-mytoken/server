package utils

import "testing"

func failSlice(t *testing.T, a, exp []string) {
	t.Errorf("Expected '%+v', but go '%+v'", exp, a)
}

func checkSlice(t *testing.T, a, exp []string) {
	if len(a) != len(exp) {
		failSlice(t, a, exp)
		return
	}
	for i, e := range exp {
		if e != a[i] {
			failSlice(t, a, exp)
			return
		}
	}
}

func TestUniqueSlice_Empty(t *testing.T) {
	a := []string{}
	exp := []string{}
	u := UniqueSlice(a)
	checkSlice(t, u, exp)
}

func TestUniqueSlice_Unique(t *testing.T) {
	a := []string{"a", "b", "c"}
	exp := []string{"a", "b", "c"}
	u := UniqueSlice(a)
	checkSlice(t, u, exp)
}

func TestUniqueSlice_Duplicates(t *testing.T) {
	a := []string{"a", "b", "a", "c", "c", "d", "c", "e"}
	exp := []string{"a", "b", "c", "d", "e"}
	u := UniqueSlice(a)
	checkSlice(t, u, exp)
}
