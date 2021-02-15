package stid

import (
	"testing"
)

func TestSTID_HashValid(t *testing.T) {
	id := STID{}
	if id.HashValid() {
		t.Errorf("Empty stid should not be valid")
	}

	id = New()
	if !id.HashValid() {
		t.Errorf("Created stid should be valid")
	}
}
