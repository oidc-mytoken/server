package oidcReqRes

import "github.com/zachmann/mytoken/internal/utils"

// RefreshRequest is the oidc request for an refresh flow
type RefreshRequest struct {
	GrantType    string `json:"grant_type"`
	RefreshToken string `json:"refresh_token"`
	Scopes       string `json:"scope,omitempty"`
	Audiences    string `json:"audience,omitempty"`
}

// NewRefreshRequest creates a new RefreshRequest for a given refresh token
func NewRefreshRequest(rt string) *RefreshRequest {
	return &RefreshRequest{
		GrantType:    "refresh_token",
		RefreshToken: rt,
	}
}

// ToFormData formats the RefreshRequest as a string map
func (r *RefreshRequest) ToFormData() map[string]string {
	return utils.StructToStringMapUsingJSONTags(r)
}

// RevokeRequest is a oidc request for revoking tokens
type RevokeRequest struct {
	Token     string `json:"token"`
	TokenType string `json:"token_type_hint"`
}

// NewRTRevokeRequest creates a new RevokeRequest for revoking the passed refresh token
func NewRTRevokeRequest(rt string) *RevokeRequest {
	return &RevokeRequest{
		Token:     rt,
		TokenType: "refresh_token",
	}
}

// ToFormData formats the RevokeRequest as a string map
func (r *RevokeRequest) ToFormData() map[string]string {
	return utils.StructToStringMapUsingJSONTags(r)
}
