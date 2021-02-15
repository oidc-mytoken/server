package supertoken

import (
	"testing"
	"time"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/jws"
	"github.com/oidc-mytoken/server/shared/supertoken/capabilities"
)

func TestSuperTokenJWTUnique(t *testing.T) {
	config.Load()
	jws.LoadKey()
	st := NewSuperToken("sub", "iss", nil, capabilities.AllCapabilities, nil)
	jwt1, err := st.ToJWT()
	if err != nil {
		t.Error(err)
	}
	jwt2, err := st.ToJWT()
	if err != nil {
		t.Error(err)
	}
	if jwt1 == jwt2 {
		t.Errorf("JWTs not unique: '%s'", jwt1)
	}
}

func TestSuperToken_ExpiresIn_Unset(t *testing.T) {
	st := SuperToken{}
	expIn := st.ExpiresIn()
	if expIn != 0 {
		t.Error("Supertoken with empty expires_at should not expire")
	}
}

func TestSuperToken_ExpiresIn_Future(t *testing.T) {
	in := uint64(100)
	st := SuperToken{ExpiresAt: time.Now().Unix() + int64(in)}
	expIn := st.ExpiresIn()
	if expIn != in {
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
