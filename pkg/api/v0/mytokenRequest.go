package api

// MytokenFromMytokenRequest is a request to create a new Mytoken from an existing Mytoken
type MytokenFromMytokenRequest struct {
	Issuer                       string       `json:"oidc_issuer"`
	GrantType                    string       `json:"grant_type"`
	Mytoken                      string       `json:"mytoken"`
	Restrictions                 Restrictions `json:"restrictions"`
	Capabilities                 []string     `json:"capabilities"`
	SubtokenCapabilities         []string     `json:"subtoken_capabilities"`
	Name                         string       `json:"name"`
	ResponseType                 string       `json:"response_type"`
	FailOnRestrictionsNotTighter bool         `json:"error_on_restrictions"`
}
