package hashUtils

import (
	"testing"
)

func TestHashUtils_SHA512(t *testing.T) {
	data := `{"nbf":1599939600,"exp":1599948600,"ip":["192.168.0.31"],"usages_AT":11}`
	hash, err := SHA512Str([]byte(data))
	if err != nil {
		t.Error(err)
	}
	expected := "052dd67c76c735489553cb09f85e3d1fdfbaf479c5b5f543cb69b6daf06fe7cf276749066e554dc5911cad337ee4d52246433b5bcd0de15a4f83011905997ede"
	if hash != expected {
		t.Errorf("Hash '%s' does not match expected hash '%s'", hash, expected)
	}
}
