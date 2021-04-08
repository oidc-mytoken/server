package utils

import "testing"

func fail(t *testing.T, expected, got []string) {
	t.Errorf("Expected '%v', got '%v'", expected, got)
}

func testCombineURLs(t *testing.T, a, b, expected string) {
	url := CombineURLPath(a, b)
	if url != expected {
		t.Errorf("Expected '%s', got '%s'", expected, url)
	}
}
func TestCombineURLPathAllEmpty(t *testing.T) {
	a := ""
	b := ""
	expected := ""
	testCombineURLs(t, a, b, expected)
}
func TestCombineURLPathOneEmpty(t *testing.T) {
	a := "https://example.com"
	b := ""
	expected := a
	testCombineURLs(t, a, b, expected)
	testCombineURLs(t, b, a, expected)
}
func TestCombineURLPathNoSlash(t *testing.T) {
	a := "https://example.com"
	b := "api"
	expected := "https://example.com/api"
	testCombineURLs(t, a, b, expected)
}
func TestCombineURLPathNoSlashTrailingSlash(t *testing.T) {
	a := "https://example.com"
	b := "api/"
	expected := "https://example.com/api/"
	testCombineURLs(t, a, b, expected)
}
func TestCombineURLPathOneSlashA(t *testing.T) {
	a := "https://example.com/"
	b := "api"
	expected := "https://example.com/api"
	testCombineURLs(t, a, b, expected)
}
func TestCombineURLPathOneSlashB(t *testing.T) {
	a := "https://example.com"
	b := "/api"
	expected := "https://example.com/api"
	testCombineURLs(t, a, b, expected)
}
func TestCombineURLPathBothSlash(t *testing.T) {
	a := "https://example.com/"
	b := "/api"
	expected := "https://example.com/api"
	testCombineURLs(t, a, b, expected)
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
	b := []string{"completely", "different"}
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

func TestStringInSliceFirstPosition(t *testing.T) {
	str := "key"
	slice := []string{str, "other", "another"}
	found := StringInSlice(str, slice)
	if found != true {
		t.Errorf("'%s' not found in slice '%v'", str, slice)
	}
}
func TestStringInSliceLastPosition(t *testing.T) {
	str := "key"
	slice := []string{"other", "another", str}
	found := StringInSlice(str, slice)
	if found != true {
		t.Errorf("'%s' not found in slice '%v'", str, slice)
	}
}
func TestStringInSliceMidPosition(t *testing.T) {
	str := "key"
	slice := []string{"other", str, "another"}
	found := StringInSlice(str, slice)
	if found != true {
		t.Errorf("'%s' not found in slice '%v'", str, slice)
	}
}
func TestStringInSliceOnly(t *testing.T) {
	str := "key"
	slice := []string{str}
	found := StringInSlice(str, slice)
	if found != true {
		t.Errorf("'%s' not found in slice '%v'", str, slice)
	}
}
func TestStringInSliceEmpty(t *testing.T) {
	str := "key"
	slice := []string{}
	found := StringInSlice(str, slice)
	if found != false {
		t.Errorf("'%s' not found in slice '%v'", str, slice)
	}
}
func TestStringInSliceNotFound(t *testing.T) {
	str := "key"
	slice := []string{"only", "other", "strings"}
	found := StringInSlice(str, slice)
	if found != false {
		t.Errorf("'%s' not found in slice '%v'", str, slice)
	}
}

func TestReplaceStringInSlice_Normal(t *testing.T) {
	strs := []string{"a", "b", "c"}
	o := "a"
	n := "b"
	exp := []string{"b", "b", "c"}
	ReplaceStringInSlice(&strs, o, n, true)
	checkSlice(t, strs, exp)
}
func TestReplaceStringInSlice_Multiple(t *testing.T) {
	strs := []string{"a", "b", "d", "a", "c"}
	o := "a"
	n := "b"
	exp := []string{"b", "b", "d", "b", "c"}
	ReplaceStringInSlice(&strs, o, n, true)
	checkSlice(t, strs, exp)
}
func TestReplaceStringInSlice_CaseSensitivity(t *testing.T) {
	strs := []string{"a", "b", "A", "b"}
	o := "a"
	n := "c"
	exp := []string{"c", "b", "A", "b"}
	ReplaceStringInSlice(&strs, o, n, true)
	checkSlice(t, strs, exp)
	strs = []string{"a", "b", "A", "b"}
	exp = []string{"c", "b", "c", "b"}
	ReplaceStringInSlice(&strs, o, n, false)
	checkSlice(t, strs, exp)
}
func TestReplaceStringInSlice_Empty(t *testing.T) {
	strs := []string{}
	o := "a"
	n := "b"
	exp := []string{}
	ReplaceStringInSlice(&strs, o, n, true)
	checkSlice(t, strs, exp)
}
func TestReplaceStringInSlice_NotFound(t *testing.T) {
	strs := []string{"a", "b", "c"}
	o := "d"
	n := "b"
	exp := []string{"a", "b", "c"}
	ReplaceStringInSlice(&strs, o, n, true)
	checkSlice(t, strs, exp)
}

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

func TestSplitIgnoreEmpty_Empty(t *testing.T) {
	s := ""
	exp := []string{}
	split := SplitIgnoreEmpty(s, " ")
	checkSlice(t, split, exp)
}
func TestSplitIgnoreEmpty_Normal(t *testing.T) {
	s := "a b c d"
	exp := []string{"a", "b", "c", "d"}
	split := SplitIgnoreEmpty(s, " ")
	checkSlice(t, split, exp)
}
func TestSplitIgnoreEmpty_MultipleEmpty(t *testing.T) {
	s := "a b  c    d"
	exp := []string{"a", "b", "c", "d"}
	split := SplitIgnoreEmpty(s, " ")
	checkSlice(t, split, exp)
}
