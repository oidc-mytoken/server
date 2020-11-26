package pkg

import (
	"github.com/zachmann/mytoken/internal/model"
)

type AccessTokenRequest struct {
	Issuer     string          `json:"oidc_issuer"`
	GrantType  model.GrantType `json:"grant_type"`
	SuperToken string          `json:"super_token"`
	Scope      string          `json:"scope"`
	Audience   string          `json:"audience"`
	Comment    string          `json:"comment"`
}