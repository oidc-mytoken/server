package api

// AccessTokenRequest holds an request for an access token
type AccessTokenRequest struct {
	Issuer    string `json:"oidc_issuer,omitempty"`
	GrantType string `json:"grant_type"`
	Mytoken   string `json:"mytoken"`
	Scope     string `json:"scope,omitempty"`
	Audience  string `json:"audience,omitempty"`
	Comment   string `json:"comment,omitempty"`
}
