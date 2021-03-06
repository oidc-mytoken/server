package pkg

import (
	"github.com/oidc-mytoken/server/pkg/model"
	"github.com/oidc-mytoken/server/shared/supertoken/token"
)

// AccessTokenRequest holds an request for an access token
type AccessTokenRequest struct {
	Issuer     string          `json:"oidc_issuer,omitempty"`
	GrantType  model.GrantType `json:"grant_type"`
	SuperToken token.Token     `json:"super_token"`
	Scope      string          `json:"scope,omitempty"`
	Audience   string          `json:"audience,omitempty"`
	Comment    string          `json:"comment,omitempty"`
}
