package pkg

import (
	"github.com/oidc-mytoken/api/v0"
	"github.com/oidc-mytoken/server/shared/model"
	"github.com/oidc-mytoken/server/shared/mytoken/token"
)

// AccessTokenRequest holds an request for an access token
type AccessTokenRequest struct {
	api.AccessTokenRequest `json:",inline"`
	GrantType              model.GrantType `json:"grant_type" xml:"grant_type" form:"grant_type"`
	Mytoken                token.Token     `json:"mytoken" xml:"mytoken" form:"mytoken"`
	RefreshToken           token.Token     `json:"refresh_token" xml:"refresh_token" form:"refresh_token"`
}
