package supertoken

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	uuid "github.com/satori/go.uuid"
	"github.com/zachmann/mytoken/internal/config"
	response "github.com/zachmann/mytoken/internal/endpoints/token/super/pkg"
	"github.com/zachmann/mytoken/internal/jws"
	"github.com/zachmann/mytoken/internal/supertoken/capabilities"
	"github.com/zachmann/mytoken/internal/supertoken/restrictions"
	"github.com/zachmann/mytoken/internal/utils/issuerUtils"
)

// SuperToken is a mytoken SuperToken
type SuperToken struct {
	Issuer       string                    `json:"iss"`
	Subject      string                    `json:"sub"`
	ExpiresAt    int64                     `json:"exp,omitempty"`
	NotBefore    int64                     `json:"nbf"`
	IssuedAt     int64                     `json:"iat"`
	ID           uuid.UUID                 `json:"jti"`
	Audience     string                    `json:"aud"`
	OIDCSubject  string                    `json:"oidc_sub"`
	OIDCIssuer   string                    `json:"oidc_iss"`
	Restrictions restrictions.Restrictions `json:"restrictions,omitempty"`
	Capabilities capabilities.Capabilities `json:"capabilities"`
	jwt          string
}

// NewSuperToken creates a new SuperToken
func NewSuperToken(oidcSub, oidcIss string, r restrictions.Restrictions, c capabilities.Capabilities) *SuperToken {
	now := time.Now().Unix()
	st := &SuperToken{
		ID:           uuid.NewV4(),
		IssuedAt:     now,
		NotBefore:    now,
		Issuer:       config.Get().IssuerURL,
		Subject:      issuerUtils.CombineSubIss(oidcSub, oidcIss),
		Audience:     config.Get().IssuerURL,
		OIDCIssuer:   oidcIss,
		OIDCSubject:  oidcSub,
		Capabilities: c,
	}
	if len(r) > 0 {
		st.Restrictions = r
		exp := r.GetExpires()
		if exp != 0 {
			st.ExpiresAt = exp
		}
		nbf := r.GetNotBefore()
		if nbf != 0 && nbf > now {
			st.NotBefore = nbf
		}
	}
	return st
}

func (st *SuperToken) ExpiresIn() uint64 {
	now := time.Now().Unix()
	expAt := st.ExpiresAt
	if expAt > 0 && expAt < now {
		return uint64(expAt - now)
	}
	return 0
}

func (st *SuperToken) Valid() error {
	standardClaims := jwt.StandardClaims{
		Audience:  st.Audience,
		ExpiresAt: st.ExpiresAt,
		Id:        st.ID.String(),
		IssuedAt:  st.IssuedAt,
		Issuer:    st.Issuer,
		NotBefore: st.NotBefore,
		Subject:   st.Subject,
	}
	if err := standardClaims.Valid(); err != nil {
		return err
	}
	if ok := standardClaims.VerifyIssuer(config.Get().IssuerURL, true); !ok {
		return fmt.Errorf("Invalid issuer")
	}
	if ok := standardClaims.VerifyAudience(config.Get().IssuerURL, true); !ok {
		return fmt.Errorf("Invalid Audience")
	}

	//TODO
	return nil
}

// ToSuperTokenResponse returns a SuperTokenResponse for this token. It requires that jwt is set, i.e. ToJWT must have been called earlier on this token. This is always the case, if the token has been stored.
func (st *SuperToken) ToSuperTokenResponse() response.SuperTokenResponse {
	return response.SuperTokenResponse{
		SuperToken:   st.jwt,
		ExpiresIn:    st.ExpiresIn(),
		Restrictions: st.Restrictions,
		Capabilities: st.Capabilities,
	}
}

// ToJWT returns the SuperToken as JWT
func (st *SuperToken) ToJWT() (string, error) {
	if st.jwt != "" {
		return st.jwt, nil
	}
	var err error
	st.jwt, err = jwt.NewWithClaims(jwt.GetSigningMethod(config.Get().Signing.Alg), st).SignedString(jws.GetPrivateKey())
	return st.jwt, err
}

// Value implements the driver.Valuer interface.
func (st *SuperToken) Value() (driver.Value, error) {
	return st.ToJWT()
}

// Scan implements the sql.Scanner interface.
func (st *SuperToken) Scan(src interface{}) error {
	tmp, err := ParseJWT(src.(string))
	if err != nil {
		return err
	}
	*st = *tmp
	return nil
}

func ParseJWT(token string) (*SuperToken, error) {
	tok, err := jwt.ParseWithClaims(token, &SuperToken{}, func(t *jwt.Token) (interface{}, error) {
		return jws.GetPublicKey(), nil
	})
	if err != nil {
		return nil, err
	}

	if st, ok := tok.Claims.(*SuperToken); ok && tok.Valid {
		return st, nil
	}
	return nil, fmt.Errorf("Propably token not valid")
}
