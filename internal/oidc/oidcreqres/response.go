package oidcreqres

// OIDCErrorResponse is the error response of an oidc provider
type OIDCErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`

	Status int `json:"-"`
}

// OIDCTokenResponse is the token response of an oidc provider
type OIDCTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scopes       string `json:"scope"`
	IDToken      string `json:"id_token"`
}
