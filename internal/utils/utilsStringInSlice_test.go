package utils

import "testing"

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
