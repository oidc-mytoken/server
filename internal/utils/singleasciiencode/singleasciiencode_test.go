package singleasciiencode

import (
	"testing"
)

func assert(t *testing.T, name string, exp, got bool) {
	if exp != got {
		t.Errorf("Expected %v for %v, but got %v", exp, name, got)
	}
}

func TestFlagEncoder(t *testing.T) {
	fe := NewFlagEncoder()
	names := [maxFlags]string{"1", "2", "3", "4", "5", "6"}
	values := [maxFlags]bool{true, false, true, false, true, false}
	for i := 0; i < maxFlags; i++ {
		fe.Set(names[i], values[i])
	}
	c := fe.Encode()
	fe2 := Decode(c, names[0], names[1], names[2], names[3], names[4], names[5])
	for i := 0; i < maxFlags; i++ {
		v, found := fe2.Get(names[i])
		if !found {
			t.Errorf("Did not find '%v'", names[i])
		}
		assert(t, names[i], values[i], v)
	}
}
