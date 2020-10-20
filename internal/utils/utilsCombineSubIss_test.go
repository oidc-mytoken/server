package utils

import "testing"

func TestCombineSubIssValid(t *testing.T) {
	str := CombineSubIss("sub", "iss")
	expected := "sub@iss"
	if str != expected {
		t.Errorf("Expected '%s', got '%s'", expected, str)
	}
}
func TestCombineSubIssEmptyIss(t *testing.T) {
	str := CombineSubIss("sub", "")
	expected := ""
	if str != expected {
		t.Errorf("Expected '%s', got '%s'", expected, str)
	}
}
func TestCombineSubIssEmptySub(t *testing.T) {
	str := CombineSubIss("", "iss")
	expected := ""
	if str != expected {
		t.Errorf("Expected '%s', got '%s'", expected, str)
	}
}
