package pkg

import (
	"github.com/zachmann/mytoken/internal/server/supertoken/token"
	"github.com/zachmann/mytoken/pkg/model"
)

// AccessTokenRequest holds an request for an access token
type AccessTokenRequest struct {
	Issuer     string          `json:"oidc_issuer"`
	GrantType  model.GrantType `json:"grant_type"`
	SuperToken token.Token     `json:"super_token"`
	Scope      string          `json:"scope"`
	Audience   string          `json:"audience"`
	Comment    string          `json:"comment"`
}
