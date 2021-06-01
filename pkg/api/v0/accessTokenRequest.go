package api

// AccessTokenRequest holds an request for an access token
type AccessTokenRequest struct {
	Issuer    string `json:"oidc_issuer,omitempty" form:"issuer" xml:"oidc_issuer"`
	GrantType string `json:"grant_type" form:"grant_type" xml:"grant_type"`
	Mytoken   string `json:"mytoken" form:"mytoken" xml:"mytoken"`
	Scope     string `json:"scope,omitempty" form:"scope" xml:"scope"`
	Audience  string `json:"audience,omitempty" form:"audience" xml:"audience"`
	Comment   string `json:"comment,omitempty" form:"comment" xml:"comment"`
}
