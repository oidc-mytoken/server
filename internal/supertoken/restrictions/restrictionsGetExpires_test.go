package restrictions

import "testing"

func TestRestrictions_GetExpiresEmpty(t *testing.T) {
	r := Restrictions{}
	expires := r.GetExpires()
	var expected int64 = 0
	if expected != expires {
		t.Errorf("Expected %d, but got %d", expected, expires)
	}
}

func TestRestrictions_GetExpiresInfinite(t *testing.T) {
	r := Restrictions{
		{ExpiresAt: 0},
	}
	expires := r.GetExpires()
	var expected int64 = 0
	if expected != expires {
		t.Errorf("Expected %d, but got %d", expected, expires)
	}
}

func TestRestrictions_GetExpiresOne(t *testing.T) {
	r := Restrictions{
		{ExpiresAt: 100},
	}
	expires := r.GetExpires()
	var expected int64 = 100
	if expected != expires {
		t.Errorf("Expected %d, but got %d", expected, expires)
	}
}

func TestRestrictions_GetExpiresMultiple(t *testing.T) {
	r := Restrictions{
		{ExpiresAt: 100},
		{ExpiresAt: 300},
		{ExpiresAt: 200},
	}
	expires := r.GetExpires()
	var expected int64 = 300
	if expected != expires {
		t.Errorf("Expected %d, but got %d", expected, expires)
	}
}

func TestRestrictions_GetExpiresMultipleAndInfinite(t *testing.T) {
	r := Restrictions{
		{ExpiresAt: 100},
		{ExpiresAt: 0},
		{ExpiresAt: 300},
		{ExpiresAt: 200},
	}
	expires := r.GetExpires()
	var expected int64 = 0
	if expected != expires {
		t.Errorf("Expected %d, but got %d", expected, expires)
	}
}
