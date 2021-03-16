package mytoken

import (
	"testing"

	"github.com/oidc-mytoken/server/shared/utils/unixtime"
)

func TestMyToken_ExpiresIn_Unset(t *testing.T) {
	mt := Mytoken{}
	expIn := mt.ExpiresIn()
	if expIn != 0 {
		t.Error("Mytoken with empty expires_at should not expire")
	}
}

func TestMyToken_ExpiresIn_Future(t *testing.T) {
	in := int64(100)
	mt := Mytoken{ExpiresAt: unixtime.InSeconds(in)}
	expIn := mt.ExpiresIn()
	if expIn != uint64(in) {
		t.Errorf("Expected expires in to be %d not %d", in, expIn)
	}
}

func TestMyToken_ExpiresIn_Past(t *testing.T) {
	mt := Mytoken{ExpiresAt: 100}
	expIn := mt.ExpiresIn()
	if expIn != 0 {
		t.Errorf("Expected expires_in to be 0 when token already expired, not %d", expIn)
	}
}
