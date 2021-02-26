package supertoken

import (
	"testing"

	"github.com/oidc-mytoken/server/shared/utils/unixtime"
)

func TestSuperToken_ExpiresIn_Unset(t *testing.T) {
	st := SuperToken{}
	expIn := st.ExpiresIn()
	if expIn != 0 {
		t.Error("Supertoken with empty expires_at should not expire")
	}
}

func TestSuperToken_ExpiresIn_Future(t *testing.T) {
	in := int64(100)
	st := SuperToken{ExpiresAt: unixtime.InSeconds(in)}
	expIn := st.ExpiresIn()
	if expIn != uint64(in) {
		t.Errorf("Expected expires in to be %d not %d", in, expIn)
	}
}

func TestSuperToken_ExpiresIn_Past(t *testing.T) {
	st := SuperToken{ExpiresAt: 100}
	expIn := st.ExpiresIn()
	if expIn != 0 {
		t.Errorf("Expected expires_in to be 0 when token already expired, not %d", expIn)
	}
}
