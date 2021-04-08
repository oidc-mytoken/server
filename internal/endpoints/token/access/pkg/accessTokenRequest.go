package pkg

import (
	"github.com/oidc-mytoken/server/pkg/api/v0"
	"github.com/oidc-mytoken/server/shared/model"
	"github.com/oidc-mytoken/server/shared/mytoken/token"
)

// AccessTokenRequest holds an request for an access token
type AccessTokenRequest struct {
	api.AccessTokenRequest `json:",inline"`
	GrantType              model.GrantType `json:"grant_type"`
	Mytoken                token.Token     `json:"mytoken"`
}
