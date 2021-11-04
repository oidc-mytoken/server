package hashUtils

import (
	"testing"
)

func TestSHA512Str(t *testing.T) {
	data := `{"nbf":1599939600,"exp":1599948600,"ip":["192.168.0.31"],"usages_AT":11}`
	hash := SHA512Str([]byte(data))
	expected := "BS3WfHbHNUiVU8sJ+F49H9+69HnFtfVDy2m22vBv588nZ0kGblVNxZEcrTN+5NUiRkM7W80N4VpPgwEZBZl+3g=="
	if hash != expected {
		t.Errorf("hash '%s' does not match expected hash '%s'", hash, expected)
	}
}

func TestSHA3_256Str(t *testing.T) {
	data := `{"nbf":1599939600,"exp":1599948600,"ip":["192.168.0.31"],"usages_AT":11}`
	hash := SHA3_256Str([]byte(data))
	expected := "MyHsOR2PSCAd6jBBYWB5C34bc+MfPlmKuEDEScg6vj4="
	if hash != expected {
		t.Errorf("hash '%s' does not match expected hash '%s'", hash, expected)
	}
}

func TestSHA3_512Str(t *testing.T) {
	data := `{"nbf":1599939600,"exp":1599948600,"ip":["192.168.0.31"],"usages_AT":11}`
	hash := SHA3_512Str([]byte(data))
	expected := "FpVkiVUOi2/BiJE3zPvfz61YcRIdqoAKwVcH06PwK83SbopSXS9uuHJLVxmo4R8TxBX5dYJ4BxVba7WtwA+jzw=="
	if hash != expected {
		t.Errorf("hash '%s' does not match expected hash '%s'", hash, expected)
	}
}

func TestHMACBasedHash(t *testing.T) {
	data := `{"nbf":1599939600,"exp":1599948600,"ip":["192.168.0.31"],"usages_AT":11}`
	hash := HMACBasedHash([]byte(data))
	expected := "eyJuYmYiOjE1OTk5Mzk2MDAsImV4cCI6MTU5OTk0ODYwMCwiaXAiOlsiMTkyLjE2OC4wLjMxIl0sInVzYWdlc19BVCI6MTF9L3Lomh4pL9+cwcig05q2/MSuKFRsr6CMVw2iUcQjRAhdOt8qGVwg1R08eICUg8sKiwNDpXr92k1iWIMDcvuA/Q=="
	if hash != expected {
		t.Errorf("hash '%s' does not match expected hash '%s'", hash, expected)
	}
}
