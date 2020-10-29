package utils

import "testing"

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
