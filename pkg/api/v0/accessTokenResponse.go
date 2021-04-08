package api

// AccessTokenResponse is the response to a access token request
type AccessTokenResponse struct {
	AccessToken string   `json:"access_token"`
	TokenType   string   `json:"token_type"`
	ExpiresIn   int64    `json:"expires_in"`
	Scope       string   `json:"scope,omitempty"`
	Audiences   []string `json:"audience,omitempty"`
}
