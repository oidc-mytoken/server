package api

// AuthCodeFlowRequest holds a authorization code flow request
type AuthCodeFlowRequest struct {
	OIDCFlowRequest
	RedirectType string `json:"redirect_type"`
}

// OIDCFlowRequest holds the request for an OIDC Flow request
type OIDCFlowRequest struct {
	Issuer               string       `json:"oidc_issuer"`
	GrantType            string       `json:"grant_type"`
	OIDCFlow             string       `json:"oidc_flow"`
	Restrictions         Restrictions `json:"restrictions"`
	Capabilities         Capabilities `json:"capabilities"`
	SubtokenCapabilities Capabilities `json:"subtoken_capabilities"`
	Name                 string       `json:"name"`
	ResponseType         string       `json:"response_type"`
}
