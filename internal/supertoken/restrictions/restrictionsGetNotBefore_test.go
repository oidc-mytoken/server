package restrictions

import "testing"

func TestRestrictions_GetNotBeforeEmpty(t *testing.T) {
	r := Restrictions{}
	expires := r.GetNotBefore()
	var expected int64 = 0
	if expected != expires {
		t.Errorf("Expected %d, but got %d", expected, expires)
	}
}

func TestRestrictions_GetNotBeforeInfinite(t *testing.T) {
	r := Restrictions{
		{NotBefore: 0},
	}
	expires := r.GetNotBefore()
	var expected int64 = 0
	if expected != expires {
		t.Errorf("Expected %d, but got %d", expected, expires)
	}
}

func TestRestrictions_GetNotBeforeOne(t *testing.T) {
	r := Restrictions{
		{NotBefore: 100},
	}
	expires := r.GetNotBefore()
	var expected int64 = 100
	if expected != expires {
		t.Errorf("Expected %d, but got %d", expected, expires)
	}
}

func TestRestrictions_GetNotBeforeMultiple(t *testing.T) {
	r := Restrictions{
		{NotBefore: 100},
		{NotBefore: 300},
		{NotBefore: 200},
	}
	expires := r.GetNotBefore()
	var expected int64 = 100
	if expected != expires {
		t.Errorf("Expected %d, but got %d", expected, expires)
	}
}

func TestRestrictions_GetNotBeforeMultipleAndInfinite(t *testing.T) {
	r := Restrictions{
		{NotBefore: 100},
		{NotBefore: 0},
		{NotBefore: 300},
		{NotBefore: 200},
	}
	expires := r.GetNotBefore()
	var expected int64 = 0
	if expected != expires {
		t.Errorf("Expected %d, but got %d", expected, expires)
	}
}
