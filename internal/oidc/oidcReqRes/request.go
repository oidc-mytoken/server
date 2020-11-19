package oidcReqRes

import "github.com/zachmann/mytoken/internal/utils"

type RefreshRequest struct {
	GrantType    string `json:"grant_type"`
	RefreshToken string `json:"refresh_token"`
	Scopes       string `json:"scope,omitempty"`
	Audiences    string `json:"audience,omitempty"`
}

func NewRefreshRequest(rt string) *RefreshRequest {
	return &RefreshRequest{
		GrantType:    "refresh_token",
		RefreshToken: rt,
	}
}

func (r *RefreshRequest) ToFormData() map[string]string {
	return utils.StructToStringMapUsingJSONTags(r)
}
