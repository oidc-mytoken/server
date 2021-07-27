package pkg

import (
	"github.com/oidc-mytoken/api/v0"
	"github.com/oidc-mytoken/server/shared/model"
	"github.com/oidc-mytoken/server/shared/mytoken/universalmytoken"
)

// AccessTokenRequest holds an request for an access token
type AccessTokenRequest struct {
	api.AccessTokenRequest `json:",inline"`
	GrantType              model.GrantType                   `json:"grant_type" xml:"grant_type" form:"grant_type"`
	Mytoken                universalmytoken.UniversalMytoken `json:"mytoken" xml:"mytoken" form:"mytoken"`
	RefreshToken           universalmytoken.UniversalMytoken `json:"refresh_token" xml:"refresh_token" form:"refresh_token"`
}

// NewAccessTokenRequest returns a new AccessTokenRequest
func NewAccessTokenRequest() AccessTokenRequest {
	return AccessTokenRequest{
		GrantType: -1, // This value will remain if grant_type is not contained in the request. We have to set it to -1,
		// because the default of 0 would be a valid GrantType
	}
}
