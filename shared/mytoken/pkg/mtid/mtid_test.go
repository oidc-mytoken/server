package mtid

import (
	"testing"
)

func TestSTID_HashValid(t *testing.T) {
	id := MTID{}
	if id.HashValid() {
		t.Errorf("Empty stid should not be valid")
	}

	id = New()
	if !id.HashValid() {
		t.Errorf("Created stid should be valid")
	}
}
