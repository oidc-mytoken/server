package hashUtils

import (
	"testing"
)

func TestHashUtils_SHA512(t *testing.T) {
	data := `{"nbf":1599939600,"exp":1599948600,"ip":["192.168.0.31"],"usages_AT":11}`
	hash := SHA512Str([]byte(data))
	expected := "BS3WfHbHNUiVU8sJ+F49H9+69HnFtfVDy2m22vBv588nZ0kGblVNxZEcrTN+5NUiRkM7W80N4VpPgwEZBZl+3g=="
	if hash != expected {
		t.Errorf("hash '%s' does not match expected hash '%s'", hash, expected)
	}
}
