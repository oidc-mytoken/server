package oidcreqres

import (
	"strconv"
	"time"

	"github.com/oidc-mytoken/utils/utils/jwtutils"
	log "github.com/sirupsen/logrus"
)

// OIDCErrorResponse is the error response of an oidc provider
type OIDCErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`

	Status int `json:"-"`
}

// OIDCTokenResponse is the token response of an oidc provider
type OIDCTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scopes       string `json:"scope"`
	IDToken      string `json:"id_token"`
}

// AccessTokenExpiresAt returns the time when the access token expires. The value is derived from the expires_in field
// of the response; if that is not given, the exp claim of the access token is used. The zero value is returned if the
// lifetime cannot be determined.
func (o *OIDCTokenResponse) AccessTokenExpiresAt(rlog log.Ext1FieldLogger) time.Time {
	if o.ExpiresIn > 0 {
		return time.Now().Add(time.Duration(o.ExpiresIn) * time.Second)
	}
	if v := jwtutils.GetValueFromJWT(rlog, o.AccessToken, "exp"); v != nil {
		switch exp := v.(type) {
		case float64:
			return time.Unix(int64(exp), 0)
		case int64:
			return time.Unix(exp, 0)
		case string:
			if secs, err := strconv.ParseInt(exp, 10, 64); err == nil {
				return time.Unix(secs, 0)
			}
		}
	}
	return time.Time{}
}
