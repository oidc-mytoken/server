package mtid

import (
	"testing"
)

func TestMTID_HashValid(t *testing.T) {
	id := MTID{}
	if id.HashValid() {
		t.Errorf("Empty mtid should not be valid")
	}

	id = New()
	if !id.HashValid() {
		t.Errorf("Created mtid should be valid")
	}
}
