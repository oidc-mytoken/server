package supertoken

import (
	"testing"
	"time"
)

func TestSuperToken_ExpiresIn_Unset(t *testing.T) {
	st := SuperToken{}
	expIn := st.expiresIn()
	if expIn != 0 {
		t.Error("Supertoken with empty expires_at should not expire")
	}
}

func TestSuperToken_ExpiresIn_Future(t *testing.T) {
	in := uint64(100)
	st := SuperToken{ExpiresAt: time.Now().Unix() + int64(in)}
	expIn := st.expiresIn()
	if expIn != in {
		t.Errorf("Expected expires in to be %d not %d", in, expIn)
	}
}

func TestSuperToken_ExpiresIn_Past(t *testing.T) {
	st := SuperToken{ExpiresAt: 100}
	expIn := st.expiresIn()
	if expIn != 0 {
		t.Errorf("Expected expires_in to be 0 when token already expired, not %d", expIn)
	}
}
